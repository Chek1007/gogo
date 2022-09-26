package pkg

import (
	"fmt"
	"github.com/chainreactors/gogo/v2/pkg/fingers"
	"github.com/chainreactors/gogo/v2/pkg/utils"
	"github.com/chainreactors/parsers"
	"net"
	"net/http"
	"strings"
	"time"
)

type Result struct {
	// baseinfo
	Ip       string `json:"ip"`             // ip
	Port     string `json:"port"`           // port
	Protocol string `json:"protocol"`       // protocol
	Status   string `json:"status"`         // http_stat
	Uri      string `json:"uri,omitempty"`  // uri
	Os       string `json:"os,omitempty"`   // os
	Host     string `json:"host,omitempty"` // host

	//Cert         string         `json:"c"`
	HttpHosts   []string `json:"-"`
	CurrentHost string   `json:"-"`
	Title       string   `json:"title"`   // title
	Midware     string   `json:"midware"` // midware

	Language     string         `json:"language"`             // language
	Frameworks   Frameworks     `json:"frameworks,omitempty"` // framework
	Vulns        Vulns          `json:"vulns,omitempty"`
	Extracts     *Extracts      `json:"-"`
	ExtractsStat map[string]int `json:"extracts_stat,omitempty"`
	//Hash         string         `json:"hs"`
	Open bool `json:"-"`
	//FrameworksMap map[string]bool `json:"-"`
	SmartProbe bool              `json:"-"`
	TcpConn    *net.Conn         `json:"-"`
	HttpConn   *http.Client      `json:"-"`
	Httpresp   *parsers.Response `json:"-"`
	Error      string            `json:"-"`
	ErrStat    int               `json:"-"`
	Content    string            `json:"-"`
}

func NewResult(ip, port string) *Result {
	var result = Result{
		Ip:           ip,
		Port:         port,
		Protocol:     "tcp",
		Status:       "tcp",
		Extracts:     &Extracts{},
		ExtractsStat: map[string]int{},
	}
	result.Extracts.Target = result.GetTarget()
	return &result
}

func (result *Result) GetHttpConn(delay int) *http.Client {
	if result.HttpConn == nil {
		result.HttpConn = HttpConn(delay)
	} else {
		result.HttpConn.Timeout = time.Duration(delay) * time.Second
	}
	return result.HttpConn
}

func (result *Result) AddVuln(vuln *fingers.Vuln) {
	if vuln.Severity == "" {
		vuln.Severity = "high"
	}
	result.Vulns = append(result.Vulns, vuln)
}

func (result *Result) AddVulns(vulns []*fingers.Vuln) {
	for _, v := range vulns {
		result.AddVuln(v)
	}
}

func (result *Result) AddFramework(f *fingers.Framework) {
	result.Frameworks = append(result.Frameworks, f)
	//result.FrameworksMap[f.ToString()] = true
}

func (result *Result) AddFrameworks(f []*fingers.Framework) {
	result.Frameworks = append(result.Frameworks, f...)
	//for _, framework := range f {
	//result.FrameworksMap[framework.ToString()] = true
	//}
}

func (result *Result) AddExtract(extract *fingers.Extracted) {
	result.Extracts.Extractors = append(result.Extracts.Extractors, extract)
	result.ExtractsStat[extract.Name] = len(extract.ExtractResult)
}

func (result *Result) AddExtracts(extracts []*fingers.Extracted) {
	for _, extract := range extracts {
		result.Extracts.Extractors = append(result.Extracts.Extractors, extract)
		result.ExtractsStat[extract.Name] = len(extract.ExtractResult)
	}
}

func (result *Result) GetExtractStat() string {
	if len(result.ExtractsStat) > 0 {
		var s []string
		for name, length := range result.ExtractsStat {
			s = append(s, fmt.Sprintf("%s:%ditems", name, length))
		}
		return fmt.Sprintf("[ extracts: %s ]", strings.Join(s, ", "))
	} else {
		return ""
	}
}

