package main

import (
	_ "github.com/chuckwe/recircle/demo"
	"github.com/chuckwe/recircle/spider"
)

func main() {
	spider.Sc.Init()
	// client := req.C().
	// 	SetUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36").
	// 	SetTLSFingerprintChrome() // 模拟 Chrome 浏览器的 TLS 握手指纹，让网站相信这是 Chrome 浏览器在访问，予以通行。
	//
	// client = client.SetProxyURL("http://127.0.0.1:7899")
	// client = client.ImpersonateChrome()
	// r, err := client.R().Get("https://twkan.com/ajax_novels/chapterlist/15606.html")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("%+v", r)
}
