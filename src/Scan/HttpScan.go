package Scan

import (
	"getitle/src/Utils"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//socket进行对网站的连接
func SocketHttp(target string, result Utils.Result) Utils.Result {
	//fmt.Println(ip)
	//socket tcp连接,超时时间
	result.Protocol = "tcp"
	conn, err := Utils.TcpSocketConn(target, Delay)
	if err != nil {
		//fmt.Println(err)
		result.Error = err.Error()
		return result
	}
	result.Stat = "OPEN"

	//发送内容
	senddata := []byte("GET / HTTP/1.1\r\nHost: " + target + "\r\n\r\n")
	_, data, _ := Utils.SocketSend(conn, senddata, 4096)
	content := string(data)
	err = conn.Close()
	if err != nil {
		result.Error = err.Error()
		return result
	}

	//获取状态码
	result.Content = content
	result.HttpStat = Utils.GetStatusCode(content)
	if result.HttpStat != "tcp" {
		result.Protocol = "http"
	}
	//所有30x,400,以及非http协议的开放端口都送到http包尝试获取更多信息
	if result.HttpStat == "400" || result.HttpStat == "tcp" || strings.HasPrefix(result.HttpStat, "3") {
		return SystemHttp(target, result)
	}

	return result

}

//使用封装好了http
func SystemHttp(target string, result Utils.Result) Utils.Result {
	var conn http.Client
	var delay time.Duration
	// 如果是400或者不可识别协议,则使用https
	var ishttps bool
	if result.HttpStat == "400" || result.HttpStat == "tcp" {
		target = "https://" + target
		ishttps = true
	} else {
		target = "http://" + target
	}

	//如果是https或者30x跳转,则增加超时时间
	if ishttps || strings.HasPrefix(result.HttpStat, "3") {
		delay = Delay + 2
	}
	conn = Utils.HttpConn(delay)
	resp, err := conn.Get(target)
	//resp, err := conn.Get(target+"/servlet/bsh.servlet.BshServlet")
	if resp != nil && resp.TLS != nil {
		result.Protocol = "https"
		result.Host = Utils.FilterCertDomain(resp.TLS.PeerCertificates[0].DNSNames)
	}
	if err != nil {
		result.Error = err.Error()
		if strings.Contains(result.Error, "http: server gave HTTP response to HTTPS client") {
			result.Protocol = "http"
		}

		// 如果已经匹配到状态码,且再次请求报错,则返回
		if result.HttpStat != "" {
			return result
		}

		// 匹配各种错误类型
		if strings.Contains(result.Error, "context deadline exceeded") {
			result.HttpStat = "timeout"
		} else if strings.Contains(result.Error, "EOF") {
			result.HttpStat = "EOF"
		}

		return result
	}
	result.Protocol = resp.Request.URL.Scheme
	result.Stat = "OPEN"
	result.HttpStat = strconv.Itoa(resp.StatusCode)
	result.Content = Utils.GetHttpRaw(*resp)
	_ = resp.Body.Close()
	return result
}
