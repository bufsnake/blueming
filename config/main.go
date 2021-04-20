package config

type Config struct {
	Thread        int
	Timeout       int
	Url           string
	Urlfile       string
	Loglevel      string
	Wordlist      string
	Index         string
	Proxy         string
	ExcludeStatus string
	ResultFile    string
}

var LogFileName string
