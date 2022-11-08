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
