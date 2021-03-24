## 简介

> 最近挺喜欢听IU的blueming，所以命名为blueming
> 主要用于获取网站备份文件

## 安装

```bash
go get github.com/bufsnake/blueming

cd $GOPATH/src/github.com/bufsnake/blueming/cmd/blueming/

go build -v
```

## 使用

```bash
└> ./blueming
Usage of ./blueming:
  -es string
    	dirscan filter status(200,206,301,302,307,400,401,402,403,404,405,406,424,500,501,502,503,504,600,etc.) (default "404")
  -f string
    	set url file
  -i string
    	set wordlist index(ex: test.php)
  -l string
    	set log level(trace,debug,info,warn,fatal) (default "debug")
  -p string
    	set download proxy
  -rf string
    	parse result file
  -s int
    	set timeout (default 10)
  -t int
    	set thread (default 10)
  -u string
    	set url
  -w string
    	set wordlist
```

## TODO

> 基本满足以下要求即可

- [x] 通过URL自动生成文件名
- [x] 根据后缀名将URL定义为对应的文件格式，如zip、tar.gz等
- [x] 自动下载备份文件，并进行重命名
- [x] 能够自定义字典
- [x] 优化内存占用
