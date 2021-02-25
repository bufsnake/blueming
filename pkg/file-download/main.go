package file_download

import (
	"crypto/tls"
	"github.com/bufsnake/blueming/pkg/useragent"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func DownloadFile(url string) error {
	client := &http.Client{
		Timeout: 1 * time.Hour,
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
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Referer", "http://www.baidu.com")
	req.Header.Add("Connection", "close")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("User-Agent", useragent.RandomUserAgent())
	do, err := client.Do(req)
	if err != nil {
		return err
	}
	defer do.Body.Close()
	temp_file := strings.ReplaceAll(url, ":", ".")
	temp_file = strings.ReplaceAll(temp_file, "/", ".")
	temp_file = strings.ReplaceAll(temp_file, "..", ".")
	temp_file = strings.ReplaceAll(temp_file, "..", ".")
	temp_file = strings.ReplaceAll(temp_file, "..", ".")
	out, err := os.Create("output/" + temp_file)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, do.Body)
	if err != nil {
		return err
	}
	return nil
}
