package plugin

import (
	. "github.com/chainreactors/gogo/v2/pkg"
	"github.com/chainreactors/logs"
)

var (
	TCP = "tcp"
	UDP = "udp"
)

func socketFingerScan(result *Result) {
	// 如果是http协议,则判断cms,如果是tcp则匹配规则库.暂时不考虑udp
	if Proxy != nil {
		// 如果存在http代理，跳过tcp指纹识别
		return
	}

	tcpsender := func(sendData []byte) ([]byte, bool) {
		target := result.GetTarget()
		logs.Log.Debugf("active detect: %s, data: %q", target, sendData)
		conn, err := NewSocket(TCP, target, RunOpt.Delay)
		if err != nil {
			logs.Log.Debugf("active detect %s error, %s", target, err.Error())
			return nil, false
		}
		defer conn.Close()

		data, err := conn.QuickRequest(sendData, 1024)
		if err != nil {
			return nil, false
		}

		return data, true
	}

	//udpsender := func(sendData []byte) ([]byte, bool) {
	//	target := result.GetTarget()
	//	logs.Log.Debugf("active detect: , data: ", target, sendData)
	//	conn, err := NewSocket(UDP, target, RunOpt.Delay)
	//	if err != nil {
	//		logs.Log.Debugf("active detect %s error, %s", target, err.Error())
	//		return nil, false
	//	}
	//	defer conn.Close()
	//
	//	data, err := conn.QuickRequest(sendData, 1024)
	//	if err != nil {
	//		return nil, false
	//	}
	//
	//	return data, true
	//}

	f, v := FingerEngine.SocketMatch(result.Content, result.Port, RunOpt.VersionLevel, tcpsender)
	if f != nil {
		result.AddFramework(f)
		if v != nil {
			result.AddVuln(v)
		}
	}
	return
}
