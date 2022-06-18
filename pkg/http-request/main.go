package http_request

import (
	"crypto/tls"
	"fmt"
	"github.com/bufsnake/blueming/pkg/useragent"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"time"
)

func HTTPRequest(urlstr, proxyx string, timeout int) (status int, contenttype, size string, body string, err error) {
	cli := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	transport := http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	if proxyx != "" {
		transport.Proxy = func(request *http.Request) (*url.URL, error) {
			return url.Parse(proxyx)
		}
	}
	cli.Transport = &transport
	req, err := http.NewRequest(http.MethodGet, urlstr, nil)
	if err != nil {
		return 0, "", "0B", "", err
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Referer", "https://www.baidu.com/")
	req.Header.Add("Connection", "close")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("User-Agent", useragent.RandomUserAgent())
	res, err := cli.Do(req)
	if err != nil {
		return 0, "", "0B", "", err
	}
	defer res.Body.Close()
	if res.StatusCode == 101 {
		return 0, "", "0B", "", err
	}
	resbody := make([]byte, 0)
	resbody, err = ioutil.ReadAll(io.LimitReader(res.Body, 1024*1024*2))
	if err != nil {
		return 0, "", "0B", "", err
	}
	content_length := float64(res.ContentLength)
	if content_length < 1 {
		content_length = float64(len(resbody))
	}
	SIZE := []string{"B", "K", "M", "G", "T"}
	i := 0
	for {
		if content_length < 1024 {
			break
		}
		content_length = content_length / 1024.0
		i++
	}
	length := ""
	if i > len(SIZE) {
		length = fmt.Sprintf("%0.1fX", content_length)
	} else {
		length = fmt.Sprintf("%0.1f%s", content_length, SIZE[i])
	}
	contenttype, _, err = mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err != nil {
		contenttype = res.Header.Get("Content-Type")
	}
	return res.StatusCode, contenttype, length, string(resbody), nil
}
