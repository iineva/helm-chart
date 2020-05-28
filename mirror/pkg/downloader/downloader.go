package downloader

import (
	"github.com/iineva/helm-chart/mirror/pkg/common"
	"github.com/otium/queue"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

type Downloader struct {
	downloadURLs  []string
	downloadQueue *queue.Queue
	concurrent    int
}

type downloadTask struct {
	url      string
	filePath string
}

func New(concurrent int) *Downloader {
	if concurrent <= 0 {
		concurrent = 1
	}
	d := &Downloader{
		downloadURLs: []string{},
		concurrent:   concurrent,
	}
	d.downloadQueue = queue.NewQueue(d.runTask, concurrent)
	return d
}

func (d *Downloader) runTask(val interface{}) {

	task := val.(downloadTask)

	log.Println("Start download URL:", task.url)
	resp, err := http.Get(task.url)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Request fail:", resp)
		return
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

}

func (d *Downloader) Push(u, filePath string) {

	if common.Index(d.downloadURLs, u) {
		return
	}
	d.downloadURLs = append(d.downloadURLs, u)

	d.downloadQueue.Push(downloadTask{
		url:      u,
		filePath: filePath,
	})
}

func (d *Downloader) Wait() {
	d.downloadQueue.Wait()
}

func (d *Downloader) Len() (int, int) {
	return d.downloadQueue.Len()
}
