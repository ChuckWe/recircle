package downloader

import (
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"github.com/imroc/req/v3"
	"io"
	"net/http"
	"net/http/cookiejar"
)

type ImrocDownloader struct {
	CookieJar *cookiejar.Jar
}

func NewImroc(cookie *cookiejar.Jar) Downloader {
	d := &ImrocDownloader{}
	if cookie != nil {
		d.CookieJar = cookie
		return d
	}
	d.CookieJar, _ = cookiejar.New(nil)
	return d
}

func (r *ImrocDownloader) Download(request *Request) (*http.Response, error) {

	if request.JsEnable {
		return request.JsFunc(request)
	}
	clientC := req.C()
	if request.Cloudflare {
		// l := len(agent.UserAgents["common"])
		// randUserAgent := rand.New(rand.NewSource(time.Now().UnixNano()))
		// clientC = clientC.SetUserAgent(agent.UserAgents["common"][randUserAgent.Intn(l)])
		clientC = clientC.SetUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")
		clientC = clientC.SetTLSFingerprintChrome()
		clientC = clientC.ImpersonateChrome()
	}
	request.Header = make(http.Header)
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")
	param, err := NewParam(request)
	if err != nil {
		fmt.Println("param request err :", err)
		return nil, err
	}
	if len(request.proxy) > 0 {
		clientC = clientC.SetProxyURL(request.proxy)
	}
	if request.EnableCookie {
		clientC = clientC.SetCookieJar(r.CookieJar)
	}
	if request.TryTimes > 0 {
		clientC = clientC.SetCommonRetryCount(request.retryTimes)
	}

	if request.DialTimeout > 0 {
		clientC = clientC.SetTimeout(request.DialTimeout)
	}
	if request.RedirectTimes > 0 {
		clientC = clientC.SetRedirectPolicy(req.MaxRedirectPolicy(request.RedirectTimes))
	}

	requestR := clientC.R().SetBody(param.body)
	if len(request.PostData) > 0 {
		requestR.SetHeader("Content-Type", "application/x-www-form-urlencoded")
		requestR = requestR.SetBody(request.PostData)
	}
	param.header.Set("Connection", "close")
	// requestR.Headers = param.header
	requestR.Method = param.method
	respR, err := requestR.Send(param.method, param.url.String())
	if err != nil {
		return nil, err
	}

	switch respR.GetHeader("Content-Encoding") {
	case "gzip":
		var gzipReader *gzip.Reader
		gzipReader, err = gzip.NewReader(respR.Body)
		if err == nil {
			respR.Body = gzipReader
		}

	case "deflate":
		respR.Body = flate.NewReader(respR.Body)

	case "zlib":
		var readCloser io.ReadCloser
		readCloser, err = zlib.NewReader(respR.Body)
		if err == nil {
			respR.Body = readCloser
		}
	}

	resp := param.writeback(respR.Response)

	return resp, nil
}
