package demo

import (
	"bufio"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chuckwe/recircle/spider"
	"github.com/liuzl/gocc"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

var (
	urlLink        = "http://www.xsbiquge.la/book/36587/"
	txtFile        *os.File
	length         int32
	txtContent     = make([]string, 0)
	maxInContent   = int32(10)
	domain         = "https://twkan.com"
	searchBookName = "从恋综开始翻盘"
)

func init() {
	searchUrl := domain + "/modules/article/search.php"

	demoSpider := spider.NewSpider(spider.NewDownloader("imroc"))
	demoSpider.UniqueKey = "twkan"
	demoSpider.SetConcurrent(1)
	demoSpider.SetRangeTime(100)
	// demoSpider.SetRules("创建文件", CreateFile)
	demoSpider.SetRules("查询", SearchBook)
	demoSpider.SetRules("列表", ListQuery)
	demoSpider.SetRules("内容", DetailQuery, DownloadTxt)
	// demoSpider.SetGlobalPreRun(func(ctx *spider.Context) error {
	// 	r := spider.NewResource(demoSpider.UniqueKey, "查询", searchUrl, nil)
	// 	r.SetProxy("http://127.0.0.1:7899")
	// 	r.SetMethod("post")
	// 	r.Cloudflare = true
	// 	r.Request.JsEnable = true
	// 	r.Request.JsFunc = func(request *downloader.Request) (*http.Response, error) {
	// 		urlLancher := launcher.New().Headless(true).Set(flags.ProxyServer, r.GetProxy()).MustLaunch()
	// 		browser := rod.New().ControlURL(urlLancher).MustConnect()
	// 		defer browser.MustClose()
	// 		page := browser.MustPage(searchUrl)
	// 		page.MustElement("#searchkey").
	// 			MustInput(searchBookName).MustType(input.Enter)
	// 		_ = page.WaitLoad()
	// 		resp := &http.Response{
	// 			Status:     "200 OK",
	// 			StatusCode: 200,
	// 			Proto:      "HTTP/1.0",
	// 			ProtoMajor: 1,
	// 			ProtoMinor: 0,
	// 			Header:     make(http.Header),
	// 			Body:       io.NopCloser(bytes.NewBufferString(page.MustHTML())),
	// 		}
	// 		return resp, nil
	// 	}
	spider.Sc.Register(demoSpider)

	r := spider.NewResource(demoSpider.UniqueKey, "查询", searchUrl, nil)
	r.SetProxy("http://127.0.0.1:7899")
	r.SetMethod("post")
	r.DialTimeout = time.Second * 60
	r.Cloudflare = true
	r.PostData = fmt.Sprintf("searchkey=%s&searchtype=all", searchBookName)
	// https://twkan.com/ajax_novels/chapterlist/70340.html
	err := spider.Sc.AddResource(r)
	if err != nil {
		fmt.Println(err)
		return
	}
	// demoSpider.SetGlobalPreRun(func(ctx *spider.Context) error {
	// 	r := spider.NewResource(demoSpider.UniqueKey, "创建文件", urlLink, nil)
	// 	// r.SetProxy("socket://10.20.13.22:7898")
	//
	// 	r.SetRetry(true, 2)
	// 	err := spider.Sc.AddResource(r)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return nil
	// })
	// demoSpider.CloseCallback = func(s *spider.Spider) {
	// 	write := bufio.NewWriter(txtFile)
	// 	_, _ = write.WriteString(strings.Join(txtContent, "\n"))
	// 	// Flush将缓存的文件真正写入到文件中
	// 	_ = write.Flush()
	// 	_ = txtFile.Close()
	// 	return
	// }
	// r := spider.NewResource("xsbiquge", "列表", urlLink, nil)
	// // r.SetProxy("http://127.0.0.1:1087")
	// r.SetRetry(true, 2)
	//
	// err := spider.Sc.AddResource(r)
	// if err != nil {
	// 	panic(err)
	// }
}

func SearchBook(ctx *spider.Context) {
	query, _ := ctx.GetDom()
	fileName := ""
	listUrl := "https://twkan.com/ajax_novels/chapterlist/%s"
	authorName := ""
	liQuery := query.Find("#article_list_content li")
	bookID := ""
	if liQuery.Length() > 0 {
		bookUrl, _ := liQuery.First().Find("a").Attr("href")
		fileName = liQuery.First().Find("h3").Text()
		authorName = liQuery.First().Find(".labelbox").Children().First().Text()
		bookUrlArr := strings.Split(bookUrl, "/")
		bookID = bookUrlArr[len(bookUrlArr)-1]
		listUrl = fmt.Sprintf(listUrl, bookID)
	} else {
		bookUrl, _ := query.Find("div.booknav2 > h1 > a").Attr("href")
		fileName = query.Find(".booknav2>h1").Text()
		authorName = query.Find(".booknav2 > p:nth-child(2)").Text()
		bookUrlArr := strings.Split(bookUrl, "/")
		bookID = bookUrlArr[len(bookUrlArr)-1]
		listUrl = fmt.Sprintf(listUrl, bookID)
	}
	if len(fileName) == 0 {
		return
	}
	txtFile, _ = os.OpenFile(fileName+".txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	r := spider.NewResource("twkan", "列表", listUrl, map[string]string{
		"authorName": authorName,
		"fileName":   fileName,
		"bookID":     bookID,
	})
	r.SetProxy("http://127.0.0.1:7899")
	r.SetMethod("get")
	r.SetRetry(true, 2)
	r.Cloudflare = true
	err := spider.Sc.AddResource(r)
	if err != nil {
		fmt.Println(err)
		ctx.Abort()
		return
	}
}

