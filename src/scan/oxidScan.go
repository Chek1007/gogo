package scan

import (
	"getitle/src/utils"
	"strings"
)

var sendData = "\x05\x00\x0b\x03\x10\x00\x00\x00\x48\x00\x00\x00\x01\x00\x00\x00\xb8\x10\xb8\x10\x00\x00\x00\x00\x01\x00\x00\x00\x00\x00\x01\x00\xc4\xfe\xfc\x99\x60\x52\x1b\x10\xbb\xcb\x00\xaa\x00\x21\x34\x7a\x00\x00\x00\x00\x04\x5d\x88\x8a\xeb\x1c\xc9\x11\x9f\xe8\x08\x00\x2b\x10\x48\x60\x02\x00\x00\x00"
var sendData2 = "\x05\x00\x00\x03\x10\x00\x00\x00\x18\x00\x00\x00\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x05\x00"

// -default
func oxidScan(result *utils.Result) {
	result.Port = "135"
	target := result.GetTarget()
	conn, err := utils.TcpSocketConn(target, Delay)
	if err != nil {

		//fmt.Println(err)
		result.Error = err.Error()
		return
	}
	defer conn.Close()
	result.Open = true

	recv, _ := utils.SocketSend(conn, []byte(sendData), 4096)
	recv, _ = utils.SocketSend(conn, []byte(sendData2), 4096)
	recvStr := string(recv)
	if len(recvStr) < 42 {
		return
	}
	recvStr_v2 := recvStr[42:]
	packet_v2_end := strings.Index(recvStr_v2, "\x09\x00\xff\xff\x00\x00")
	if packet_v2_end == -1 {
		return
	}
	packet_v2 := recvStr_v2[:packet_v2_end]
	packet_v2 = strings.Replace(packet_v2, "\x00", "", -1)
	hostname_list := strings.Split(packet_v2, "\x07")

	result.Host = hostname_list[0]
	result.Title += strings.Join(hostname_list[1:], ",")
	result.HttpStat = "OXID"
	result.Protocol = "wmi"
	return
}
