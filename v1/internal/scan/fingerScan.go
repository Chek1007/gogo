package scan

import (
	. "getitle/v1/pkg"
	. "getitle/v1/pkg/fingers"
	"github.com/chainreactors/logs"
	"net"
)

func fingerScan(result *Result) {
	//如果是http协议,则判断cms,如果是tcp则匹配规则库.暂时不考虑udp
	if result.IsHttp() {
		getFramework(result, HttpFingers, httpFingerMatch)
	} else {
		getFramework(result, TcpFingers, tcpFingerMatch)
	}
	return
}

func getFramework(result *Result, fingermap FingerMapper, matcher func(*Result, *Finger) (*Framework, *Vuln)) {
	// 优先匹配默认端口,第一次循环只匹配默认端口
	//var fs Frameworks
	var alreadyFrameworks Fingers
	for _, finger := range fingermap.GetFingers(result.Port) {
		framework, vuln := matcher(result, finger)
		alreadyFrameworks = append(alreadyFrameworks, finger)
		if framework != nil {
			if vuln != nil {
				result.AddVuln(vuln)
			}
			result.AddFramework(framework)
			if result.Protocol == "tcp" {
				// 如果是tcp协议,并且识别到一个指纹,则退出.
				// 如果是http协议,可能存在多个指纹,则进行扫描
				return
			}
		}
	}

	for port, fingers := range fingermap {
		if port == result.Port {
			// 跳过已经扫过的默认端口
			continue
		}
		for _, finger := range fingers {
			if alreadyFrameworks.Contain(finger) {
				continue
			} else {
				alreadyFrameworks = append(alreadyFrameworks, finger)
			}
			framework, vuln := matcher(result, finger)
			if framework != nil {
				result.AddFramework(framework)
				if result.Protocol == "tcp" {
					return
				}
			}
			if vuln != nil {
				result.AddVuln(vuln)
			}
		}
	}
	return
}

func httpFingerMatch(result *Result, finger *Finger) (*Framework, *Vuln) {
	resp := result.Httpresp
	content := result.Content
	var body string
	var rerequest bool
	//var cookies map[string]string
	for i, rule := range finger.Rules {
		if RunOpt.VersionLevel >= 1 && rule.SendDataStr != "" {
			// 如果level大于1,并且存在主动发包, 则重新获取resp与content
			conn := result.GetHttpConn(RunOpt.Delay)
			url := result.GetURL() + rule.SendDataStr
			tmpresp, err := conn.Get(url)
			if err == nil {
				logs.Log.Debugf("request finger %s %d for %s", url, tmpresp.StatusCode, finger.Name)
				resp = tmpresp
				content, body = GetHttpRaw(resp)
				rerequest = true
			} else {
				logs.Log.Debugf("request finger %s %s for %s", url, err.Error(), finger.Name)
			}
		}

		hasFrame, hasVuln, res := RuleMatcher(rule, content, true)
		if hasFrame {
			frame, vuln := finger.ToResult(hasFrame, hasVuln, res, i)
			if rerequest {
				// 如果主动发包匹配到了指纹,则重新进行信息收集
				frame.From = "active"
				CollectHttpInfo(result, resp, content, body)
			}
			return frame, vuln
		}
	}
	return nil, nil
}

func tcpFingerMatch(result *Result, finger *Finger) (*Framework, *Vuln) {
	content := result.Content
	var data []byte
	var err error

	for i, rule := range finger.Rules {
		// 某些规则需要主动发送一个数据包探测
		if rule.SendDataStr != "" && RunOpt.VersionLevel >= rule.Level {
			logs.Log.Debugf("request finger %s for %s", result.GetTarget(), finger.Name)
			var conn net.Conn
			conn, err = TcpSocketConn(result.GetTarget(), 2)
			if err != nil {
				return nil, nil
			}
			data, err = SocketSend(conn, rule.SendData, 1024)
			// 如果报错为EOF,则需要重新建立tcp连接
			if err != nil {
				return nil, nil
			}
		}
		// 如果主动探测有回包,则正则匹配回包内容, 若主动探测没有返回内容,则直接跳过该规则
		if len(data) != 0 {
			content = string(data)
		}

		hasFrame, hasVuln, res := RuleMatcher(rule, content, false)
		if hasFrame {
			frame, vuln := finger.ToResult(hasFrame, hasVuln, res, i)
			return frame, vuln
		}
	}

	return nil, nil
}