func CreateFile(ctx *spider.Context) {
	query, _ := ctx.GetDom()
	fileName := query.Find("#info > h1").Text()
	txtFile, _ = os.OpenFile(fileName+".txt", os.O_WRONLY|os.O_CREATE, 0666)
}

func ListQuery(ctx *spider.Context) {
	query, _ := ctx.GetDom()
	query.Find("li").Each(func(i int, selection *goquery.Selection) {
		urlVal, _ := selection.Find("a").Attr("href")
		r := spider.NewResource("twkan", "内容", urlVal, map[string]string{
			"authorName": ctx.Info["authorName"],
			"fileName":   ctx.Info["fileName"],
			"bookID":     ctx.Info["bookID"],
		})
		r.SetProxy("http://127.0.0.1:7899")
		r.SetMethod("get")
		r.SetRetry(true, 2)
		r.Cloudflare = true
		err := spider.Sc.AddResource(r)
		if err != nil {
			fmt.Println(err)
			ctx.Abort()
			return
		}
	})
}

func DetailQuery(ctx *spider.Context) {
	query, _ := ctx.GetDom()
	contentQuery := query.Find(".txtnav")
	articleTitle := contentQuery.Find("h1").Text()
	articleTitle = strings.Trim(articleTitle, " ")
	createTime := contentQuery.Find(".txtinfo span").First().Text()
	createTime = strings.Trim(createTime, " ")
	authorName := ctx.Info["authorName"]
	contentQuery.Find("h1").Remove()
	contentQuery.Find(".txtinfo").Remove()
	contentQuery.Find(".txtright").Remove()
	contentQuery.Find(".txtcenter").Remove()
	content := contentQuery.Text()

	// contentQuery.RemoveMatcher(goquery.Single("h1"))
	// contentQuery.RemoveMatcher(goquery.Single(".txtinfo"))
	// contentQuery.RemoveFiltered(".txtinfo")
	// contentQuery.RemoveFiltered(".txtright")
	// contentQuery.RemoveFiltered("script")
	// contentQuery.RemoveFiltered("span")
	// content := contentQuery.Text()
	content = strings.ReplaceAll(content, " ", "")
	t2s, err := gocc.New("t2s")
	if err != nil {
		fmt.Println(err)
		return
	}
	articleTitle, err = t2s.Convert(articleTitle)
	if err != nil {
		fmt.Println(err)
		return
	}
	content, err = t2s.Convert(content)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("章节名：", articleTitle, "创建时间：", createTime, "字数：", len(content))
	txt := fmt.Sprintf("%s %s 作者：%s\n%s\n\n", articleTitle, authorName, createTime, content)
	txtContent = append(txtContent, txt)
	atomic.AddInt32(&length, 1)
}

func DownloadTxt(ctx *spider.Context) {
	if atomic.CompareAndSwapInt32(&length, maxInContent, 0) {
		write := bufio.NewWriter(txtFile)
		_, _ = write.WriteString(strings.Join(txtContent, "\n"))
		// Flush将缓存的文件真正写入到文件中
		_ = write.Flush()
		txtContent = txtContent[:0]
	}
	return
}
