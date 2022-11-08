package spider

import (
	"errors"
	"fmt"
	"github.com/chuckwe/recircle/downloader"
	"sync"
	"time"
)

type MiddlewareHandler func(ctx *Context)
type MiddlewareHandlerErr func(ctx *Context) error

type Spider struct {
	UniqueKey     string                         // 唯一标识符
	STATUS        uint                           // 状态
	globalPreRun  MiddlewareHandlerErr           // 全局预备中间件
	Downloader    downloader.Downloader          // 下载器
	RuleHandlers  map[string][]MiddlewareHandler // 规则中间件
	CloseCallback func(s *Spider)                // 回调关闭
	timeTicker    time.Duration                  // 爬虫自我探活时间

	concurrent int
	resource   chan Resource
	signal     chan struct{}
	lock       sync.Locker // 锁
	wg         sync.WaitGroup
}

func NewSpider() *Spider {
	return &Spider{
		Downloader:   downloader.New(),
		lock:         &sync.Mutex{},
		RuleHandlers: make(map[string][]MiddlewareHandler),
		resource:     make(chan Resource, 1000),
		wg:           sync.WaitGroup{},
		timeTicker:   time.Second * 10,
	}
}

func (s *Spider) SetConcurrent(num int) *Spider {
	s.concurrent = num
	return s
}

// SetTimeTicker 设置探活时间 默认十秒
func (s *Spider) SetTimeTicker(num int) *Spider {
	s.timeTicker = time.Second * time.Duration(num)
	return s
}

// SetRules 设置爬虫key=规则名
func (s *Spider) SetRules(key string, h ...MiddlewareHandler) *Spider {
	if _, ok := s.RuleHandlers[key]; !ok {
		s.RuleHandlers[key] = make([]MiddlewareHandler, 0)
	}
	s.RuleHandlers[key] = append(s.RuleHandlers[key], h...)
	return s
}

func (s *Spider) SetGlobalPreRun(f MiddlewareHandlerErr) *Spider {
	s.globalPreRun = f
	return s
}

func (s *Spider) gPreRun() error {
	ctx := new(Context)
	if s.globalPreRun == nil {
		return nil
	}
	return s.globalPreRun(ctx)
}

func (s *Spider) Start() {
	defer func() {
		if err := recover(); err != nil {
			s.Start()
		}
	}()
	s.STATUS = 1
	err := s.gPreRun()
	if err != nil {
		panic(err)
	}

	if s.concurrent == 0 {
		s.concurrent = 1
	}
	s.wg.Add(s.concurrent)
	for i := s.concurrent; i > 0; i-- {
		go func(i int) {
			for {
				timer := time.NewTicker(s.timeTicker)
				select {
				case r, ok := <-s.resource:
					if !ok {
						s.wg.Done()
						return
					}
					err = s.run(r)
					if err != nil {
						fmt.Println(err)
						continue
					}
				case <-timer.C:
					s.wg.Done()
					return
				case <-s.signal:
					s.wg.Done()
					return
				}
			}
		}(i)
	}
	s.wg.Wait()
	s.Stop()
	return
}

func (s *Spider) Stop() {
	Sc.UnReg(s)
	s.STATUS = 0
	close(s.resource)
	s.CloseCallback(s)
	return
}

func (s *Spider) run(r Resource) (err error) {
	if _, ok := s.RuleHandlers[r.Rule]; !ok {
		return errors.New("无此规则")
	}
	ctx := initContext()
	defer func() {
		if err := recover(); err != nil {
			ctx.Cancel()
		}
	}()
	resp, err := s.Downloader.Download(r.Request)
	if err != nil {
		fmt.Println("下载失败：", err.Error())
		return
	}
	ctx.Response = resp

	for _, f := range s.RuleHandlers[r.Rule] {
		err = ctx.Ctx.Err()
		if err != nil {
			return
		}
		f(ctx)
	}
	return
}
