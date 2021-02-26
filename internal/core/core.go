package core

import (
	file_download "github.com/bufsnake/blueming/pkg/file-download"
	general_file_name "github.com/bufsnake/blueming/pkg/general-file-name"
	http_request "github.com/bufsnake/blueming/pkg/http-request"
	"github.com/bufsnake/blueming/pkg/log"
	"io/ioutil"
	"sync"
)

func NewCore(url []string, thread, timeout int, proxy string) {
	requestlist := make([][]string, 0)
	for i := 0; i < len(url); i++ {
		genURL, err := general_file_name.NewGenURL(url[i])
		if err != nil {
			log.Warn(err)
			continue
		}
		getURL := genURL.GetURL()
		requestlist = append(requestlist, *getURL)
	}
	httpr := sync.WaitGroup{}
	httpc := make(chan string, thread)
	httpd := make(chan string, thread)
	for i := 0; i < thread; i++ {
		httpr.Add(1)
		go httprequest(&httpr, httpc, httpd, timeout)
	}
	down := sync.WaitGroup{}
	for i := 0; i < thread; i++ {
		down.Add(1)
		go httpdownload(&down, httpd, proxy)
	}
	i := 0
	for {
		flag := 0
		for j := 0; j < len(requestlist); j++ {
			if i >= len(requestlist[j]) {
				flag++
				continue
			}
			httpc <- requestlist[j][i]
		}
		i++
		if flag == len(requestlist) {
			break
		}
	}

	close(httpc)
	httpr.Wait()
	close(httpd)
	down.Wait()
}

func httprequest(wait *sync.WaitGroup, httpc, httpd chan string, timeout int) {
	defer wait.Done()
	for url := range httpc {
		log.Trace(url)
		status, size, err := http_request.HTTPRequest(url, timeout)
		if err != nil {
			log.Warn(err)
			continue
		}
		if status == 200 && (size == "0B" || size == "0.0B") {
			log.Debug(status, size, url)
		}
		if size != "0B" && size != "0.0B" && status == 200 {
			log.Info(size, url)
			httpd <- url
		}
	}
}

func httpdownload(wait *sync.WaitGroup, httpd chan string, proxy string) {
	defer wait.Done()
	for url := range httpd {
		err := file_download.DownloadFile(url, proxy)
		if err != nil {
			log.Info("file download error", err)
			// 将URL保存到文件
			err := ioutil.WriteFile("output/download_error", []byte(url+"\n"), 644)
			if err != nil {
				log.Warn(err)
			}
		}
	}
}
