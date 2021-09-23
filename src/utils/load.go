package utils

import (
	"encoding/json"
	"fmt"
	"getitle/src/nuclei"
	"os"
	"regexp"
	"strings"
)

var Mmh3fingers, Md5fingers = loadHashFinger()
var Tcpfingers = loadFingers("tcp")
var Httpfingers = loadFingers("http")
var Tagmap, Namemap, Portmap = loadPortConfig()
var Compiled = make(map[string][]regexp.Regexp)
var CommonCompiled = initregexp()
var TemplateMap = loadTemplates()

func loadTemplates() map[string][]*nuclei.Template {
	var templates []nuclei.Template
	var templatemap = make(map[string][]*nuclei.Template)
	err := json.Unmarshal([]byte(LoadConfig("nuclei")), &templates)
	if err != nil {
		fmt.Println("[-] nuclei config load FAIL!")
		os.Exit(0)
	}
	for index, template := range templates {
		// 以指纹归类
		if template.Finger != "" {
			templatemap[strings.ToLower(template.Finger)] = append(templatemap[template.Finger], &templates[index])
		}

		// 以tag归类
		for _, tag := range template.GetTags() {
			templatemap[tag] = append(templatemap[tag], &templates[index])
		}
	}
	return templatemap
}
func loadPortConfig() (PortMapper, PortMapper, PortMapper) {
	var portfingers []PortFinger
	err := json.Unmarshal([]byte(LoadConfig("port")), &portfingers)

	if err != nil {
		fmt.Println("[-] port config load FAIL!")
		os.Exit(0)
	}
	tagmap := make(PortMapper)  // 以服务名归类
	namemap := make(PortMapper) // 以tag归类
	portmap := make(PortMapper) // 以端口号归类

	for _, v := range portfingers {
		v.Ports = Ports2PortSlice(v.Ports)
		namemap[v.Name] = append(namemap[v.Name], v.Ports...)
		for _, t := range v.Type {
			tagmap[t] = append(tagmap[t], v.Ports...)
		}
		for _, p := range v.Ports {
			portmap[p] = append(portmap[p], v.Name)
		}
	}

	return tagmap, namemap, portmap
}

//加载指纹到全局变量
func loadFingers(t string) *FingerMapper {
	var tmpfingers []Finger
	var fingermap = make(FingerMapper)
	// 根据权重排序在python脚本中已经实现
	err := json.Unmarshal([]byte(LoadConfig(t)), &tmpfingers)
	if err != nil {
		fmt.Println("[-] finger load FAIL!")
		os.Exit(0)
	}

	//初步处理tcp指纹
	for _, finger := range tmpfingers {
		finger.Decode() // 防止\xff \x00编码解码影响结果

		// 普通指纹
		for _, regstr := range finger.Regexps.Regexp {
			Compiled[finger.Name] = append(Compiled[finger.Name], compile("(?im)"+regstr))
		}
		// 漏洞指纹,指纹名称后接 "_vuln"
		for _, regstr := range finger.Regexps.Vuln {
			Compiled[finger.Name+"_vuln"] = append(Compiled[finger.Name], compile("(?im)"+regstr))
		}

		// http默认为80
		if finger.Defaultport == nil && finger.Protocol == "http" {
			finger.Defaultport = []string{"80"}
		}

		// 根据端口分类指纹
		for _, ports := range finger.Defaultport {
			for _, port := range port2PortSlice(ports) {
				fingermap[port] = append(fingermap[port], finger)
			}
		}

	}
	return &fingermap
}

func loadHashFinger() (map[string]string, map[string]string) {
	var mmh3fingers, md5fingers map[string]string
	var err error
	err = json.Unmarshal([]byte(LoadConfig("mmh3")), &mmh3fingers)
	if err != nil {
		fmt.Println("[-] mmh3 load FAIL!")
		os.Exit(0)
	}

	err = json.Unmarshal([]byte(LoadConfig("md5")), &md5fingers)
	if err != nil {
		fmt.Println("[-] mmh3 load FAIL!")
		os.Exit(0)
	}
	return mmh3fingers, md5fingers
}

func initregexp() map[string]regexp.Regexp {
	comp := make(map[string]regexp.Regexp)
	comp["title"] = compile("(?Uis)<title>(.*)</title>")
	comp["server"] = compile("(?i)Server: ([\x20-\x7e]+)")
	comp["xpb"] = compile("(?i)X-Powered-By: ([\x20-\x7e]+)")
	comp["sessionid"] = compile("(?i) (.*SESS.*?ID)")
	return comp
}
