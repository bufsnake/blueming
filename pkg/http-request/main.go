package http_request

import (
	"crypto/tls"
	"fmt"
	"github.com/bufsnake/blueming/config"
	"github.com/bufsnake/blueming/pkg/log"
	"github.com/bufsnake/blueming/pkg/useragent"
	"io/ioutil"
	"net/http"
	url2 "net/url"
	"os"
	"time"
)

func HTTPRequest(method, url, proxyx string, timeout int) (status int, contenttype, size string, body string, err error) {
	client := &http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: nil,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	transport := &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	if proxyx != "" {
		proxy, err := url2.Parse(proxyx)
		if err != nil {
			return 0, "", "", "", err
		}
		transport.Proxy = http.ProxyURL(proxy)
		client.Transport = transport
	} else {
		client.Transport = transport
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return 0, "", "0B", "", err
	}
	if method == http.MethodGet {
		req.Header.Add("Range", "bytes=0-8100")
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Referer", "http://www.baidu.com")
	req.Header.Add("Connection", "close")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("User-Agent", useragent.RandomUserAgent())
	do, err := client.Do(req)
	if err != nil {
		return 0, "", "0B", "", err
	}
	defer do.Body.Close()
	all, err := ioutil.ReadAll(do.Body)
	flag := false
	if err != nil {
		flag = true
	}
	temp := float64(do.ContentLength)
	SIZE := []string{"B", "K", "M", "G", "T"}
	i := 0
	for {
		if temp < 1024 {
			break
		}
		temp = temp / 1024.0
		i++
	}
	length := ""
	if i > len(SIZE) {
		length = fmt.Sprintf("%0.1fX", temp)
	} else {
		length = fmt.Sprintf("%0.1f%s", temp, SIZE[i])
	}
	if do.ContentLength > 104857600 {
		file, err := os.OpenFile(config.LogFileName, os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Warn(err)
		}
		_, err = file.WriteString(do.Header.Get("Content-Type") + " " + length + " " + url + "\n")
		if err != nil {
			file.Close()
			log.Warn(err)
		}
		file.Close()
		return 0, "", "0B", "", err
	}
	if flag { // GET 获取 body 失败时
		return do.StatusCode, do.Header.Get("Content-Type"), length, "", nil
	}
	return do.StatusCode, do.Header.Get("Content-Type"), length, string(all), nil
}
