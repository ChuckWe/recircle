package downloader

import (
	"net/http"
	"sync"
	"time"
)

type Request struct {
	Rule       string
	Url        string // url域名
	retry      bool   // 是否重试
	method     string
	Header     http.Header
	PostData   string // 参数
	retryTimes int    // 重试几次
	proxy      string
	// 是否使用cookies，在Spider的EnableCookie设置
	EnableCookie bool

	// dial tcp: i/o timeout
	DialTimeout time.Duration
	// WSARecv tcp: i/o timeout
	ConnTimeout time.Duration
	// the max times of download
	TryTimes int
	// how long pause when retry
	RetryPause time.Duration
	// max redirect times
	// when RedirectTimes equal 0, redirect times is ∞
	// when RedirectTimes less than 0, redirect times is 0
	RedirectTimes int
	lock          sync.Locker

	Cloudflare bool
	JsEnable   bool
	JsFunc     func(request *Request) (*http.Response, error)
}

// SetRetry 设置重试
func (r *Request) SetRetry(b bool, times int) *Request {
	r.retry = b
	if times == 0 {
		times = 1
	}
	r.retryTimes = times
	return r
}

// GetRetry 获取重试
func (r *Request) GetRetry() (b bool, times int) {
	return r.retry, r.retryTimes
}

// SetMethod 请求方式
func (r *Request) SetMethod(s string) *Request {
	if len(s) == 0 {
		s = "GET"
	}
	r.method = s
	return r
}

// GetMethod 获取请求方式
func (r *Request) GetMethod() (s string) {
	if len(r.method) == 0 {
		r.method = "GET"
	}
	return r.method
}

// SetProxy 代理
func (r *Request) SetProxy(s string) *Request {
	r.proxy = s
	return r
}

// GetProxy 获取代理方式
func (r *Request) GetProxy() (s string) {
	return r.proxy
}

// dial tcp: i/o timeout
func (r *Request) GetDialTimeout() time.Duration {
	return r.DialTimeout
}

// WSARecv tcp: i/o timeout
func (r *Request) GetConnTimeout() time.Duration {
	return r.ConnTimeout
}

// the max times of download
func (r *Request) GetTryTimes() int {
	return r.TryTimes
}

// the pause time of retry
func (r *Request) GetRetryPause() time.Duration {
	return r.RetryPause
}

// max redirect times
func (r *Request) GetRedirectTimes() int {
	return r.RedirectTimes
}
