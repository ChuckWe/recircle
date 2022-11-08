package downloader

import (
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/chuckwe/recircle/downloader/agent"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"
)

type Downloader interface {
	Download(request *Request) (*http.Response, error)
}

var dnsCache = &DnsCache{ipPortLib: sync.Map{}}

// DnsCache DNS cache
type DnsCache struct {
	ipPortLib sync.Map
}

// Reg registers ipPort to DNS cache.
func (d *DnsCache) Reg(addr, ipPort string) {
	d.ipPortLib.Store(addr, ipPort)
}

// Del deletes ipPort from DNS cache.
func (d *DnsCache) Del(addr string) {
	d.ipPortLib.Delete(addr)
}

// Query queries ipPort from DNS cache.
func (d *DnsCache) Query(addr string) (string, bool) {
	ipPort, ok := d.ipPortLib.Load(addr)
	if !ok {
		return "", false
	}
	return ipPort.(string), true
}

type DefaultDownloader struct {
	CookieJar *cookiejar.Jar
}

func New(cookie ...*cookiejar.Jar) Downloader {
	d := &DefaultDownloader{}
	if len(cookie) != 0 {
		d.CookieJar = cookie[0]
	} else {
		d.CookieJar, _ = cookiejar.New(nil)
	}
	return d
}

func (d *DefaultDownloader) Download(request *Request) (*http.Response, error) {
	param, err := NewParam(request)
	if err != nil {
		fmt.Println("param request err :", err)
		return nil, err
	}
	param.header.Set("Connection", "close")
	if err != nil {
		return nil, err
	}
	param.client = d.buildClient(param)
	resp, err := d.httpRequest(param)

	if err != nil {
		return nil, err
	}
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		var gzipReader *gzip.Reader
		gzipReader, err = gzip.NewReader(resp.Body)
		if err == nil {
			resp.Body = gzipReader
		}

	case "deflate":
		resp.Body = flate.NewReader(resp.Body)

	case "zlib":
		var readCloser io.ReadCloser
		readCloser, err = zlib.NewReader(resp.Body)
		if err == nil {
			resp.Body = readCloser
		}
	}

	resp = param.writeback(resp)

	return resp, nil
}

func (d *DefaultDownloader) buildClient(param *Param) *http.Client {
	client := &http.Client{
		CheckRedirect: param.checkRedirect,
	}

	if param.enableCookie {
		client.Jar = d.CookieJar
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			var (
				c          net.Conn
				err        error
				ipPort, ok = dnsCache.Query(addr)
			)
			if !ok {
				ipPort = addr
				defer func() {
					if err == nil {
						dnsCache.Reg(addr, c.RemoteAddr().String())
					}
				}()
			} else {
				defer func() {
					if err != nil {
						dnsCache.Del(addr)
					}
				}()
			}
			c, err = net.DialTimeout(network, ipPort, param.dialTimeout)
			if err != nil {
				fmt.Println("net dial timeout :", err)
				return nil, err
			}
			if param.connTimeout > 0 {
				_ = c.SetDeadline(time.Now().Add(param.connTimeout))
			}
			return c, nil
		},
	}

	if param.proxy != nil {
		transport.Proxy = http.ProxyURL(param.proxy)
	}

	if strings.ToLower(param.url.Scheme) == "https" {
		transport.TLSClientConfig = &tls.Config{RootCAs: nil, InsecureSkipVerify: true}
		transport.DisableCompression = true
	}
	client.Transport = transport
	return client
}

// httpRequest send uses the given *http.Request to make an HTTP request.
func (d *DefaultDownloader) httpRequest(param *Param) (resp *http.Response, err error) {
	req, err := http.NewRequest(param.method, param.url.String(), param.body)
	if err != nil {
		return nil, err
	}

	req.Header = param.header

	if param.tryTimes <= 0 {
		resp, err = param.client.Do(req)
		if err != nil {
			if !param.enableCookie {
				l := len(agent.UserAgents["common"])
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				req.Header.Set("User-Agent", agent.UserAgents["common"][r.Intn(l)])
			}
			fmt.Println("client do timeout :", err)
		}
	} else {
		for i := 0; i < param.tryTimes; i++ {
			resp, err = param.client.Do(req)
			if err != nil {
				if !param.enableCookie {
					l := len(agent.UserAgents["common"])
					r := rand.New(rand.NewSource(time.Now().UnixNano()))
					req.Header.Set("User-Agent", agent.UserAgents["common"][r.Intn(l)])
				}
				time.Sleep(param.retryPause)
				continue
			}
			break
		}
	}

	return resp, err
}
