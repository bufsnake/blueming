package config

type Config struct {
	Thread        int
	Timeout       int
	Url           string
	Urlfile       string
	Loglevel      string
	Wordlist      string
	Proxy         string
	ExcludeStatus string
}
