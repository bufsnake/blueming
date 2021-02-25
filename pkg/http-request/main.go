package http_request

import (
	"crypto/tls"
	"fmt"
	"github.com/bufsnake/blueming/pkg/useragent"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func HTTPRequest(url string, timeout int) (status int, size string, err error) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return 0, "0B", err
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Referer", "http://www.baidu.com")
	req.Header.Add("Connection", "close")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("User-Agent", useragent.RandomUserAgent())
	do, err := client.Do(req)
	if err != nil {
		return 0, "0B", err
	}
	defer do.Body.Close()
	_, err = io.Copy(ioutil.Discard, do.Body)
	if err != nil {
		return 0, "0B", err
	}
	if !strings.Contains(do.Header.Get("Content-Type"), "application/octet-stream") {
		return 0, "0B", err
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
		length = fmt.Sprintf("%0.2fX", temp)
	} else {
		length = fmt.Sprintf("%0.1f%s", temp, SIZE[i])
	}
	return do.StatusCode, length, nil
}
