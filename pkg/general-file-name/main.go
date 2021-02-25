package general_file_name

import (
	"net/url"
	"regexp"
	"strings"
)

type general_file_name struct {
	url string
}

func NewGenURL(url string) (*general_file_name, error) {
	return &general_file_name{url: strings.TrimRight(url, "/")}, nil
}

func (g *general_file_name) GetURL() *[]string {
	ret := make([]string, 0)
	prefix := []string{"www", "admin", "wwwroot", "web", "data", "backup", "db", "database", "code", "test", "user", "sql"}
	suffix := []string{".zip", ".rar", ".tar.gz", ".tgz", ".tar.bz2", ".tar", ".jar", ".war", ".7z", ".bak", ".sql"}
	parse, err := url.Parse(g.url)
	if err == nil {
		if strings.Contains(parse.Host, ":") {
			parse.Host = strings.Split(parse.Host, ":")[0]
		}
		parse.Host = strings.TrimLeft(parse.Host, "www.")
		prefix = append(prefix, parse.Host)
		if isdomain(parse.Host) {
			split := strings.Split(parse.Host, ".")
			if len(split) > 2 {
				prefix = append(prefix, split[len(split)-2]+"."+split[len(split)-1])
				prefix = append(prefix, split[len(split)-2])
			}
		}
	}
	for i := 0; i < len(prefix); i++ {
		for j := 0; j < len(suffix); j++ {
			ret = append(ret, g.url+"/"+prefix[i]+suffix[j])
		}
	}
	return &ret
}

func isdomain(str string) bool {
	if matched, _ := regexp.MatchString("\\d{0,3}\\.\\d{0,3}\\.\\d{0,3}\\.\\d{0,3}", str); matched {
		return false
	}
	return true
}
