package downloader

import (
	"bytes"
	"fmt"
	"github.com/chuckwe/recircle/downloader/agent"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Param struct {
	method        string
	url           *url.URL
	proxy         *url.URL
	body          io.Reader
	header        http.Header
	enableCookie  bool
	dialTimeout   time.Duration
	connTimeout   time.Duration
	tryTimes      int
	retryPause    time.Duration
	redirectTimes int
	client        *http.Client
}

func UrlEncode(urlStr string) (*url.URL, error) {
	urlObj, err := url.Parse(urlStr)
	urlObj.RawQuery = urlObj.Query().Encode()
	return urlObj, err
}

func NewParam(req *Request) (param *Param, err error) {
	param = new(Param)
	param.url, err = UrlEncode(req.Url)
	if err != nil {
		return nil, err
	}

	if req.GetProxy() != "" {
		if param.proxy, err = url.Parse(req.GetProxy()); err != nil {
			return nil, err
		}
	}

	param.header = req.Header
	if param.header == nil {
		param.header = make(http.Header)
	}

	switch method := strings.ToUpper(req.GetMethod()); method {
	case "GET", "HEAD":
		param.method = method
	case "POST":
		param.method = method
		param.header.Add("Content-Type", "application/x-www-form-urlencoded")
		param.body = strings.NewReader(req.PostData)
	case "POST-M":
		param.method = "POST"
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		values, _ := url.ParseQuery(req.PostData)
		for k, vs := range values {
			for _, v := range vs {
				writer.WriteField(k, v)
			}
		}
		err := writer.Close()
		if err != nil {
			return nil, err
		}
		param.header.Add("Content-Type", writer.FormDataContentType())
		param.body = body

	default:
		param.method = "GET"
	}

	param.enableCookie = req.EnableCookie

	if len(param.header.Get("User-Agent")) == 0 {
		if param.enableCookie {
			param.header.Add("User-Agent", agent.UserAgents["common"][0])
		} else {
			l := len(agent.UserAgents["common"])
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			param.header.Add("User-Agent", agent.UserAgents["common"][r.Intn(l)])
		}
	}

	param.dialTimeout = req.GetDialTimeout()
	if param.dialTimeout < 0 {
		param.dialTimeout = 0
	}

	param.connTimeout = req.GetConnTimeout()
	b, times := req.GetRetry()
	if b {
		param.tryTimes = times
	}
	param.retryPause = req.GetRetryPause()
	param.redirectTimes = req.GetRedirectTimes()
	return
}

// 回写Request内容
func (self *Param) writeback(resp *http.Response) *http.Response {
	if resp == nil {
		resp = new(http.Response)
		resp.Request = new(http.Request)
	} else if resp.Request == nil {
		resp.Request = new(http.Request)
	}

	if resp.Header == nil {
		resp.Header = make(http.Header)
	}

	resp.Request.Method = self.method
	resp.Request.Header = self.header
	resp.Request.Host = self.url.Host

	return resp
}

// checkRedirect is used as the value to http.Client.CheckRedirect
// when redirectTimes equal 0, redirect times is ∞
// when redirectTimes less than 0, not allow redirects
func (self *Param) checkRedirect(req *http.Request, via []*http.Request) error {
	if self.redirectTimes == 0 {
		return nil
	}
	if len(via) >= self.redirectTimes {
		if self.redirectTimes < 0 {
			return fmt.Errorf("not allow redirects.")
		}
		return fmt.Errorf("stopped after %v redirects.", self.redirectTimes)
	}
	return nil
}
