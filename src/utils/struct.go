package utils

import (
	b64 "encoding/base64"
)

// common struct
func decode(s string) []byte {
	var bs []byte
	if s[:4] == "b64|" {
		bs, _ = b64.StdEncoding.DecodeString(s[4:])
	} else {
		bs = []byte(s)
	}
	return bs
}

type Finger struct {
	Name         string   `json:"name"`
	Protocol     string   `json:"protocol"`
	SendData_str string   `json:"send_data"`
	SendData     senddata `json:"-"`
	Vuln         string   `json:"vuln"`
	Level        int      `json:"level"`
	Defaultport  []string `json:"default_port"`
	Regexps      Regexps  `json:"regexps"`
}

func (f *Finger) Decode() {
	if f.Protocol != "tcp" {
		return
	}

	if f.SendData_str != "" {
		f.SendData = decode(f.SendData_str)
	}
	// todo
	// regexp decode
}

type senddata []byte

func (d senddata) IsNull() bool {
	if len(d) == 0 {
		return true
	}
	return false
}

type Regexps struct {
	Body   []string `json:"body"`
	MD5    []string `json:"md5"`
	Regexp []string `json:"regexp"`
	Cookie []string `json:"cookie"`
	Header []string `json:"header"`
	Vuln   []string `json:"vuln"`
}

type PortFinger struct {
	Name  string   `json:"name"`
	Ports []string `json:"ports"`
	Type  []string `json:"type"`
}

type PortMapper map[string][]string

type FingerMapper map[string][]*Finger

func (fm FingerMapper) GetFingers(port string) []*Finger {
	return fm[port]
}

func (fm FingerMapper) GetOthersFingers(port string) []*Finger {
	var tmpfingers []*Finger
	for _, fingers := range fm {
		for _, finger := range fingers {
			isrepeat := false
			for _, tmpfinger := range tmpfingers {
				if finger == tmpfinger {
					isrepeat = true
				}
			}
			if !isrepeat {
				tmpfingers = append(tmpfingers, finger)
			}
		}
	}
	return tmpfingers
}
