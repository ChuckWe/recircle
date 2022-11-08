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
	spider.Sc.SetConcurrent(2)
	spider.Sc.Init()
	spider.Sc.Wg.Wait()
}

~~~
demo文件夹中存放一个案例，当然需要你有科学上网的方式才可爬行。
~~~
package demo

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chuckwe/recircle/spider"
)

func init() {
	demoSpider := spider.NewSpider()
	demoSpider.UniqueKey = "56shu"
	demoSpider.SetRules("列表", ListQuery)
	demoSpider.SetRules("详情页", DetailQuery)
	spider.Sc.Register(demoSpider)
	r := spider.NewResource("69shu", "列表", "https://www.69shu.com/38938/")
	spider.Sc.AddResource(r)
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
		spider.Sc.AddResource(r)
	})
}

func DetailQuery(ctx *spider.Context) {
	query, _ := ctx.GetDom()
	title := query.Find("h1").Text()
	fmt.Println(title)
	if title == "" {
		return
	}
	creatTime := query.Find(".txtinfo>span").Text()
	fmt.Println(title, creatTime)
	t := query.Find(".txtnav").Contents().FilterFunction(func(i int, selection *goquery.Selection) bool {
		return i > 5 && i < query.Find(".txtnav").Contents().Length()-2
	})
	content := t.Text()
	fmt.Println(title, creatTime, len(content))
}
~~~

# Thanks
- 部分代码copy自🙏 [@andeya/pholcus](https://github.com/andeya/pholcus)
