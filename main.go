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
