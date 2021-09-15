package utils

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

type Result struct {
	Ip         string         `json:"i"` // ip
	Port       string         `json:"p"` // port
	Uri        string         `json:"u"` // uri
	Os         string         `json:"o"` // os
	Host       string         `json:"h"` // host
	Title      string         `json:"t"` // title
	Midware    string         `json:"m"` // midware
	HttpStat   string         `json:"s"` // http_stat
	Language   string         `json:"l"` // language
	Frameworks Frameworks     `json:"f"` // framework
	Protocol   string         `json:"r"` // protocol
	Vulns      Vulns          `json:"v"`
	Open       bool           `json:"-"`
	TcpCon     *net.Conn      `json:"-"`
	Httpresp   *http.Response `json:"-"`
	Error      string         `json:"-"`
	Content    string         `json:"-"`
}

func (result *Result) InfoFilter() {
	if result.IsHttp() {
		result.Title = getTitle(result.Content)
		result.Language = getLanguage(result.Httpresp, result.Content)
		result.Midware = getMidware(result.Httpresp, result.Content)

	} else {
		result.Title = getTitle(result.Content)
	}
	//处理错误信息
	if result.Content != "" {
		result.errHandler()
	}
}

func (result *Result) AddVuln(vuln Vuln) {
	result.Vulns = append(result.Vulns, vuln)
}

func (result *Result) AddFramework(f Framework) {
	result.Frameworks = append(result.Frameworks, f)
}

func (result *Result) NoFramework() bool {
	if len(result.Frameworks) == 0 {
		return true
	}
	return false
}

func (result Result) IsHttp() bool {
	if strings.HasPrefix(result.Protocol, "http") {
		return true
	}
	return false
}

func (result Result) IsHttps() bool {
	if strings.HasPrefix(result.Protocol, "https") {
		return true
	}
	return false
}

//从错误中收集信息
func (result *Result) errHandler() {

	if strings.Contains(result.Error, "wsasend") || strings.Contains(result.Error, "wsarecv") {
		result.HttpStat = "reset"
	} else if result.Error == "EOF" {
		result.HttpStat = "EOF"
	} else if strings.Contains(result.Error, "http: server gave HTTP response to HTTPS client") {
		result.Protocol = "http"
	} else if strings.Contains(result.Error, "first record does not look like a TLS handshake") {
		result.Protocol = "tcp"
	}
}

func (result *Result) GetURL() string {
	return fmt.Sprintf("%s://%s:%s", result.Protocol, result.Ip, result.Port)
}

func (result *Result) GetTarget() string {
	return fmt.Sprintf("%s:%s", result.Ip, result.Port)
}

func (result *Result) AddNTLMInfo(m map[string]string) {
	result.Title = m["MsvAvNbDomainName"] + "/" + m["MsvAvNbComputerName"]
	result.Host = m["MsvAvDnsDomainName"] + "/" + m["MsvAvDnsComputerName"]
	result.AddFramework(Framework{m["Version"], ""})
}

type Vuln struct {
	Id      string                 `json:"vn"`
	Payload map[string]interface{} `json:"vp"`
	Detail  map[string]interface{} `json:"vd"`
}

func (v *Vuln) GetPayload() string {
	return MaptoString(v.Payload)
}

func (v *Vuln) GetDetail() string {
	return MaptoString(v.Detail)
}

func (v *Vuln) ToString() string {
	s := v.Id
	if payload := v.GetPayload(); payload != "" {
		s += fmt.Sprintf(" payloads:%s", payload)
	}
	if detail := v.GetDetail(); detail != "" {
		s += fmt.Sprintf(" payloads:%s", detail)
	}
	return s
}

type Vulns []Vuln

func (vs Vulns) ToString() string {
	var s string
	for _, vuln := range vs {
		s += fmt.Sprintf("[ Find Vuln: %s ] ", vuln.ToString())
	}
	return s
}

type Framework struct {
	Title   string `json:"ft"`
	Version string `json:"fv"`
}

func (f Framework) ToString() string {
	return fmt.Sprintf("%s%s", f.Title, f.Version)
}

type Frameworks []Framework

func (fs Frameworks) ToString() string {
	framework_strs := make([]string, len(fs))
	for i, f := range fs {
		framework_strs[i] = f.ToString()
	}
	return strings.Join(framework_strs, "||")
}

func (fs Frameworks) GetTitles() []string {
	titles := make([]string, len(fs))
	for i, f := range fs {
		titles[i] = f.Title
	}
	return titles
}
