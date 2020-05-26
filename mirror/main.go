package main

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/otium/queue"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"
)

type Source struct {
	Name   string   `yaml:name`
	Url    string   `yaml:url`
	Charts []string `yaml:charts`
}

type Config struct {
	Source  []Source `yaml:source`
	RootURL string   `yaml:rootURL`
}

func (c *Config) GetPublicURL(p string) string {
	u, _ := url.Parse(c.RootURL)
	u.Path = path.Join(u.Path, p)
	return u.String()
}

func (s *Source) GetIconPath(u string) string {
	u2, _ := url.Parse(u)
	return path.Join("icons", s.Name+"-"+path.Base(u2.Path))
}

func (s *Source) GetURLPath(u string) string {
	u2, _ := url.Parse(u)
	return path.Join("charts", s.Name+"-"+path.Base(u2.Path))
}

func (s *Source) GetFullURL(u string) string {
	if strings.HasPrefix(u, "http") {
		return u
	}
	u2, _ := url.Parse(s.Url)
	u2.Path = path.Join(u2.Path, u)
	return u2.String()
}

func main() {

	log.Println("Sync charts start!!!")

	b, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(b, config); err != nil {
		panic(err)
	}

	newIndex := &repo.IndexFile{
		Entries: map[string]repo.ChartVersions{},
	}
	for _, s := range config.Source {
		if index, err := handleSourceDownload(config, s); err != nil {
			log.Println(index)
			panic(err)
		} else {
			newIndex.Merge(index)
		}
	}

	// save index.yaml file
	newIndex.SortEntries()
	indexFileData, err := yaml.Marshal(newIndex)
	if err != nil {
		panic(err)
	}
	file, err := os.OpenFile("index.yaml", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	file.Write(indexFileData)

	downloadQueue.Wait()

	log.Println("Sync charts finish!!!")
}

func handleSourceDownload(config *Config, s Source) (*repo.IndexFile, error) {
	url, err := url.Parse(s.Url)
	if err != nil {
		return nil, err
	}
	url.Path = path.Join(url.Path, "index.yaml")
	body, err := Get(url.String())
	index := &repo.IndexFile{}
	yaml.Unmarshal(body, index)
	index.SortEntries()

	newIndex := &repo.IndexFile{
		Entries: map[string]repo.ChartVersions{},
	}
	for k, v := range index.Entries {
		if Index(s.Charts, k) {
			for _, c := range v {
				if len(c.URLs) == 0 {
					return nil, errors.New("chart metadata.urls is empty!")
				}
				for _, u := range c.URLs {
					AddDownloadQueue(s.GetFullURL(u), s.GetURLPath(u))
				}

				if len(c.Icon) > 0 {
					AddDownloadQueue(s.GetFullURL(c.Icon), s.GetIconPath(c.Icon))
				}

				// Parse new Entries
				urls := []string{}
				for _, u := range c.URLs {
					urls = append(urls, config.GetPublicURL(s.GetURLPath(u)))
				}
				chart := CopyChartVersion(c)
				chart.URLs = urls
				chart.Icon = config.GetPublicURL(s.GetIconPath(c.Icon))
				newIndex.Merge(&repo.IndexFile{
					Entries: map[string]repo.ChartVersions{
						c.Name: repo.ChartVersions{chart},
					},
				})
			}
		}
	}

	return newIndex, nil
}

func Index(arr []string, s string) bool {
	for _, i := range arr {
		if i == s {
			return true
		}
	}
	return false
}
func CopyChartVersion(c *repo.ChartVersion) *repo.ChartVersion {
	chart := &repo.ChartVersion{}
	b, _ := yaml.Marshal(c)
	yaml.Unmarshal(b, chart)
	return chart
}

func Get(u string) ([]byte, error) {
	log.Println("Start request URL:", u)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

var (
	downloadURLs  = []string{}
	downloadQueue *queue.Queue
)

func AddDownloadQueue(u, filePath string) {

	if Index(downloadURLs, u) {
		return
	}
	downloadURLs = append(downloadURLs, u)

	type downloadTask struct {
		url      string
		filePath string
	}

	if downloadQueue == nil {
		downloadQueue = queue.NewQueue(func(val interface{}) {

			task := val.(downloadTask)

			log.Println("Start download URL:", task.url)
			resp, err := http.Get(task.url)
			if err != nil {
				panic(err)
			}

			dir := path.Dir(task.filePath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				panic(err)
			}

			file, err := os.OpenFile(task.filePath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
			if err != nil {
				panic(err)
			}

			if _, err := io.Copy(file, resp.Body); err != nil {
				panic(err)
			}

		}, 20)
	}

	downloadQueue.Push(downloadTask{
		url:      u,
		filePath: filePath,
	})
}
