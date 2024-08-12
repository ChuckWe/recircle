package downloader

import (
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"crypto/tls"
	"fmt"
	tlsclient "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/chuckwe/recircle/downloader/transport"
	"github.com/go-resty/resty/v2"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

type RestyDownloader struct {
	CookieJar *cookiejar.Jar
}

func NewResty(cookie *cookiejar.Jar) Downloader {
	d := &RestyDownloader{}
	if cookie != nil {
		d.CookieJar = cookie
		return d
	}
	d.CookieJar, _ = cookiejar.New(nil)
	return d
}

func (r *RestyDownloader) Download(request *Request) (*http.Response, error) {

	if request.JsEnable {
		return request.JsFunc(request)
	}

	clientR := resty.New()
	if request.Cloudflare {
		tr := &transport.SpoofedRoundTripper{}
		var err error
		if len(request.proxy) > 0 {
			tr, err = transport.NewSpoofedRoundTripper(
				tlsclient.WithRandomTLSExtensionOrder(),
				tlsclient.WithClientProfile(profiles.Chrome_120),
				tlsclient.WithProxyUrl(request.proxy),
			)
		} else {
			tr, err = transport.NewSpoofedRoundTripper(
				tlsclient.WithRandomTLSExtensionOrder(),
				tlsclient.WithClientProfile(profiles.Chrome_120),
			)
		}

		if err != nil {
			fmt.Println("build tr err :", err)
			return nil, err
		}
		clientR = clientR.SetTransport(tr)
	}
	param, err := NewParam(request)
	if err != nil {
		fmt.Println("param request err :", err)
		return nil, err
	}
	if !request.Cloudflare && len(request.proxy) > 0 {
		clientR = clientR.SetProxy(request.proxy)
	}
	if request.EnableCookie {
		clientR = clientR.SetCookieJar(r.CookieJar)
	}
	if request.TryTimes > 0 {
		clientR = clientR.SetRetryCount(request.retryTimes)
		clientR = clientR.SetRetryWaitTime(time.Second)
	}
	if request.ConnTimeout > 0 {
		clientR = clientR.SetRetryWaitTime(request.ConnTimeout)
	}
	if request.DialTimeout > 0 {
		clientR = clientR.SetTimeout(request.DialTimeout)
	}
	if request.RedirectTimes > 0 {
		clientR = clientR.SetRedirectPolicy(resty.FlexibleRedirectPolicy(request.RedirectTimes))
	}
	if strings.ToLower(param.url.Scheme) == "https" {
		clientR = clientR.SetTLSClientConfig(&tls.Config{RootCAs: nil, InsecureSkipVerify: true})
		clientR = clientR.EnableTrace()
		transport, err := clientR.Transport()
		if err != nil {
			return nil, err
		}
		transport.DisableCompression = true
		clientR = clientR.SetTransport(transport)
	}

	requestR := clientR.R().SetBody(param.body)
	param.header.Set("Connection", "close")
	requestR.Header = param.header
	requestR.Method = param.method
	requestR.URL = param.url.String()

	respR, err := requestR.Send()

	if err != nil {
		return nil, err
	}

	switch respR.Header().Get("Content-Encoding") {
	case "gzip":
		var gzipReader *gzip.Reader
		gzipReader, err = gzip.NewReader(respR.RawBody())
		if err == nil {
			respR.RawResponse.Body = gzipReader
		}

	case "deflate":
		respR.RawResponse.Body = flate.NewReader(respR.RawBody())

	case "zlib":
		var readCloser io.ReadCloser
		readCloser, err = zlib.NewReader(respR.RawBody())
		if err == nil {
			respR.RawResponse.Body = readCloser
		}
	}

	resp := param.writeback(respR.RawResponse)

	return resp, nil
}
