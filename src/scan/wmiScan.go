package scan

import (
	"bytes"
	"getitle/src/utils"
)

var payload = []byte{5, 0, 11, 3, 16, 0, 0, 0, 120, 0, 40, 0, 3, 0, 0, 0, 184, 16, 184, 16, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 1, 0, 160, 1, 0, 0, 0, 0, 0, 0, 192, 0, 0, 0, 0, 0, 0, 70, 0, 0, 0, 0, 4, 93, 136, 138, 235, 28, 201, 17, 159, 232, 8, 0, 43, 16, 72, 96, 2, 0, 0, 0, 10, 2, 0, 0, 0, 0, 0, 0, 78, 84, 76, 77, 83, 83, 80, 0, 1, 0, 0, 0, 7, 130, 8, 162, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 6, 1, 177, 29, 0, 0, 0, 15}

func wmiScan(target string, result *utils.Result) {
	conn, err := utils.TcpSocketConn(target, Delay)
	if err != nil {
		return
	}
	result.Stat = "OPEN"
	ret, err := utils.SocketSend(conn, payload, 4096)
	if err != nil {
		return
	}
	off_ntlm := bytes.Index(ret, []byte("NTLMSSP"))
	if off_ntlm != -1 {
		result.Protocol = "wmi"
		result.HttpStat = "WMI"
		tinfo := NTLMInfo(ret[off_ntlm:])
		result.AddNTLMInfo(tinfo)
	}
}
