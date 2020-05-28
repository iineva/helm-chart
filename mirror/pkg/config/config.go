package config

import (
	"net/url"
	"path"
	"strings"
)

type Source struct {
	Name   string   `yaml:name`
	Url    string   `yaml:url`
	Charts []string `yaml:charts`
}

type Config struct {
	Source             []Source `yaml:source`
	RootURL            string   `yaml:rootURL`
	MajorVersions      uint64   `yaml:majorVersions`
	DownloadFiles      bool     `yaml:downloadFiles`
	DownloadConcurrent uint64   `yaml:downloadConcurrent`
}

func (c *Config) GetPublicURL(p string) string {
	u, _ := url.Parse(c.RootURL)
	u.Path = path.Join(u.Path, p)
	return u.String()
}

func (s *Source) GetIconPath(u string) string {
	u2, _ := url.Parse(u)
	return path.Join("icons", s.Name, path.Base(u2.Path))
}

func (s *Source) GetURLPath(u string) string {
	u2, _ := url.Parse(u)
	return path.Join("charts", s.Name, path.Base(u2.Path))
}

func (s *Source) GetFullURL(u string) string {
	if strings.HasPrefix(u, "http") {
		return u
	}
	u2, _ := url.Parse(s.Url)
	u2.Path = path.Join(u2.Path, u)
	return u2.String()
}
