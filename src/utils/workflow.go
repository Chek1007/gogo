package utils

import (
	"encoding/json"
	"path"
)

type Workflow struct {
	Name       string   `json:"name"`
	IP         string   `json:"ip"`
	Ports      string   `json:"ports"`
	Mod        string   `json:"mod"`
	Ping       bool     `json:"ping"`
	Arp        bool     `json:"arp"`
	NoScan     bool     `json:"no-scan"`
	IpProbe    string   `json:"ipprobe"`
	SmartProbe string   `json:"portprobe"`
	Exploit    string   `json:"exploit"`
	Version    int      `json:"version"`
	File       string   `json:"file"`
	Path       string   `json:"path"`
	Tags       []string `json:"tags"`
}

func ParseWorkflowsFromInput(content []byte) []*Workflow {
	var workflows []*Workflow
	var err error
	err = json.Unmarshal(content, &workflows)
	if err != nil {
		Panic("workflow load FAIL, " + err.Error())
	}
	return workflows
}

func (w *Workflow) PrepareConfig() *Config {
	var config = &Config{
		IP:        w.IP,
		Ports:     w.Ports,
		Mod:       w.Mod,
		IpProbe:   w.IpProbe,
		SmartPort: w.SmartProbe,
	}
	// 参数默认值
	if w.IpProbe == "" {
		w.IpProbe = "default"
	}
	if w.SmartProbe == "" {
		w.SmartProbe = "default"
	}
	if w.Ports == "" {
		w.Ports = "top1"
	}
	if w.Mod == "" {
		w.Mod = "default"
	}
	if w.File == "" {
		w.File = "auto"
	}
	if w.Exploit == "" {
		w.Exploit = "none"
	}

	if w.Arp {
		config.AliveSprayMod = append(config.AliveSprayMod, "arp")
	}
	if w.Ping {
		config.AliveSprayMod = append(config.AliveSprayMod, "icmp")
	}

	var autofile, hiddenfile bool
	if w.File == "auto" {
		autofile = true
	} else if w.File == "hidden" {
		hiddenfile = true
	} else {
		config.Filename = path.Join(w.Path, w.File)
	}

	if config.Filename == "" {
		config.Filename = GetFilename(config, autofile, hiddenfile, w.Path, "json")
		if config.IsSmartScan() && !w.NoScan {
			config.SmartFilename = GetFilename(config, autofile, hiddenfile, w.Path, "cidr")
		}
		if config.HasAlivedScan() {
			config.PingFilename = GetFilename(config, autofile, hiddenfile, w.Path, "alived")
		}
	}
	return config
}
