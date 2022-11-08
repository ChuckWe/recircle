package spider

import (
	"fmt"
	"github.com/chuckwe/recircle/downloader"
	"sync"
)

type MiddlewareHandler func(ctx *Context)
type MiddlewareHandlerErr func(ctx *Context) error

type Spider struct {
	UniqueKey    string // 唯一标识符
	globalPreRun MiddlewareHandlerErr
	Downloader   downloader.Downloader
	RuleHandlers map[string][]MiddlewareHandler
	lock         sync.Locker
}

func NewSpider() *Spider {
	return &Spider{
		Downloader:   downloader.New(),
		lock:         &sync.Mutex{},
		RuleHandlers: make(map[string][]MiddlewareHandler),
	}
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
	return s.globalPreRun(ctx)
}

func (s *Spider) Start() (err error) {
	err = s.gPreRun()
	if err != nil {
		return
	}
	return
}

func (s *Spider) Run(r Resource) {
	if _, ok := s.RuleHandlers[r.Rule]; !ok {
		fmt.Println("无此规则：", r.Rule)
		return
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
		err := ctx.Ctx.Err()
		if err != nil {
			fmt.Println(err)
			return
		}
		f(ctx)
	}
}
