package core

import (
	"fmt"
	"github.com/antlabs/strsim"
	"github.com/bufsnake/blueming/config"
	file_download "github.com/bufsnake/blueming/pkg/file-download"
	general_file_name "github.com/bufsnake/blueming/pkg/general-file-name"
	http_request "github.com/bufsnake/blueming/pkg/http-request"
	"github.com/bufsnake/blueming/pkg/log"
	. "github.com/logrusorgru/aurora"
	"io/ioutil"
	"net/http"
	url2 "net/url"
	"os"
	"regexp"
	"strings"
	"sync"
)

type core struct {
	config         config.Config
	url            []string
	wordlist       []string
	htmlsimilarity map[string][]string // 耗时又占内存
	hslock         sync.Mutex
}

func NewCore(url []string, config config.Config) core {
	wordlist := make([]string, 0)
	if config.Wordlist != "" {
		file, err := ioutil.ReadFile(config.Wordlist)
		if err != nil {
			log.Fatal(err)
		}
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
	hs := make(map[string][]string)
	return core{htmlsimilarity: hs, url: url, config: config, wordlist: wordlist}
}

// 目录扫描 和 备份文件扫描 分开
func (c *core) Core() {
	if c.config.Wordlist != "" { // 目录扫描
		c.dirscan()
	} else { // 备份文件扫描
		c.backup()
	}
}

func (c *core) dirscan() {
	httpr := sync.WaitGroup{}
	httpc := make(chan string, c.config.Thread)
	for i := 0; i < c.config.Thread; i++ {
		httpr.Add(1)
		go c.httprequest(&httpr, httpc, nil, c.config.Timeout)
	}
	length := general_file_name.InitGeneral(c.wordlist)
	for w := 0; w < length; w++ {
		for i := 0; i < len(c.url); i++ {
			genURL, err := general_file_name.NewGenURL(c.url[i])
			if err != nil {
				log.Warn(err)
				continue
			}
			httpc <- genURL.GetDirURI(w)
		}
	}
	close(httpc)
	httpr.Wait()
}

func (c *core) backup() {
	log.Info("start scan backup")
	httpr := sync.WaitGroup{}
	httpc := make(chan string, c.config.Thread)
	httpd := make(chan config.HTTPStatus, c.config.Thread)
	for i := 0; i < c.config.Thread; i++ {
		httpr.Add(1)
		go c.httprequest(&httpr, httpc, httpd, c.config.Timeout)
	}
	down := sync.WaitGroup{}
	for i := 0; i < c.config.Thread; i++ {
		down.Add(1)
		go c.httpdownload(&down, httpd)
	}
	// 一阶段 扫描 固定死的URI
	length := general_file_name.InitGeneral([]string{})
	for index := 0; index < length; index++ {
		for i := 0; i < len(c.url); i++ {
			genURL, err := general_file_name.NewGenURL(c.url[i])
			if err != nil {
				log.Warn(err)
				continue
			}
			httpc <- genURL.GetBackupURI(index)
		}
	}

	// 扫描生成的URI
	index := 0
	for {
		flag := true
		for i := 0; i < len(c.url); i++ {
			genURL, err := general_file_name.NewGenURL(c.url[i])
			if err != nil {
				log.Warn(err)
				continue
			}
			getURL := genURL.GetBackupExtURI()
			if len(*getURL) <= index {
				continue
			}
			flag = false
			httpc <- (*getURL)[index]
		}
		if flag {
			break
		}
		index++
	}
	// 分析内存占用
	//memStat := new(runtime.MemStats)
	//runtime.ReadMemStats(memStat)
	//fmt.Println(len(to), memStat.Alloc)
	close(httpc)
	httpr.Wait()
	close(httpd)
	down.Wait()
}

func (c *core) httprequest(wait *sync.WaitGroup, httpc chan string, httpd chan config.HTTPStatus, timeout int) {
	defer wait.Done()
	for url := range httpc {
		var (
			status int
			ct     string
			size   string
			body   string
			err    error
		)
		if c.config.Wordlist == "" { // 备份文件扫描
			status, ct, size, body, err = http_request.HTTPRequest(http.MethodHead, url, c.config.Proxy, timeout)
		} else { // 目录扫描
			status, ct, size, body, err = http_request.HTTPRequest(http.MethodGet, url, c.config.Proxy, timeout)
		}
		log.Trace(status, ct, size, body, err, url)
		if err != nil && c.config.Wordlist == "" {
			log.Warn(err)
		}
		if err != nil && strings.Contains(err.Error(), "proxyconnect tcp") {
			log.Warn(err)
		}
		if c.config.Wordlist != "" {
			if status == 404 || status == 0 {
				continue
			}
			// 计算页面相似度 -- 耗时严重 - 默认使用head请求
			// 与当前URL的所有历史记录进行匹配
			// 相似值低于0.75则追加
			parse, err := url2.Parse(url)
			if err != nil {
				log.Warn(err)
				continue
			}
			c.hslock.Lock()
			similarity := false
			for i := 0; i < len(c.htmlsimilarity[parse.Host]); i++ {
				// 2.516400257s
				//compare := strsim.Compare(body, c.htmlsimilarity[parse.Host][i], strsim.DiceCoefficient(1))

				// 2.916750287s
				//compare := strsim.Compare(body, c.htmlsimilarity[parse.Host][i], strsim.Jaro())

				// 1.839468012s
				compare := strsim.Compare(body, c.htmlsimilarity[parse.Host][i], strsim.Hamming())
				if compare >= 0.75 {
					similarity = true
					break
				}
			}

			if similarity { // 相似 退出
				c.hslock.Unlock()
				continue
			}
			c.htmlsimilarity[parse.Host] = append(c.htmlsimilarity[parse.Host], body)
			c.hslock.Unlock()

			pr := make([]interface{}, 0)
			if status >= 200 && status < 300 {
				pr = append(pr, "["+BrightGreen(status).String()+"]")
			} else if status >= 300 && status < 400 {
				pr = append(pr, "["+BrightYellow(status).String()+"]")
			} else if status >= 400 && status < 500 {
				pr = append(pr, "["+BrightMagenta(status).String()+"]")
			} else {
				pr = append(pr, "["+BrightWhite(status).String()+"]")
			}
			pr = append(pr, "["+BrightCyan(size).String()+"]")
			pr = append(pr, "["+BrightWhite(url).String()+"]")
			if ct == "" {
				pr = append(pr, "["+"null"+"]")
			} else {
				pr = append(pr, "["+ct+"]")
			}
			fmt.Println(pr...)
		} else {
			if status != 200 && status != 206 {
				continue
			}
			if size == "0B" || size == "0.0B" {
				continue
			}
			matchString, err := regexp.MatchString("application/[-\\w.]+", ct)
			if err == nil && matchString {
				log.Info(size, ct, url)
				httpd <- config.HTTPStatus{
					URL:         url,
					Size:        size,
					ContentType: ct,
				}
			}
		}
	}
}

func (c *core) httpdownload(wait *sync.WaitGroup, httpd chan config.HTTPStatus) {
	defer wait.Done()
	for url := range httpd {
		err := file_download.DownloadFile(url.URL, c.config.Proxy)
		if err != nil {
			log.Info("file download error", err)
			// 将URL保存到文件
			file, err := os.OpenFile(config.LogFileName, os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				log.Warn(err)
				continue
			}
			_, err = file.WriteString(url.ContentType + " " + url.Size + " " + url.URL + "\n")
			if err != nil {
				file.Close()
				log.Warn(err)
				continue
			}
			file.Close()
		}
	}
}
