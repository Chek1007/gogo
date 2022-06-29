package pkg

import (
	"encoding/json"
	"getitle/v1/pkg/utils"
	"strings"
)

type Config struct {
	IP            string              `json:"ip"`
	IPlist        []string            `json:"ips"`
	ExcludeIPs    string              `json:"-"`
	ExcludeMap    map[uint]bool       `json:"-"`
	Ports         string              `json:"ports"`     // 预设字符串
	Portlist      []string            `json:"-"`         // 处理完的端口列表
	JsonFile      string              `json:"json_file"` // gt的结果json文件,可以再次读入扫描
	Results       Results             `json:"-"`         // json反序列化后的内网,保存在内存中
	ListFile      string              `json:"list_file"` // 目标ip列表
	HostsMap      map[string][]string `json:"-"`         // host映射表
	Threads       int                 `json:"threads"`   // 线程数
	Mod           string              `json:"mod"`       // 扫描模式
	SmartPort     string              `json:"-"`         // 启发式扫描预设探针
	SmartPortList []string            `json:"-"`         // 启发式扫描预设探针
	IpProbe       string              `json:"-"`
	IpProbeList   []uint              `json:"-"`
	Output        string              `json:"-"`
	Filename      string              `json:"-"`
	SmartFilename string              `json:"-"`
	PingFilename  string              `json:"-"`
	PortSpray     bool                `json:"port_spray"`
	AliveSprayMod []string            `json:"alive_spray"`
	NoSpray       bool                `json:"-"`
	Exploit       string              `json:"exploit"`
	JsonType      string              `json:"json_type"`
	VersionLevel  int                 `json:"version_level"`
	IsListInput   bool                `json:"-"`
	IsJsonInput   bool                `json:"-"`
}

func (config *Config) IsScan() bool {
	if config.IP != "" || config.ListFile != "" || config.JsonFile != "" || config.Mod == "a" {
		return true
	}
	return false
}

func (config *Config) IsSmart() bool {
	if utils.SliceContains([]string{"ss", "s", "sc"}, config.Mod) {
		return true
	}
	return false
}

func (config *Config) IsSmartScan() bool {
	if utils.SliceContains([]string{"ss", "s"}, config.Mod) {
		return true
	}
	return false
}

func (config *Config) IsASmart() bool {
	if utils.SliceContains([]string{"ss", "sc"}, config.Mod) {
		return true
	}
	return false
}

func (config *Config) IsBSmart() bool {
	if utils.SliceContains([]string{"s", "sb"}, config.Mod) {
		return true
	}
	return false
}

func (config *Config) HasAlivedScan() bool {
	if len(config.AliveSprayMod) > 0 {
		return true
	}
	return false
}

func (config *Config) InitIP() {
	config.HostsMap = make(map[string][]string)
	// 优先处理ip
	if config.IP != "" {
		if strings.Contains(config.IP, ",") {
			config.IPlist = strings.Split(config.IP, ",")
		} else {
			var host string
			config.IP, host = ParseCIDR(config.IP)
			if host != "" {
				ip, _ := SplitCIDR(config.IP)
				config.HostsMap[ip] = append(config.HostsMap[ip], host)
			}
			if config.IP == "" {
				utils.Fatal("IP format error")
			}
		}
	}

	// 如果输入的是文件,则格式化所有输入值.如果无有效ip
	if config.IPlist != nil {
		var iplist []string
		for _, ip := range config.IPlist {
			ip, host := ParseCIDR(ip)
			if host != "" {
				i, _ := SplitCIDR(ip)
				config.HostsMap[i] = append(config.HostsMap[i], host)
			}
			if ip != "" {
				iplist = append(iplist, ip)
			}
		}
		config.IPlist = utils.SliceUnique(iplist) // 去重
		if len(config.IPlist) == 0 {
			utils.Fatal("all targets format error")
		}
	}
}

func (config *Config) GetTarget() string {
	if config.IP != "" {
		return config.IP
	} else if config.ListFile != "" {
		return strings.Join(config.IPlist, ",")
	} else if config.JsonFile != "" {
		return config.JsonFile
	} else {
		return ""
	}
}

func (config *Config) GetTargetName() string {
	var target string
	if config.ListFile != "" {
		target = config.ListFile
	} else if config.JsonFile != "" {
		target = config.JsonFile
	} else if config.Mod == "a" {
		target = "auto"
	} else if config.IP != "" {
		target = config.IP
	}
	return target
}

func (config *Config) ToJson(json_type string) string {
	config.JsonType = json_type
	s, err := json.Marshal(config)
	if err != nil {
		return err.Error()
	}
	return string(s)
}
