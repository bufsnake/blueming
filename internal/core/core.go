package core

import (
	"fmt"
	"github.com/bufsnake/blueming/config"
	file_download "github.com/bufsnake/blueming/pkg/file-download"
	general_file_name "github.com/bufsnake/blueming/pkg/general-file-name"
	http_request "github.com/bufsnake/blueming/pkg/http-request"
	"github.com/bufsnake/blueming/pkg/log"
	. "github.com/logrusorgru/aurora"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
)

type core struct {
	config        config.Config
	url           []string
	excludestatus []int
	wordlist      []string
}

func NewCore(url []string, config config.Config) core {
	statuss := make([]int, 0)
	split := strings.Split(config.ExcludeStatus, ",")
	for i := 0; i < len(split); i++ {
		if len(split[i]) == 0 {
			continue
		}
		status, err := strconv.Atoi(split[i])
		if err != nil {
			log.Fatal(split[i], err)
		}
		statuss = append(statuss, status)
	}
	wordlist := make([]string, 0)
	if config.Wordlist != "" {
		file, err := ioutil.ReadFile(config.Wordlist)
		if err != nil {
			log.Fatal(err)
		}
		split := strings.Split(string(file), "\n")
		for i := 0; i < len(split); i++ {
			split[i] = strings.Trim(split[i], "\r")
			if split[i] == "" || split[i] == "/" {
				continue
			}
			wordlist = append(wordlist, split[i])
		}
	}
	return core{url: url, config: config, excludestatus: statuss, wordlist: wordlist}
}

func (c *core) Core() {
	index := 0 //
again:
	if c.config.Wordlist != "" && len(c.wordlist) == index {
		return
	}
	requestlist := make([][]string, 0)
	for i := 0; i < len(c.url); i++ {
		genURL, err := general_file_name.NewGenURL(c.url[i], c.wordlist[index])
		if err != nil {
			log.Warn(err)
			continue
		}
		getURL := genURL.GetURL()
		requestlist = append(requestlist, *getURL)
	}
	if len(requestlist) == 0 {
		index++
		goto again
	}
	httpr := sync.WaitGroup{}
	httpc := make(chan string, c.config.Thread)
	httpd := make(chan string, c.config.Thread)
	for i := 0; i < c.config.Thread; i++ {
		httpr.Add(1)
		go c.httprequest(&httpr, httpc, httpd, c.config.Timeout)
	}
	down := sync.WaitGroup{}
	for i := 0; i < c.config.Thread; i++ {
		down.Add(1)
		go c.httpdownload(&down, httpd)
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

	index++
	goto again
}

func (c *core) httprequest(wait *sync.WaitGroup, httpc, httpd chan string, timeout int) {
	defer wait.Done()
	for url := range httpc {
		log.Trace(url)
		status, ct, size, err := http_request.HTTPRequest(url, timeout)
		if err != nil && c.config.Wordlist == "" {
			log.Warn(err)
			continue
		}
		if c.config.Wordlist != "" {
			flag := true
			for i := 0; i < len(c.excludestatus); i++ {
				if status == c.excludestatus[i] {
					flag = false
				}
			}
			if flag && status != 0 {
				pr := make([]interface{}, 0)
				if status >= 200 && status < 300 {
					pr = []interface{}{BrightGreen(status).String()}
				} else if status >= 300 && status < 400 {
					pr = []interface{}{BrightYellow(status).String()}
				} else if status >= 400 && status < 500 {
					pr = []interface{}{BrightMagenta(status).String()}
				} else {
					pr = []interface{}{BrightWhite(status).String()}
				}
				pr = append(pr, BrightCyan(fmt.Sprintf("%8s", size)).String())
				if ct != "" {
					pr = append(pr, ct)
				}
				pr = append(pr, BrightWhite(url))
				fmt.Println(pr...)
			}
			continue
		}
		if status == 200 && (size == "0B" || size == "0.0B") {
			log.Debug(status, size, ct, url)
		}
		if size != "0B" && size != "0.0B" && status == 200 {
			log.Info(size, ct, url)
			httpd <- url
		}
	}
}

func (c *core) httpdownload(wait *sync.WaitGroup, httpd chan string) {
	defer wait.Done()
	for url := range httpd {
		err := file_download.DownloadFile(url, c.config.Proxy)
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
