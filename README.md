# Describe
一个简单的网络爬虫spider,书写具体业务即可正常爬行。

# Usage

~~~
git clone https://github.com/ChuckWe/recircle.git 
cd recircle
go mod tidy
~~~
已经自动注册demo文件夹所以直接在demo文件夹下面写业务逻辑即可。
~~~
package main

import (
	_ "github.com/chuckwe/recircle/demo"
	"github.com/chuckwe/recircle/spider"
)

func main() {
	spider.Sc.Init()
}

~~~
demo文件夹中存放一个案例，当然需要你有科学上网的方式才可爬行。
~~~
package demo

import (
	"bufio"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chuckwe/recircle/spider"
	"os"
	"strings"
	"sync/atomic"
)

var (
	txtFile, _   = os.OpenFile("明克街13号.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	length       int32
	txtContent   = make([]string, 0)
	maxInContent = int32(10)
)

func init() {
	demoSpider := spider.NewSpider()
	demoSpider.UniqueKey = "691shu"
	demoSpider.SetRules("列表", ListQuery)
	demoSpider.SetRules("详情页", DetailQuery, DownloadTxt)
	demoSpider.CloseCallback = func(s *spider.Spider) {
		_ = txtFile.Close()
		return
	}
	spider.Sc.Register(demoSpider)
	r := spider.NewResource("69shu", "列表", "https://www.69shu.com/38938/")
	r.SetProxy("http://127.0.0.1:1087")
	err := spider.Sc.AddResource(r)
	if err != nil {
		panic(err)
	}
}

func ListQuery(ctx *spider.Context) {
	query, _ := ctx.GetDom()
	box := query.Find("#catalog li")
	box.Each(func(i int, selection *goquery.Selection) {
		link, ok := selection.Find("a").Attr("href")
		if !ok {
			return
		}
		fmt.Println(link)
		r := spider.NewResource("69shu", "详情页", link)
		r.SetProxy("http://127.0.0.1:1087")
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
	title := query.Find("h1").Text()
	if title == "" {
		return
	}
	createTime := query.Find(".txtinfo>span").Text()
	t := query.Find(".txtnav").Contents().FilterFunction(func(i int, selection *goquery.Selection) bool {
		return i > 5 && i < query.Find(".txtnav").Contents().Length()-2
	})
	content := t.Text()
	fmt.Println(title, createTime, len(content))
	txt := fmt.Sprintf("%s  %s\n%s\n\n", title, createTime, content)
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

~~~

# Thanks
- 部分代码copy自🙏 [@andeya/pholcus](https://github.com/andeya/pholcus)
