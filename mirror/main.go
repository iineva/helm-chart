package main

import (
	"errors"
	"github.com/iineva/helm-chart/mirror/pkg/common"
	"github.com/iineva/helm-chart/mirror/pkg/config"
	"github.com/iineva/helm-chart/mirror/pkg/downloader"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/Masterminds/semver/v3"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"
)

func main() {

	log.Println("Sync charts start!!!")

	b, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}

	conf := &config.Config{}
	if err := yaml.Unmarshal(b, conf); err != nil {
		panic(err)
	}

	download := downloader.New(int(conf.DownloadConcurrent))

	newIndex := &repo.IndexFile{
		Entries: map[string]repo.ChartVersions{},
	}
	for _, s := range conf.Source {
		if index, err := handleSourceDownload(conf, s, download); err != nil {
			log.Println(index)
			panic(err)
		} else {
			newIndex.Merge(index)
		}
	}

	// save index.yaml file
	newIndex.APIVersion = "v1"
	newIndex.Generated = time.Now()
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

	if conf.DownloadFiles {
		download.Wait()
		a, b := download.Len()
		log.Printf("download len: %v / %v", a, b)
	}

	log.Println("Sync charts finish!!!")
}

func handleSourceDownload(conf *config.Config, s config.Source, download *downloader.Downloader) (*repo.IndexFile, error) {
	url, err := url.Parse(s.Url)
	if err != nil {
		return nil, err
	}
	url.Path = path.Join(url.Path, "index.yaml")
	body, err := common.HTTPGet(url.String())
	index := &repo.IndexFile{}
	yaml.Unmarshal(body, index)
	index.SortEntries()

	newIndex := &repo.IndexFile{
		Entries: map[string]repo.ChartVersions{},
	}
	for k, v := range index.Entries {
		if len(s.Charts) == 0 || common.Index(s.Charts, k) {

			if len(v) == 0 {
				continue
			}
			maxVersion, err := semver.NewVersion(v[0].Version)
			if err != nil {
				panic(err)
			}

			for _, c := range v {

				curVersion, err := semver.NewVersion(c.Version)
				if err != nil {
					panic(err)
				}
				if conf.MajorVersions > 0 && maxVersion.Major()-curVersion.Major() >= conf.MajorVersions {
					continue
				}

				if len(c.URLs) == 0 {
					return nil, errors.New("chart metadata.urls is empty!")
				}
				for _, u := range c.URLs {
					if conf.DownloadFiles {
						download.Push(s.GetFullURL(u), s.GetURLPath(u))
					}
				}

				if len(c.Icon) > 0 {
					if conf.DownloadFiles {
						download.Push(s.GetFullURL(c.Icon), s.GetIconPath(c.Icon))
					}
				}

				chart := CopyChartVersion(c)
				// Parse new Entries
				urls := []string{}
				for _, u := range c.URLs {
					if conf.DownloadFiles {
						urls = append(urls, conf.GetPublicURL(s.GetURLPath(u)))
					} else {
						urls = append(urls, s.GetFullURL(u))
					}
				}
				chart.URLs = urls
				if conf.DownloadFiles {
					chart.Icon = conf.GetPublicURL(s.GetIconPath(c.Icon))
				} else {
					chart.Icon = s.GetFullURL(c.Icon)
				}

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

func CopyChartVersion(c *repo.ChartVersion) *repo.ChartVersion {
	chart := &repo.ChartVersion{}
	b, _ := yaml.Marshal(c)
	yaml.Unmarshal(b, chart)
	return chart
}
