package downloader

import (
	"fmt"
	"testing"
)

func newDownloader() Downloader {
	return New()
}

func TestDownload(t *testing.T) {
	d := newDownloader()
	r := &Request{
		Url: "https://www.69shu.com/38938/",
	}
	r.SetMethod("GET").SetRetry(true, 2).SetProxy("http://127.0.0.1:1087")
	resp, err := d.Download(r)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.Body)
}
