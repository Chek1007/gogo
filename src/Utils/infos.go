package Utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Result struct {
	Ip        string
	Port      string
	Stat      string
	TcpCon    *net.Conn
	HttpCon   *http.Client
	Httpresp  *http.Response
	Os        string
	Host      string
	Title     string
	Midware   string
	HttpStat  string
	Language  string
	Framework string
	Vuln      string
	Protocol  string
	Error     string
	Content   string
}

type Finger struct {
	Name        string   `json:"name"`
	Protocol    string   `json:"protocol"`
	SendData    string   `json:"send_data"`
	Level       int      `json:"level"`
	Defaultport string   `json:"default_port"`
	Regexps     []string `json:"regexps"`
}

var Tcpfingers, Httpfingers = getFingers()

func InfoFilter(result *Result) *Result {

	if strings.HasPrefix(result.Protocol, "http") {
		result.Title = GetTitle(result.Content)
		result.Language = GetLanguage(result.Httpresp, result.Content)
		result.Midware = GetMidware(result.Httpresp, result.Content)

	} else {
		result.Title = GetTitle(result.Content)
	}

	return result

}

func GetDetail(result *Result) *Result {
	var ishttp = false
	if strings.HasPrefix(result.Protocol, "http") {
		ishttp = true
	}

	//如果是http协议,则判断cms,如果是tcp则匹配规则库.暂时不考虑udp
	if ishttp {
		getHttpCMS(result)
	} else {
		getTCPFrameWork(result)
	}

	return result
}

func Encode(s string) string {
	s = strings.Replace(s, "\r", "%13", -1)
	s = strings.Replace(s, "\n", "%10", -1)
	return s
}

func Match(regexpstr string, s string) string {
	Reg, err := regexp.Compile(regexpstr)
	if err != nil {
		return ""
	}
	res := Reg.FindStringSubmatch(s)
	if len(res) == 1 {
		return "matched"
	} else if len(res) == 2 {
		return res[1]
	}
	return ""
}

func GetTitle(content string) string {
	title := Match("(?im)<title>(.*)</title>", content)
	if title != "" {
		return title
	}
	return Encode(string([]byte(content)[:13]))
}

func GetMidware(resp *http.Response, content string) string {
	var server string
	if resp == nil {
		server = Match("(?i)Server: ([\x20-\x7e]+)", strings.Split(content, "\r\n\r\n")[0])
	} else {
		server = resp.Header.Get("Server")
	}
	if server != "" {
		return server
	}

	return ""

}

func GetLanguage(resp *http.Response, content string) string {
	// update: 减少正则匹配的速度略微提升性能
	var powered string
	if resp == nil {
		powered = Match("(?i)X-Powered-By: ([\x20-\x7e]+)", strings.Split(content, "\r\n\r\n")[0])
		if powered != "" {
			return powered
		}

		sessionid := Match("(?i) (.*SESS.*?ID)", "")

		if sessionid != "" {
			switch sessionid {
			case "JSESSIONID":
				return "JAVA"
			case "ASP.NET_SessionId":
				return "ASP.NET"
			case "PHPSESSID":
				return "PHP"
			}
		}

	} else {
		powered = resp.Header.Get("X-Powered-By")
		if powered != "" {
			return powered
		}

		cookies := getCookies(resp)
		if cookies["JSESSIONID"] != "" {
			return "JAVA"
		} else if cookies["ASP.NET_SessionId"] != "" {
			return "ASP"
		} else if cookies["PHPSESSID"] != "" {
			return "PHP"
		} else {
			return ""
		}
	}

	return ""
}

func getHttpCMS(result *Result) *Result {
	//for _,finger :=range Httpfingers{
	//
	//}
	return result
}

//第一个返回值为详细的版本信息,第二个返回值为规则名字
func getTCPFrameWork(result *Result) *Result {
	// 第一遍循环只匹配默认端口
	for _, finger := range Tcpfingers {
		if finger.Defaultport == result.Port {
			tcpFingerMatch(result, finger)
		}
		if result.Framework != "" {
			return result
		}
	}

	// 若默认端口未匹配到结果,这匹配全部
	for _, finger := range Tcpfingers {
		if finger.Defaultport != result.Port {
			tcpFingerMatch(result, finger)
		}

		if result.Framework != "" {
			return result
		}
	}

	return result
}

func tcpFingerMatch(result *Result, finger Finger) *Result {
	content := result.Content
	var data []byte
	var err error

	// 某些规则需要主动发送一个数据包探测
	if finger.SendData != "" {
		// 复用tcp链接
		_, data, err = SocketSend(*result.TcpCon, []byte(finger.SendData), 1024)

		// 如果报错为EOF,则需要重新建立tcp连接
		if err.Error() == "EOF" {
			target := GetTarget(result)
			// 如果对端已经关闭,则本地socket也关闭

			(*result.TcpCon).Close()

			*result.TcpCon, _ = TcpSocketConn(target, 2)
			_, data, err = SocketSend(*result.TcpCon, []byte(finger.SendData), 1024)

			// 重新建立链接后再次报错,则跳过该规则匹配
			if err != nil {
				result.Error = err.Error()
				return result
			}
		}
	}
	// 如果主动探测有回包,则正则匹配回包内容
	if string(data) != "" {
		content = string(data)
	}

	//遍历正则
	for _, regexpstr := range finger.Regexps {
		res := Match("(?im)"+regexpstr, content)
		if res == "matched" {
			//println("[*] " + res)
			result.Framework = finger.Name
		} else if res != "" {
			result.HttpStat = "tcp"
			result.Framework = finger.Name
			result.Title = res
		}
	}

	return result
}

func GetHttpRaw(resp *http.Response) string {
	var raw string

	raw += fmt.Sprintf("%s %s\r\n", resp.Proto, resp.Status)
	for k, v := range resp.Header {
		for _, i := range v {
			raw += fmt.Sprintf("%s: %s\r\n", k, i)
		}
	}
	raw += "\r\n"
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return raw
	}
	raw += string(body)
	return raw
}

func GetBody(resp *http.Response) string {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(body)
}

func getCookies(resp *http.Response) map[string]string {
	cookies := make(map[string]string)
	for _, cookie := range resp.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}
	return cookies
}
func GetStatusCode(content string) string {
	if strings.Contains(content, "HTTP") {
		return content[9:12]
	}

	return "tcp"
}

func FilterCertDomain(domins []string) string {
	var res string
	if len(domins) == 0 {
		return ""
	} else if len(domins) == 1 {
		return domins[0]
	}
	for _, domain := range domins {
		if !strings.Contains(domain, "www.") {
			res += domain + ","
		}
	}
	return res[:len(res)-1]
}

func getFingers() ([]Finger, []Finger) {
	fingersJson := loadFingers()

	var tcpfingers, httpfingers, fingers []Finger
	// 根据权重排序在python脚本中已经实现
	err := json.Unmarshal([]byte(fingersJson), &fingers)

	if err != nil {
		println("[-] fingers load FAIL!")
		os.Exit(0)
	}

	// tcp与http规则适用不同的扫描逻辑
	for _, finger := range fingers {
		if finger.Protocol == "tcp" {
			tcpfingers = append(tcpfingers, finger)
		} else if finger.Protocol == "http" {
			httpfingers = append(httpfingers, finger)
		}
	}

	return tcpfingers, httpfingers
}
