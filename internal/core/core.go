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
	url2 "net/url"
	"strconv"
	"strings"
	"sync"
)

type core struct {
	config        config.Config
	url           []string
	excludestatus []int
	wordlist      []string
	first         map[string]string
	lock          sync.Mutex
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
		wordlist = append(wordlist, "/admino.ini")
		split := strings.Split(string(file), "\n")
		flag := false
		if config.Index == "" {
			flag = true
		}
		for i := 0; i < len(split); i++ {
			split[i] = strings.Trim(split[i], "\r")
			if split[i] == "" || split[i] == "/" {
				continue
			}
			if flag {
				wordlist = append(wordlist, split[i])
			}
			if split[i] == config.Index {
				flag = true
			}
		}
		if len(wordlist) == 0 {
			log.Fatal("specify index not found")
		}
	}
	return core{url: url, config: config, excludestatus: statuss, wordlist: wordlist}
}

func (c *core) Core() {
	c.first = make(map[string]string)
	index := 0 //
again:
	if c.config.Wordlist != "" && len(c.wordlist) == index {
		return
	}
	requestlist := make([][]string, 0)
min:
	for i := 0; i < len(c.url); i++ {
		uri := ""
		if len(c.wordlist) != 0 {
			uri = c.wordlist[index]
		}
		genURL, err := general_file_name.NewGenURL(c.url[i], uri)
		if err != nil {
			log.Warn(err)
			continue
		}
		getURL := genURL.GetURL()
		requestlist = append(requestlist, *getURL)
	}
	if len(c.url) < 10 && len(c.wordlist) != 0 && index < len(c.wordlist)-1 {
		index++
		goto min
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

	if c.config.Wordlist == "" {
		return
	}
	index++
	goto again
}

func (c *core) httprequest(wait *sync.WaitGroup, httpc, httpd chan string, timeout int) {
	defer wait.Done()
	for url := range httpc {
		parse, _ := url2.Parse(url)
		log.Trace(url)
		status, ct, size, err := http_request.HTTPRequest(url, timeout)
		c.lock.Lock()
		if err != nil && c.config.Wordlist == "" {
			if parse.Path == "/admino.ini" {
				c.first[parse.Scheme+parse.Host] = "err"
			}
			log.Warn(err)
			c.lock.Unlock()
			continue
		}
		if parse.Path == "/admino.ini" {
			c.first[parse.Scheme+parse.Host] = size
		}
		c.lock.Unlock()

		if c.config.Wordlist != "" {
			if size == "0.0B" || size == "-1.0B" {
				continue
			}
			c.lock.Lock()
			_, ok := c.first[parse.Scheme+parse.Host]
			if ok {
				if c.first[parse.Scheme+parse.Host] == size {
					c.lock.Unlock()
					log.Trace("rm", url)
					continue
				}
			}
			c.lock.Unlock()
			flag := true
			for i := 0; i < len(c.excludestatus); i++ {
				if status == c.excludestatus[i] {
					flag = false
				}
			}
			if flag && status != 0 && len(c.url) > 1 {
				pr := make([]interface{}, 0)
				pr = append(pr, BrightWhite(url))
				if status >= 200 && status < 300 {
					pr = append(pr, BrightGreen(status).String())
				} else if status >= 300 && status < 400 {
					pr = append(pr, BrightYellow(status).String())
				} else if status >= 400 && status < 500 {
					pr = append(pr, BrightMagenta(status).String())
				} else {
					pr = append(pr, BrightWhite(status).String())
				}
				pr = append(pr, BrightCyan(size).String())
				if ct == "" {
					pr = append(pr, "null")
				} else {
					pr = append(pr, ct)
				}
				fmt.Println(pr...)
			} else if flag && status != 0 {
				pr := make([]interface{}, 0)
				if status >= 200 && status < 300 {
					pr = append(pr, BrightGreen(status).String())
				} else if status >= 300 && status < 400 {
					pr = append(pr, BrightYellow(status).String())
				} else if status >= 400 && status < 500 {
					pr = append(pr, BrightMagenta(status).String())
				} else {
					pr = append(pr, BrightWhite(status).String())
				}
				pr = append(pr, BrightCyan(size).String())
				pr = append(pr, BrightWhite(url))
				if ct == "" {
					pr = append(pr, "null")
				} else {
					pr = append(pr, ct)
				}
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
