package core

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/bufsnake/blueming/config"
	"github.com/bufsnake/blueming/pkg/log"
	"github.com/bufsnake/blueming/pkg/parseip"
	"github.com/google/martian"
	log2 "github.com/google/martian/log"
	"github.com/google/martian/mitm"
	"github.com/weppos/publicsuffix-go/publicsuffix"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"
)

type passive struct {
	conf config.Config
}

func NewPassive(conf_ config.Config) passive {
	return passive{conf: conf_}
}

// REF: https://github.com/google/martian/blob/master/cmd/proxy/main.go
func (c *passive) Start() error {
	file, err := ioutil.ReadFile(c.conf.URLStrs)
	if err != nil {
		return err
	}
	urls := make([]string, 0)
	split := strings.Split(string(file), "\n")
	for i := 0; i < len(split); i++ {
		trim := strings.Trim(split[i], "\r")
		if trim == "" {
			continue
		}
		if strings.HasPrefix(trim, "http://") || strings.HasPrefix(trim, "https://") {
			parse, err := url.Parse(trim)
			if err != nil {
				log.Warn(err)
				continue
			}
			trim = parse.Host
		}
		if strings.Contains(trim, ":") {
			trim = strings.Split(trim, ":")[0]
		}
		// IP 解析
		_, _, err = parseip.ParseIP(trim)
		if err != nil {
			domain, err := publicsuffix.Domain(trim)
			if err != nil {
				log.Warn(err)
				continue
			}
			urls = append(urls, domain)
			continue
		}
		urls = append(urls, trim)
	}
	log.Info("number of assets that can be scanned", len(urls))
	// 启动代理 // 建立请求队列 // 筛选过滤HTTP链接 // 进行漏洞扫描
	log2.SetLevel(-1)
	martian.Init()
	p := martian.NewProxy()
	defer p.Close()
	l, err := net.Listen("tcp", c.conf.Listen)
	if err != nil {
		log.Fatal(err)
	}
	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 60 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   60 * time.Second,
		ExpectContinueTimeout: time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	p.SetRoundTripper(tr)

	var x509c *x509.Certificate
	var priv interface{}
	var raw []byte
	_, err = ioutil.ReadFile(c.conf.Cert)
	if err != nil && strings.HasSuffix(err.Error(), "no such file or directory") {
		x509c, priv, raw, err = NewAuthority("blueming", "localhost", 365*24*time.Hour)
		if err != nil {
			log.Fatal(err)
		}
		certOut, _ := os.Create("./ca.crt")
		err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: raw})
		err = certOut.Close()

		priv_ := &rsa.PrivateKey{}
		switch priv.(type) {
		case *rsa.PrivateKey:
			priv_ = priv.(*rsa.PrivateKey)
		}
		keyOut, _ := os.Create("./ca.key")
		err = pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv_)})
		err = keyOut.Close()
		log.Info("generate certificate success in current dir")
	} else if c.conf.Cert != "" && c.conf.Key != "" {
		tlsc, err := tls.LoadX509KeyPair(c.conf.Cert, c.conf.Key)
		if err != nil {
			log.Fatal(err)
		}
		priv = tlsc.PrivateKey

		x509c, err = x509.ParseCertificate(tlsc.Certificate[0])
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Info("loading cert from", c.conf.Cert, "and", c.conf.Key)
	log.Info(fmt.Sprintf("martian: starting proxy on %s", l.Addr().String()))

	var mitm_config *mitm.Config
	mitm_config, err = mitm.NewConfig(x509c, priv)
	if err != nil {
		log.Fatal(err)
	}
	mitm_config.SkipTLSVerify(false)
	p.SetMITM(mitm_config)

	// 设置下级代理
	if c.conf.Proxy != "" {
		var url_proxy *url.URL
		url_proxy, err = url.Parse(c.conf.Proxy)
		if err != nil {
			log.Fatal(err)
		}
		p.SetDownstreamProxy(url_proxy)
	}

	urlstrs := make(chan string, 1000000)
	m := modify{urlstrs: urlstrs}
	p.SetRequestModifier(&m)

	go func() {
		err = p.Serve(l)
		if err != nil {
			log.Fatal(err)
		}
	}()
	urlstr_ := make(map[string]bool)
	urlstrc := make(chan string, 1000)
	for i := 0; i < c.conf.Thread; i++ {
		go func() {
			// 获取链接并进行扫描
			for urlstr := range urlstrc {
				flag := false
				for _, host := range urls {
					if strings.Contains(urlstr, host) {
						flag = true
						break
					}
				}
				if !flag {
					continue
				}
				newCore := NewCore([]string{urlstr}, c.conf)
				newCore.Core()
			}
		}()
	}

	for urlstr := range urlstrs {
		urlstr = strings.ReplaceAll(urlstr, " ", "")
		urlstrs_, err := c.getUrlLayerDirectory(urlstr)
		if err != nil {
			log.Warn(err)
			continue
		}
		for i := 0; i < len(urlstrs_); i++ {
			temp := strings.Trim(urlstrs_[i], "/") + "/"
			if _, ok := urlstr_[temp]; !ok {
				urlstr_[temp] = true
				urlstrc <- temp
			}
		}
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	<-sigc
	log.Fatal("martian: shutting down")
	return nil
}

type modify struct {
	urlstrs chan string
}

func (v *modify) ModifyRequest(req *http.Request) error {
	if strings.ToUpper(req.Method) == "OPTIONS" || strings.ToUpper(req.Method) == "CONNECT" {
		return nil
	}
	v.urlstrs <- req.URL.String()
	return nil
}

func NewAuthority(name, organization string, validity time.Duration) (x509c *x509.Certificate, priv *rsa.PrivateKey, raw []byte, err error) {
	priv, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, nil, err
	}
	pub := priv.Public()

	// Subject Key Identifier support for end entity certificate.
	// https://www.ietf.org/rfc/rfc3280.txt (section 4.2.1.2)
	pkixpub, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, nil, nil, err
	}
	h := sha1.New()
	h.Write(pkixpub)
	keyID := h.Sum(nil)

	// TODO: keep a map of used serial numbers to avoid potentially reusing a
	// serial multiple times.
	serial, err := rand.Int(rand.Reader, big.NewInt(0).SetBytes(bytes.Repeat([]byte{255}, 20)))
	if err != nil {
		return nil, nil, nil, err
	}

	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   name,
			Organization: []string{organization},
		},
		SubjectKeyId:          keyID,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		NotBefore:             time.Now().Add(-validity),
		NotAfter:              time.Now().Add(validity),
		DNSNames:              []string{name},
		IsCA:                  true,
	}

	raw, err = x509.CreateCertificate(rand.Reader, tmpl, tmpl, pub, priv)
	if err != nil {
		return nil, nil, nil, err
	}

	x509c, err = x509.ParseCertificate(raw)
	if err != nil {
		return nil, nil, nil, err
	}

	return x509c, priv, raw, nil
}
