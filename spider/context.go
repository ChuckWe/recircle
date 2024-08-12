package spider

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

type Context struct {
	Ctx      context.Context
	Cancel   context.CancelFunc
	Response *http.Response
	text     []byte
	Temp     map[string]interface{}
	Info     map[string]string
}

func initContext(info map[string]string) *Context {
	ctx, cancel := context.WithCancel(context.Background())
	return &Context{
		Ctx:    ctx,
		Cancel: cancel,
		Temp:   make(map[string]interface{}),
		Info:   info,
	}
}

func (c *Context) GetDom() (*goquery.Document, error) {
	c.initText()
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(c.text))
	return doc, err
}

// GetBodyStr returns plain string crawled.
func (c *Context) initText() {
	var err error
	// 采用surf内核下载时，尝试自动转码
	var contentType, pageEncode string
	// 优先从响应头读取编码类型
	contentType = c.Response.Header.Get("Content-Type")
	if _, params, err := mime.ParseMediaType(contentType); err == nil {
		if cs, ok := params["charset"]; ok {
			pageEncode = strings.ToLower(strings.TrimSpace(cs))
		}
	}
	// 响应头未指定编码类型时，从请求头读取
	if len(pageEncode) == 0 {
		contentType = c.Response.Header.Get("Content-Type")
		if _, params, err := mime.ParseMediaType(contentType); err == nil {
			if cs, ok := params["charset"]; ok {
				pageEncode = strings.ToLower(strings.TrimSpace(cs))
			}
		}
	}

	switch pageEncode {
	// 不做转码处理
	case "utf8", "utf-8", "unicode-1-1-utf-8":
	default:
		// 指定了编码类型，但不是utf8时，自动转码为utf8
		// get converter to utf-8
		// Charset auto determine. Use golang.org/x/net/html/charset. Get response body and change it to utf-8
		var destReader io.Reader

		if len(pageEncode) == 0 {
			destReader, err = charset.NewReader(c.Response.Body, "")
		} else {
			destReader, err = charset.NewReaderLabel(pageEncode, c.Response.Body)
		}

		if err == nil {
			c.text, err = ioutil.ReadAll(destReader)
			if err == nil {
				c.Response.Body.Close()
				return
			} else {
				fmt.Printf(" *     [convert][%v]: (ignore transcoding)\n\n", err)
			}
		} else {
			fmt.Printf(" *     [convert][%v]: (ignore transcoding)\n\n", err)
		}
	}
	// 不做转码处理
	c.text, err = ioutil.ReadAll(c.Response.Body)
	c.Response.Body.Close()
	if err != nil {
		panic(err.Error())
		return
	}

}

func (c *Context) Abort() {
	c.Cancel()
}
