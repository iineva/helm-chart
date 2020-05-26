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

	"sigs.k8s.io/yaml"
	"helm.sh/helm/v3/pkg/repo"
)

type Source struct {
	Name   string   `yaml:name`
	Url    string   `yaml:url`
	Charts []string `yaml:charts`
}

type Config struct {
	Source   []Source `yaml:source`
	RootURL  string   `yaml:rootURL`
	rootPath string
	rootHost string
}

func (c *Config) RootPath() string {
	if len(c.rootPath) == 0 {
		urls, _ := url.Parse(c.RootURL)
		if len(urls.Path) == 0 {
			c.rootPath = "/"
		}
		c.rootPath = urls.Path
	}
	return c.rootPath
}

func (c *Config) PublicURL(p string) string {
	if len(c.rootHost) == 0 {
		urls, _ := url.Parse(c.RootURL)
		urls.Path = ""
		c.rootHost = urls.String()
	}
	return c.rootHost + p
}

func (c *Config) GetIconURL(u, chartName string) string {
	u2, _ := url.Parse(u)
	return c.PublicURL(path.Join(c.RootPath(), "icons", chartName+"-"+path.Base(u2.Path)))
}

func (c *Config) GetURLs(us []string, chartName string) []string {
	urls := []string{}
	for _, u := range us {
		u2, _ := url.Parse(u)
		urls = append(urls, c.PublicURL(path.Join(c.RootPath(), "icons", chartName+"-"+path.Base(u2.Path))))
	}
	return urls
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
					if err := Download(u, "charts", s); err != nil {
						return nil, err
					}
				}

				if len(c.Icon) > 0 {
					if err := Download(c.Icon, "icons", s); err != nil {
						return nil, err
					}
				}

				// Parse new Entries
				c.URLs = config.GetURLs(c.URLs, c.Name)
				c.Icon = config.GetIconURL(c.Icon, c.Name)
				newIndex.Merge(&repo.IndexFile{
					Entries: map[string]repo.ChartVersions{
						c.Name: repo.ChartVersions{c},
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
	downloadURLs = []string{}
)

func Download(u, dir string, s Source) error {

	if Index(downloadURLs, u) {
		return nil
	}

	log.Println("Start download URL:", u)
	downloadURLs = append(downloadURLs, u)
	resp, err := http.Get(u)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	fileName := path.Join(dir, s.Name+"-"+path.Base(u))
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)

	if err != nil {
		return err
	}

	if _, err := io.Copy(file, resp.Body); err != nil {
		return err
	}

	return nil
}
