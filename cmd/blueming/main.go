package main

import (
	"flag"
	"github.com/bufsnake/blueming/config"
	"github.com/bufsnake/blueming/internal/core"
	"github.com/bufsnake/blueming/pkg/log"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	conf := config.Config{}
	flag.IntVar(&conf.Thread, "t", 10, "set thread")
	flag.IntVar(&conf.Timeout, "s", 10, "set timeout")
	flag.StringVar(&conf.Url, "u", "", "set url")
	flag.StringVar(&conf.Urlfile, "f", "", "set url file")
	flag.StringVar(&conf.Loglevel, "l", log.DEBUG, "set log level(trace,debug,info,warn,fatal)")
	flag.StringVar(&conf.Wordlist, "w", "", "set wordlist")
	flag.StringVar(&conf.Proxy, "p", "", "set download proxy")
	flag.StringVar(&conf.ExcludeStatus, "es", "404", "dirscan filter status(200,206,301,302,401,403,404,405,500,501,502,503,504,600,etc.)")
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
			urls = append(urls, split[i])
		}
	} else {
		flag.Usage()
		return
	}
	log.Info(len(urls), "个URL,", conf.Thread, "线程,", conf.Timeout, "超时")
	create, err := os.Create("output/download_error")
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
