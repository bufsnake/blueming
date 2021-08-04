package core

import (
	"net/url"
	"strings"
)

// 解析发送的URL,获取二级、三级、四级目录
func (c *passive) getUrlLayerDirectory(urlstr string) (ret []string, err error) {
	parse, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}
	url_ := parse.Scheme + "://" + parse.Host + "/"
	path := strings.Split(parse.Path, "/")
	for i := 0; i < len(path); i++ {
		if path[i] == "" {
			continue
		}
		if strings.Contains(path[i], ".") && i == len(path)-1 {
			continue
		}
		url_c := concat(path, url_, i)
		if strings.Trim(url_c, "/") == strings.Trim(url_, "/") || strings.Trim(url_c, "/") == strings.Trim(urlstr, "/") {
			continue
		}
		ret = append(ret, url_c)
	}
	if !strings.Contains(urlstr, "?") && !strings.Contains(urlstr, ".") {
		if !strings.HasSuffix(urlstr, "/") {
			urlstr = urlstr + "/"
		}
	}
	ret = append(ret, urlstr)
	ret = append(ret, url_)
	return ret, nil
}

func concat(path []string, urlstr string, index int) string {
	urlstr = strings.Trim(urlstr, "/")
	for i := 0; i <= index; i++ {
		urlstr += strings.Trim(path[i], "/") + "/"
	}
	return urlstr
}

// 解析获取HOST，不包含端口/包含端口
// RET: HOST,PATH
func (c *passive) getUrlHostAndPath(urlstr string, contain_port bool) (host string, path string, err error) {
	parse, err := url.Parse(urlstr)
	if err != nil {
		return "", "", err
	}
	if parse.RawQuery != "" {
		parse.Path = parse.Path + "?" + parse.RawQuery
	}
	if !contain_port && strings.Contains(parse.Host, ":") {
		split := strings.Split(parse.Host, ":")
		if len(split) == 2 {
			return split[0], parse.Path, nil
		}
	}
	return parse.Host, parse.Path, nil
}
