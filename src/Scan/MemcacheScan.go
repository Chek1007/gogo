package Scan

import (
	"getitle/src/Utils"
	"strings"
)

func MemcacheScan(target string, result *Utils.Result) *Utils.Result {
	conn, err := Utils.TcpSocketConn(target, Delay)
	if err != nil {

		//fmt.Println(err)
		result.Error = err.Error()
		return result
	}

	_, recv, _ := Utils.SocketSend(conn, []byte("stats\n"), 1024)
	if strings.Contains(string(recv), "STAT version") {
		result.Protocol = "tcp"
		result.Framework = "memcache"
		result.Title = "memcache " + Utils.Match("STAT version (.*)", string(recv))
		result.Vuln = "memcache Unauth"
	}
	return result
}
