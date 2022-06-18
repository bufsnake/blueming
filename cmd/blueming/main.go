package main

import (
	"fmt"
	"github.com/bufsnake/blueming/config"
	"github.com/bufsnake/blueming/internal/core"
	"github.com/bufsnake/blueming/pkg/log"
	"github.com/spf13/cobra"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

func init() {
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
}

func main() {
	conf := config.Config{}
	var blueming = &cobra.Command{
		Use: "blueming",
	}
	blueming.CompletionOptions = cobra.CompletionOptions{
		DisableDefaultCmd:   true,
		DisableNoDescFlag:   true,
		DisableDescriptions: true,
		HiddenDefaultCmd:    true,
	}
	var backupscan = &cobra.Command{
		Use:   "backupscan",
		Short: "backupscan scan",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if (conf.Urlfile == "" && conf.Url == "") && !conf.FilterOutput {
				cmd.Usage()
				log.Fatal("please specify a goal")
			}
			if conf.FilterOutput {
				allfiles, _ := ioutil.ReadDir("./output")
				for _, f := range allfiles {
					if !f.IsDir() {
						if f.Size() <= 1048576 {
							err := os.Remove("./output/" + f.Name())
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
				os.Exit(1)
			}
		},
	}
	backupscan.Flags().StringVarP(&conf.Url, "url", "u", "", "scan a website")
	backupscan.Flags().StringVarP(&conf.Urlfile, "url-list", "l", "", "scan multiple websites")
	backupscan.Flags().StringVarP(&conf.Proxy, "proxy", "p", "", "set up proxy(http://localhost:8080)")
	backupscan.Flags().IntVarP(&conf.Thread, "thread", "t", 100, "set up thread")
	backupscan.Flags().IntVarP(&conf.Timeout, "timeout", "s", 10, "set up timeout")
	backupscan.Flags().StringVarP(&conf.Loglevel, "log-level", "v", log.DEBUG, "set up log level(trace,debug,info,warn,fatal)")
	backupscan.Flags().BoolVarP(&conf.FilterOutput, "filter-output", "f", false, "empty junk data, must be used for backup scan")
	
	var dirscan = &cobra.Command{
		Use:   "dirscan",
		Short: "dirscan scan",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if (conf.Urlfile == "" && conf.Url == "") || conf.Wordlist == "" {
				cmd.Usage()
				log.Fatal("please specify target and dictionary")
			}
		},
	}
	dirscan.Flags().StringVarP(&conf.Url, "url", "u", "", "scan a website")
	dirscan.Flags().StringVarP(&conf.Urlfile, "url-list", "l", "", "scan multiple websites")
	dirscan.Flags().StringVarP(&conf.Proxy, "proxy", "p", "", "set up proxy(http://localhost:8080)")
	dirscan.Flags().IntVarP(&conf.Thread, "thread", "t", 100, "set up thread")
	dirscan.Flags().IntVarP(&conf.Timeout, "timeout", "s", 10, "set up timeout")
	dirscan.Flags().StringVarP(&conf.Wordlist, "wordlist", "w", "", "set up wordlist")
	dirscan.Flags().StringVarP(&conf.Index, "index", "i", "", "set the starting position of the dictionary(-i /test.php)")
	
	var passive = &cobra.Command{
		Use:   "passive",
		Short: "passive scan",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if conf.Urlfile == "" || conf.Wordlist == "" {
				cmd.Usage()
				log.Fatal("please specify target and dictionary")
			}
			passives := core.NewPassive(conf)
			err := passives.Start()
			if err != nil {
				log.Fatal(err)
			}
			os.Exit(1)
		},
	}
	passive.Flags().StringVarP(&conf.Urlfile, "url-list", "l", "", "scan multiple websites")
	passive.Flags().StringVarP(&conf.Listen, "listen", "i", "127.0.0.1:8091", "set listening address")
	passive.Flags().IntVarP(&conf.Timeout, "timeout", "s", 10, "set up timeout")
	passive.Flags().IntVarP(&conf.Thread, "thread", "t", 100, "set up thread")
	passive.Flags().StringVarP(&conf.Wordlist, "wordlist", "w", "", "set up wordlist")
	passive.Flags().StringVarP(&conf.Cert, "cert", "c", "ca.crt", "set up the certificate")
	passive.Flags().StringVarP(&conf.Key, "key", "k", "ca.key", "set up the key")
	blueming.AddCommand(backupscan, dirscan, passive)
	err := blueming.Execute()
	if err != nil {
		log.Fatal(err)
	}
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
	}
	if len(urls) == 0 {
		return
	}
	
	// 判断 output 文件夹是否存在
	if !exists("./output") {
		log.Info("create output file path")
		err = os.Mkdir("./output/", os.ModePerm)
		if err != nil {
			log.Warn("create output file path error", err)
			os.Exit(1)
		}
	}
	
	// 创建 Log 文件夹
	if !exists("./logs") {
		log.Info("create logs file path")
		err = os.Mkdir("./logs/", os.ModePerm)
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
