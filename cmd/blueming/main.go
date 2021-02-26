package main

import (
	"flag"
	"github.com/bufsnake/blueming/internal/core"
	"github.com/bufsnake/blueming/pkg/log"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	err := os.Remove("output/download_error")
	if err != nil {
		log.Warn(err)
	}
	thread := flag.Int("t", 10, "set blueming thread")
	timeout := flag.Int("s", 10, "set blueming timeout")
	url := flag.String("u", "", "set blueming url")
	urlfile := flag.String("f", "", "set blueming url file")
	loglevel := flag.String("l", log.DEBUG, "set blueming log level(trace debug info warn fatal)")
	proxy := flag.String("p", "", "set blueming download proxy")
	flag.Parse()
	log.SetLevel(*loglevel)
	urls := []string{}
	if *url != "" {
		urls = append(urls, *url)
	} else if *urlfile != "" {
		file, err := ioutil.ReadFile(*urlfile)
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
	log.Info(len(urls), "个URL,", *thread, "线程,", *timeout, "超时")
	create, err := os.Create("output/download_error")
	if err != nil {
		log.Warn(err)
	}
	err = create.Close()
	if err != nil {
		log.Warn(err)
	}
	core.NewCore(urls, *thread, *timeout, *proxy)
}
