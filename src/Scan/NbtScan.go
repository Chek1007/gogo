package Scan

import (
	"encoding/hex"
	"getitle/src/Utils"
	"net"
	"strconv"
	"strings"
	"time"
)

var UNIQUE_NAMES, GROUP_NAMES, NetBIOS_ITEM_TYPE map[string]string

func Sendpayload(target string) []byte {
	conn, err := net.DialTimeout("udp", target+":137", 2*time.Second)
	if err != nil {

		//fmt.Println(err)
		return []byte("\x00")
	}

	err = conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	if err != nil {

		//fmt.Println(err)
		return []byte("\x00")
	}

	payload := []byte("ff\x00\x00\x00\x01\x00\x00\x00\x00\x00\x00 CKAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\x00\x00!\x00\x01")

	reply := Utils.SocketSend(conn, payload)

	return reply
}

func Byte2Int(input []byte) (int, error) {
	encodedStr := hex.EncodeToString(input)
	output, err := strconv.Atoi(encodedStr)
	return output, err
}

func init() {
	UNIQUE_NAMES = map[string]string{
		"\x00": "Workstation Service",
		"\x03": "Messenger Service",
		"\x06": "RAS Server Service",
		"\x1F": "NetDDE Service",
		"\x20": "File Server Service",
		"\x21": "RAS Client Service",
		"\xBE": "Network Monitor Agent",
		"\xBF": "Network Monitor Application",
		"\x1D": "Master Browser",
		"\x1B": "Domain Master Browser",
	}

	GROUP_NAMES = map[string]string{
		"\x00": "Domain Name",
		"\x1C": "Domain Controllers",
		"\x1E": "Browser Service Elections",
	}

	NetBIOS_ITEM_TYPE = map[string]string{
		"\x01\x00": "NetBIOS computer name",
		"\x02\x00": "NetBIOS domain name",
		"\x03\x00": "DNS computer name",
		"\x04\x00": "DNS domain name",
		"\x05\x00": "DNS tree name",
		"\x07\x00": "Time stamp",
	}
}

func NbtScan(ip string, result Utils.Result) Utils.Result {

	var Share bool = false
	reply := Sendpayload(ip)

	if len(reply) > 58 {
		result.Stat = "OPEN"

	} else {
		result.Stat = "CLOSE"
		return result
	}

	num, err := Byte2Int(reply[56:57])
	if err != nil {
		return result
	}
	var name, group, unique string
	var flag_bit []byte
	data := reply[57:]
	for i := 0; i < num; i++ {
		name = string(data[18*i : 18*i+15])
		flag_bit = data[18*i+15 : 18*i+16]
		//fmt.Println(name)
		if string(flag_bit) == "\x00" {
			name_flags := data[18*i+16 : 18*i+18]
			num, _ := Byte2Int(name_flags[0:1])
			if num >= 80 {
				group = strings.Trim(name, " ")
				//fmt.Printf("%s\t%s\t%s\n",name,"G",GROUP_NAMES[string(flag_bit)])
			} else {
				unique = name
				if string(flag_bit) == "\x20" {
					Share = true
				}
				//fmt.Printf("%s\t%s\t%s\n",name,"U",UNIQUE_NAMES[string(flag_bit)])
			}
		} else {
			if _, ok := GROUP_NAMES[string(flag_bit)]; ok {
				//fmt.Printf("%s\t%s\t%s\n",name,"G",GROUP_NAMES[string(flag_bit)])
			} else if _, ok := UNIQUE_NAMES[string(flag_bit)]; ok {
				if string(flag_bit) == "\x20" {
					Share = true
				}
				//fmt.Printf("%s\t%s\t%s\n",name,"U",UNIQUE_NAMES[string(flag_bit)])
			}
		}

	}
	msg := group + "\\" + unique

	if Share {
		msg = msg + "        sharing"
	}
	result.Title = msg

	return result
}
