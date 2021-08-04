package general_file_name

import (
	"github.com/bufsnake/blueming/pkg/log"
	"github.com/weppos/publicsuffix-go/publicsuffix"
	"net/url"
	"regexp"
	"strings"
)

type general_file_name struct {
	url       string
	backupuri []string
}

var ret = []string{}
var wordlist = []string{}

func InitGeneral(wordlists []string) int {
	prefix := []string{"index", "site", "db", "archive", "auth", "website", "backup", "test", "sql", "2016", "com", "dump", "master", "sales", "1", "2013", "members", "wwwroot", "clients", "back", "php", "localhost", "local", "127.0.0.1", "2019", "joomla", "wp", "html", "home", "tar", "vb", "database", "2012", "2020", "engine", "error_log", "mysql", "2018", "my", "new", "wordpress", "user", "2015", "customers", "dat", "media", "2014", "users", "2011", "2021", "old", "code", "jsp", "js", "store", "www", "2017", "web", "orders", "admin", "forum", "aspx", "data", "2010", "backups", "files", "bin"}
	suffix := []string{".zip", ".rar", ".tar.gz", ".tgz", ".tar.bz2", ".tar", ".jar", ".war", ".7z", ".bak", ".sql"}

	for i := 0; i < len(prefix); i++ {
		for j := 0; j < len(suffix); j++ {
			ret = append(ret, "/"+prefix[i]+suffix[j])
		}
	}
	wordlist = wordlists
	if len(wordlist) != 0 {
		return len(wordlist)
	}
	return len(ret)
}

func NewGenURL(url string) (*general_file_name, error) {
	return &general_file_name{backupuri: ret, url: strings.TrimRight(url, "/")}, nil
}

func (g *general_file_name) GetDirURI(index int) string {
	return g.url + "/" + strings.TrimLeft(wordlist[index], "/")
}

func (g *general_file_name) GetBackupURI(index int) string {
	return g.url + g.backupuri[index]
}

func (g *general_file_name) GetBackupExtURI() *[]string {
	rets := make([]string, 0)
	// *** 属于拓展 URI 每个URL不同 单独进行获取
	prefix := []string{}
	suffix := []string{".zip", ".rar", ".tar.gz", ".tgz", ".tar.bz2", ".tar", ".jar", ".war", ".7z", ".bak", ".sql"}
	// 去掉域名 www. 前缀 添加到prefix
	// 获取域名根域，添加到prefix
	// 获取IP，添加到prefix

	// 获取host
	parse, err := url.Parse(g.url)
	if err != nil {
		log.Warn(err)
		return nil
	}
	parse.Host = strings.TrimLeft(parse.Host, "www")
	parse.Host = strings.TrimLeft(parse.Host, ".")
	if strings.Contains(parse.Host, ":") {
		parse.Host = strings.Split(parse.Host, ":")[0]
	}
	prefix = append(prefix, parse.Host)
	if isdomain(g.url) { // 域名 - 获取根域
		domain, err := publicsuffix.Domain(parse.Host)
		if err != nil {
			log.Warn(err)
			return nil
		}
		exist := false
		for _, vv := range prefix {
			if vv == domain {
				exist = true
			}
		}
		if !exist {
			prefix = append(prefix, domain)
		}
	}
	for i := 0; i < len(prefix); i++ {
		for j := 0; j < len(suffix); j++ {
			rets = append(rets, g.url+"/"+prefix[i]+suffix[j])
		}
	}
	return &rets
}

func isdomain(str string) bool {
	if matched, _ := regexp.MatchString("\\d{0,3}\\.\\d{0,3}\\.\\d{0,3}\\.\\d{0,3}", str); matched {
		return false
	}
	return true
}
