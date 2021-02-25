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
  -f string
    	set blueming url file
  -l string
    	set blueming log level(trace debug info warn fatal) (default "debug")
  -s int
    	set blueming timeout (default 10)
  -t int
    	set blueming thread (default 10)
  -u string
    	set blueming url
```

## TODO

> 基本满足以下要求即可

- [x] 通过URL自动生成文件名
- [x] 根据后缀名将URL定义为对应的文件格式，如zip、tar.gz等
- [x] 自动下载备份文件，并进行重命名