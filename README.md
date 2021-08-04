## 简介

> 最近挺喜欢听IU的blueming，所以命名为blueming
> 主要用于获取网站备份文件

## 安装

```bash
go env -w GO111MODULE=on

go get github.com/bufsnake/blueming

cd $GOPATH/pkg/mod/github.com/bufsnake/blueming<TAB>键后进入/cmd/blueming

go build -v
```

## 使用

```bash
Usage of ./blueming:
  -b	filter output data
  -crt string
    	listen cert (default "ca.crt")
  -f string
    	set url file
  -i string
    	set wordlist index(exp: test.php)
  -key string
    	listen key (default "ca.key")
  -l string
    	set log level(trace,debug,info,warn,fatal) (default "debug")
  -listen string
    	listen to scan dir (default "127.0.0.1:9099")
  -p string
    	set proxy, support http proxy(exp: http://localhost:8080)
  -s int
    	set timeout (default 10)
  -t int
    	set thread (default 100)
  -u string
    	set url
  -v int
    	log level
  -w string
    	set wordlist
```

> ./blueming -b 可删除output下的垃圾数据(必须使用)

## TODO

> 基本满足以下要求即可

- [ ] 常见文件泄露扫描 .git .hg .idea .DS_Store ...
- [x] 开启被动扫描模式，配合httpx自动进行目录扫描(二级、三级、四级...)
- [x] 通过URL自动生成文件名
- [x] 根据后缀名将URL定义为对应的文件格式，如zip、tar.gz等
- [x] 自动下载备份文件，并进行重命名
- [x] 能够自定义字典
- [x] 优化内存占用
- [x] filter.sh 移至程序内部
- [x] 目录扫描部分添加 页面相似度比较，每个新产生的都会与前面所有的请求进行比较一次(耗时)
  - 比较时，各网站相互独立，采用协程的方式
- [x] 采用 GET 请求，查看文件过大时的response
  - 文件过大导致的超时 则获取 header，比较历史记录中的length
  - 正常情况，比较body
