package config

type Config struct {
	Thread       int
	Timeout      int
	Url          string
	Urlfile      string
	Loglevel     string
	Wordlist     string
	Index        string
	Proxy        string
	FilterOutput bool // 过滤 output 文件夹中的垃圾数据
	Listen       string
	Cert         string
	Key          string
}

var LogFileName string

type HTTPStatus struct {
	URL         string
	Status      int
	ContentType string
	Size        string
	Body        string
}