func (result Result) NoFramework() bool {
	if len(result.Frameworks) == 0 {
		return true
	}
	return false
}

func (result *Result) GuessFramework() {
	for _, v := range PortMap.Get(result.Port) {
		if TagMap.Get(v) == nil && !utils.SliceContains([]string{"top1", "top2", "top3", "other", "windows"}, v) {
			result.AddFramework(&fingers.Framework{Name: v, From: fingers.GUESS})
		}
	}
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

func (result *Result) Get(key string) string {
	switch key {
	case "ip":
		return result.Ip
	case "port":
		return result.Port
	case "frameworks", "framework", "frame":
		return result.Frameworks.ToString()
	case "vulns", "vuln":
		return result.Vulns.ToString()
	case "host":
		return result.Host
	case "title":
		return result.Title
	case "target":
		return result.GetTarget()
	case "url":
		return result.GetBaseURL()
	case "midware":
		return result.Midware
	//case "hash":
	//	return result.Hash
	case "language":
		return result.Language
	case "protocol":
		return result.Protocol
	default:
		return ""
	}
}

//从错误中收集信息
func (result *Result) errHandler() {
	if result.Error == "" {
		return
	}
	if strings.Contains(result.Error, "wsasend") || strings.Contains(result.Error, "wsarecv") {
		result.Status = "reset"
	} else if result.Error == "EOF" {
		result.Status = "EOF"
	} else if strings.Contains(result.Error, "http: server gave HTTP response to HTTPS client") {
		result.Protocol = "http"
	} else if strings.Contains(result.Error, "first record does not look like a TLS handshake") {
		result.Protocol = "tcp"
	}
}

func (result Result) GetBaseURL() string {
	return fmt.Sprintf("%s://%s:%s", result.Protocol, result.Ip, result.Port)
}

func (result Result) GetHostBaseURL() string {
	if result.CurrentHost == "" {
		return result.GetBaseURL()
	} else {
		return fmt.Sprintf("%s://%s:%s", result.Protocol, result.CurrentHost, result.Port)
	}
}

func (result Result) GetURL() string {
	if result.IsHttp() {
		return result.GetBaseURL() + result.Uri
	} else {
		return result.GetBaseURL()
	}
}

func (result Result) GetHostURL() string {
	return result.GetHostBaseURL() + result.Uri
}

func (result Result) GetTarget() string {
	return fmt.Sprintf("%s:%s", result.Ip, result.Port)
}

func (result Result) GetFirstFramework() string {
	if !result.NoFramework() {
		return result.Frameworks[0].Name
	}
	return ""
}

func (result *Result) AddNTLMInfo(m map[string]string, t string) {
	result.Title = m["MsvAvNbDomainName"] + "/" + m["MsvAvNbComputerName"]
	result.Host = strings.Trim(m["MsvAvDnsDomainName"], "\x00") + "/" + m["MsvAvDnsComputerName"]
	result.AddFramework(&fingers.Framework{Name: t, Version: m["Version"]})
}

func (result *Result) Filter(k, v, op string) bool {
	var matchfunc func(string, string) bool
	if op == "::" {
		matchfunc = strings.Contains
	} else {
		matchfunc = strings.EqualFold
	}

	if matchfunc(strings.ToLower(result.Get(k)), v) {
		return true
	}
	return false
}

type zombiemeta struct {
	IP      string `json:"ip"`
	Port    string `json:"port"`
	Service string `json:"service"`
}

type Results []*Result

func (rs Results) Filter(k, v, op string) Results {
	var filtedres Results

	for _, result := range rs {
		if result.Filter(k, v, op) {
			filtedres = append(filtedres, result)
		}
	}
	return filtedres
}

func (results Results) GetValues(key string) []string {
	values := make([]string, len(results))
	for i, result := range results {
		//if focus && !result.Frameworks.IsFocus() {
		//	// 如果需要focus, 则跳过非focus标记的framework
		//	continue
		//}
		values[i] = result.Get(key)
	}
	return values
}
