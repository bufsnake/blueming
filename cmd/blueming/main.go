package main

import (
	"flag"
	"fmt"
	"github.com/bufsnake/blueming/config"
	"github.com/bufsnake/blueming/internal/core"
	"github.com/bufsnake/blueming/pkg/log"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	conf := config.Config{}
	flag.IntVar(&conf.Thread, "t", 10, "set thread")
	flag.IntVar(&conf.Timeout, "s", 10, "set timeout")
	flag.StringVar(&conf.Url, "u", "", "set url")
	flag.StringVar(&conf.Urlfile, "f", "", "set url file")
	flag.StringVar(&conf.Loglevel, "l", log.DEBUG, "set log level(trace,debug,info,warn,fatal)")
	flag.StringVar(&conf.Wordlist, "w", "", "set wordlist")
	flag.StringVar(&conf.Index, "i", "", "set wordlist index(ex: test.php)")
	flag.StringVar(&conf.Proxy, "p", "", "set download proxy")
	flag.StringVar(&conf.ExcludeStatus, "es", "404", "dirscan filter status(200,206,301,302,307,400,401,402,403,404,405,406,424,500,501,502,503,504,600,etc.)")
	flag.StringVar(&conf.ResultFile, "rf", "", "parse result file")
	flag.Parse()
	log.SetLevel(conf.Loglevel)
	urls := []string{}
	if conf.Url != "" {
		urls = append(urls, conf.Url)
	} else if conf.Urlfile != "" {
		file, err := ioutil.ReadFile(conf.Urlfile)
		if err != nil {
			log.Warn(err)
		}
		split := strings.Split(string(file), "\n")
		for i := 0; i < len(split); i++ {
			split[i] = strings.Trim(split[i], "\r")
			if len(split[i]) != 0 {
				urls = append(urls, split[i])
			}
		}
	} else if conf.ResultFile != "" {
		file, err := ioutil.ReadFile(conf.ResultFile)
		if err != nil {
			log.Warn(err)
		}
		split := strings.Split(string(file), "\n")
		results := make(map[string][][]string)
		for i := 1; i < len(split); i++ {
			split[i] = strings.Trim(split[i], "\r")
			temp := ""
			for color := 48; color < 56; color++ {
				temp = strings.ReplaceAll(split[i], fmt.Sprintf("\x1b\x5b\x39%c\x6d", color), "")
			}
			temp = strings.ReplaceAll(temp, "\x1b\x5b\x30\x6d", "")
			result := strings.SplitN(temp, " ", 4)
			result2 := strings.SplitN(split[i], " ", 4)
			if len(result) != 4 {
				continue
			}
			if strings.Contains(result[2], "-1.0B") {
				continue
			}
			parse, _ := url.Parse(result[0])
			results[parse.Scheme+"-"+parse.Host] = append(results[parse.Scheme+"-"+parse.Host], []string{result2[0], result2[1], result2[2]})
		}
		for _, k := range results {
			temp_results := make(map[string][][]string)
			for i := 0; i < len(k); i++ {
				if _, ok := temp_results[k[i][2]]; ok {
					if len(temp_results[k[i][2]]) == 1 {
						temp_results[k[i][2]][0] = []string{}
					}
					continue
				}
				temp_results[k[i][2]] = append(temp_results[k[i][2]], k[i])
			}
			for _, val := range temp_results {
				for i := 0; i < len(val); i++ {
					if len(val[i]) != 3 {
						continue
					}
					fmt.Println(val[i][1], fmt.Sprintf("%16s", val[i][2]), val[i][0])
				}
			}
		}
		return
	} else {
		flag.Usage()
		return
	}
	log.Info(len(urls), "个URL,", conf.Thread, "线程,", conf.Timeout, "超时")
	config.LogFileName = "Log-" + time.Now().Format("2006-01-02 15:04:05")
	create, err := os.Create(config.LogFileName)
	if err != nil {
		log.Warn(err)
	}
	err = create.Close()
	if err != nil {
		log.Warn(err)
	}
	newCore := core.NewCore(urls, conf)
	newCore.Core()
}
