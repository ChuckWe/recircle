package spider

import (
	"errors"
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
	lock             sync.Locker
	ticker           *time.Ticker
}

func NewSchedule() *Schedule {
	return &Schedule{
		lock:    &sync.Mutex{},
		spiders: make(map[string]*Spider),
		ticker:  time.NewTicker(time.Second),
	}
}

func (s *Schedule) AddResource(resource Resource) (err error) {
	sp, ok := s.spiders[resource.SpiderUniqueKey]
	if !ok {
		return errors.New(fmt.Sprintf("----------暂无[%s]爬虫----------", resource.SpiderUniqueKey))
	}
	sp.resource <- resource
	return
}

func (s *Schedule) Close() {
	return
}

func (s *Schedule) Register(spider *Spider) *Schedule {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.spiders[spider.UniqueKey] = spider
	return s
}

func (s *Schedule) UnReg(spider *Spider) *Schedule {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.spiders, spider.UniqueKey)
	return s
}

func (s *Schedule) Init() {
	for _, sp := range s.spiders {
		go sp.Start()
	}
	for {
		select {
		case <-s.ticker.C:
			s.lock.Lock()
			if len(s.spiders) == 0 {
				fmt.Println("所有爬虫运行完毕,调度器自动退出...")
				s.lock.Unlock()
				s.Close()
				return
			}
			s.lock.Unlock()
		}
	}
}
