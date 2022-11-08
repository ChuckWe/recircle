package spider

import (
	"fmt"
	"github.com/chuckwe/recircle/downloader"
	"sync"
	"time"
)

var Sc = NewSchedule()

type Resource struct {
	SpiderUniqueKey string
	*downloader.Request
}

func NewResource(key, rule, link string) Resource {
	return Resource{
		SpiderUniqueKey: key,
		Request: &downloader.Request{
			Rule:        rule,
			Url:         link,
			DialTimeout: time.Second * 3,
			ConnTimeout: time.Second * 2,
		},
	}
}

type Schedule struct {
	ResourcePoolList chan Resource
	ConcurrentNum    int                // 并发数量
	spiders          map[string]*Spider // 爬虫
	Wg               sync.WaitGroup
	lock             sync.Locker
}

func NewSchedule() *Schedule {
	return &Schedule{
		ResourcePoolList: make(chan Resource, 1000),
		lock:             &sync.Mutex{},
		spiders:          make(map[string]*Spider),
		Wg:               sync.WaitGroup{},
	}
}

func (s *Schedule) AddResource(resource Resource) *Schedule {
	s.ResourcePoolList <- resource
	return s
}

func (s *Schedule) Close() {
	close(s.ResourcePoolList)
}

func (s *Schedule) Register(spider *Spider) *Schedule {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.spiders[spider.UniqueKey] = spider
	return s
}

func (s *Schedule) SetConcurrent(num int) *Schedule {
	s.ConcurrentNum = num
	return s
}

func (s *Schedule) Init() {
	defer func() {
		if err := recover(); err != nil {
			s.Init()
		}
	}()
	// 调度超时直接关闭调度
	if s.ConcurrentNum == 0 {
		s.ConcurrentNum = 1
	}
	s.Wg.Add(s.ConcurrentNum)
	for i := s.ConcurrentNum; i > 0; i-- {
		go func(i int) {
			for {
				timer := time.NewTimer(time.Second * 10)
				select {
				case r, ok := <-s.ResourcePoolList:
					if !ok {
						fmt.Println("关闭爬虫:", i)
						s.Wg.Done()
						return
					}
					s.scheduleSpider(r)
				case <-timer.C:
					s.Close()
					fmt.Println("超时关闭爬虫调度")
				}
			}
		}(i)
	}

}

func (s *Schedule) scheduleSpider(r Resource) {
	s.lock.Lock()
	defer s.lock.Unlock()
	spider, ok := s.spiders[r.SpiderUniqueKey]
	if !ok {
		fmt.Println("暂无爬虫")
		return
	}
	spider.Run(r)
}
