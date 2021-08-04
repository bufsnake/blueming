package main

import (
	"flag"
	"fmt"
	"github.com/bufsnake/blueming/config"
	"github.com/bufsnake/blueming/internal/core"
	"github.com/bufsnake/blueming/pkg/log"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	conf := config.Config{}
	flag.IntVar(&conf.Thread, "t", 100, "set thread")
	flag.IntVar(&conf.Timeout, "s", 10, "set timeout")
	flag.StringVar(&conf.Url, "u", "", "set url")
	flag.StringVar(&conf.Urlfile, "f", "", "set url file")
	flag.StringVar(&conf.Loglevel, "l", log.DEBUG, "set log level(trace,debug,info,warn,fatal)")
	flag.StringVar(&conf.Wordlist, "w", "", "set wordlist")
	flag.StringVar(&conf.Index, "i", "", "set wordlist index(exp: test.php)")
	flag.StringVar(&conf.Proxy, "p", "", "set proxy, support http proxy(exp: http://localhost:8080)")
	flag.StringVar(&conf.Listen, "listen", "127.0.0.1:9099", "listen to scan dir")
	flag.StringVar(&conf.URLStrs, "urls", "", "set url file")
	flag.StringVar(&conf.Cert, "crt", "ca.crt", "listen cert")
	flag.StringVar(&conf.Key, "key", "ca.key", "listen key")
	flag.BoolVar(&conf.FilterOutput, "b", false, "filter output data")
	// 暂不考虑
	//flag.StringVar(&conf.ResultFile, "rf", "", "parse result file")
	flag.Parse()
	// 开启多核模式
	runtime.GOMAXPROCS(runtime.NumCPU() * 3 / 4)
	// 关闭 GIN Debug模式
	// 设置工具可打开的文件描述符
	var rLimit syscall.Rlimit
	rLimit.Max = 999999
	rLimit.Cur = 999999
	if runtime.GOOS == "darwin" {
		rLimit.Cur = 10240
	}
	err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_ = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	log.SetLevel(conf.Loglevel)
	if conf.FilterOutput {
		// 获取 output 下的所有文件 不包含文件夹
		allfiles, _ := ioutil.ReadDir("./output")
		for _,f := range allfiles {
			if !f.IsDir() {
				if f.Size() <= 1048576 {
					err = os.Remove("./output/" + f.Name())
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		}

		wait := sync.WaitGroup{}
		files, _ := ioutil.ReadDir("./output")
		fmt.Println("current exist", len(files), "files")
		go func() {
			for {
				fmt.Printf("\r%.2f%%", math.Trunc(((increase/float64(len(files)))*100)*1e2)*1e-2)
				time.Sleep(1 * time.Second / 10)
			}
		}()
		for _, f := range files {
			if !f.IsDir() {
				wait.Add(1)
				go filter(&wait, strings.ReplaceAll("./output/"+f.Name(), " ", ` `), float64(len(files)))
			} else {
				increaseAdd()
				fmt.Printf("\r%.2f%%", math.Trunc(((increase/float64(len(files)))*100)*1e2)*1e-2)
			}
		}
		wait.Wait()
		// function filter { if [[ $(file $1 | grep $1": data") == "" && $(file $1 | grep "image data") == "" && $(file $1 | grep "HTML") == "" && $(file $1 | grep "empty") == "" && $(file $1 | grep "JSON") == "" && $(file $1 | grep "text") == "" ]]; then file $1; else rm -rf $1; fi } && filter logs/data.tar.gz
		os.Exit(1)
	}
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
	} else if conf.Listen != "" {
		if conf.Wordlist == "" {
			log.Fatal("If passive scanning is started, a dictionary must be specified")
		}
		if conf.URLStrs == "" {
			log.Fatal("urls must be specified")
		}
		passive := core.NewPassive(conf)
		err = passive.Start()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		flag.Usage()
		return
	}
	// 判断 output 文件夹是否存在
	if !exists("./output") {
		log.Info("create output file path")
		err := os.Mkdir("./output/", os.ModePerm)
		if err != nil {
			log.Warn("create output file path error", err)
			os.Exit(1)
		}
	}
	// 创建 Log 文件夹
	if !exists("./logs") {
		log.Info("create logs file path")
		err := os.Mkdir("./logs/", os.ModePerm)
		if err != nil {
			log.Warn("create logs file path error", err)
			os.Exit(1)
		}
	}

	log.Info(len(urls), "个URL,", conf.Thread, "线程,", conf.Timeout, "超时")
	config.LogFileName = "./logs/Log-" + time.Now().Format("2006-01-02 15:04:05")
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

func exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

var increase float64 = 0
var inc_l sync.Mutex

func increaseAdd() {
	inc_l.Lock()
	defer inc_l.Unlock()
	increase++
}

func filter(wait *sync.WaitGroup, filename string, totalcount float64) {
	defer wait.Done()
	bin := []string{"-c", "function filter { if [[ $(file $1 | grep $1\": data\") == \"\" && $(file $1 | grep \"image data\") == \"\" && $(file $1 | grep \"HTML\") == \"\" && $(file $1 | grep \"empty\") == \"\" && $(file $1 | grep \"JSON\") == \"\" && $(file $1 | grep \"text\") == \"\" ]]; then file $1; else rm -rf $1; fi } && filter '" + filename + "'"}
	// 其他的shell环境太烦了
	run := exec.Command("/bin/zsh", bin...)
	output, err := run.Output()
	if err != nil {
		log.Fatal(err)
	}
	output, err = simplifiedchinese.GB18030.NewDecoder().Bytes(output)
	if err != nil {
		log.Fatal(err)
	}
	if len(output) != 0 {
		fmt.Print("\r" + string(output))
	}
	increaseAdd()
	fmt.Printf("\r%.2f%%", math.Trunc(((increase/totalcount)*100)*1e2)*1e-2)
}
