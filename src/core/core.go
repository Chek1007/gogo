package core

import (
	"fmt"
	"getitle/src/scan"
	. "getitle/src/structutils"
	. "getitle/src/utils"
	"github.com/panjf2000/ants/v2"
	"strings"
	"sync"
)

var Alivesum int

type targetConfig struct {
	ip     string
	port   string
	finger Frameworks
}

//直接扫描
func StraightMod(config Config) {
	var wgs sync.WaitGroup
	targetChannel := generator(config)

	scanPool, _ := ants.NewPoolWithFunc(config.Threads, func(i interface{}) {
		defaultScan(i.(targetConfig))
		wgs.Done()
	})
	defer scanPool.Release()

	for t := range targetChannel {
		wgs.Add(1)
		_ = scanPool.Invoke(t)
	}

	wgs.Wait()

}

func defaultScan(tc targetConfig) {
	//fmt.Println(ip)
	var result = new(Result)
	result.Ip = tc.ip
	result.Port = tc.port
	scan.Dispatch(result)
	//res := scan.SystemHttp(ip)

	if result.Open {
		Alivesum++
		if !Clean {
			fmt.Print(output(result, Output))
		}
		if FileHandle != nil {
			Datach <- output(result, FileOutput)
		}

	}
}

func SmartMod(config Config) {
	var taskname string
	var mask int

	// 初始化ip目标
	if config.IP != "" {
		taskname = config.IP
	} else if config.IPlist != nil {
		taskname = fmt.Sprintf("%d cidrs", len(config.IPlist))
	}

	// 初始化mask
	switch config.Mod {
	case "ss", "sc":
		mask = 16
	case "s", "sb":
		mask = 24
	}

	processLogln(fmt.Sprintf("[*] SmartScan %s, Mod: %s", taskname, config.Mod))
	var wg sync.WaitGroup
	var temp sync.Map

	//go safeMap(&temp, aliveC)
	//var ipChannel chan string
	ipChannel := ipGenerator(config.IP, config.Mod, config.IpProbeList, &temp)

	var tcChannel chan targetConfig

	// 输出启发式扫描探针
	probeconfig := fmt.Sprintf("[*] Smart probe ports: %s , ", strings.Join(config.SmartPortList, ","))
	if config.Mod == "ss" {
		probeconfig += "Smart IP probe: " + config.IpProbe
	}
	processLogln(probeconfig)

	tcChannel = tcGenerator(ipChannel, config.SmartPortList)
	scanPool, _ := ants.NewPoolWithFunc(config.Threads, func(i interface{}) {
		tc := i.(targetConfig)
		smartScan(tc, &temp, mask, config.Mod)
		wg.Done()
	})

	defer scanPool.Release()
	for t := range tcChannel {
		wg.Add(1)
		_ = scanPool.Invoke(t)
	}
	wg.Wait()

	if Noscan {
		return
	}

	var iplist []string
	temp.Range(func(ip, _ interface{}) bool {
		iplist = append(iplist, fmt.Sprintf("%s/%d", ip.(string), mask))
		return true
	})

	if iplist == nil {
		return
	}

	sort_cidr(iplist)
	if SmartFileHandle != nil {
		sync_smartips(iplist)
	}

	// 启发式扫描逐步降级,从喷洒B段到喷洒C段到默认扫描
	if config.Mod == "ss" {
		config.Mod = "s"
		declineScan(config, iplist)
	} else if config.Mod == "sc" {
		config.Mod = "sb"
		declineScan(config, iplist)
	} else if config.Mod == "s" {
		config.Mod = "default"
		config.IPlist = iplist
		spend := guessTime(config)
		processLogln(fmt.Sprintf("[*] Scan all task time is about %d seconds, Total found %d C class CIDRs ", spend, len(iplist), spend*len(iplist)))
		StraightMod(config)
	}
}

func alived(ip string, temp *sync.Map, mask int, mod string) {
	alivecidr := ip2superip(ip, mask)

	_, ok := temp.Load(alivecidr)
	if !ok {
		temp.Store(alivecidr, 1)
		cidr := fmt.Sprintf("%s/%d\n", ip, mask)
		fmt.Print("[*] Found " + cidr)
		Alivesum++
		if FileHandle != nil && SliceContains([]string{"ss", "s", "sb"}, mod) && (Noscan || mod == "sb") {
			Datach <- cidr
		}
	}
}

func smartScan(tc targetConfig, temp *sync.Map, mask int, mod string) {
	result := Result{Ip: tc.ip, Port: tc.port}
	scan.Dispatch(&result)

	if result.Open {
		alived(result.Ip, temp, mask, mod)
	}
}

func declineScan(config Config, iplist []string) {
	t := guessSmarttime(config)
	processLogln(fmt.Sprintf("[*] Every Sub smartscan task time is about %d seconds, total found %d B Class CIDRs about %ds", t, len(iplist), t*len(iplist)))
	for _, ip := range iplist {
		config.IP = ip
		tmpalive := Alivesum
		processLogln(fmt.Sprintf("[*] Spraying B class IP: %s, Estimated to take %d seconds", ip, t))
		if config.SmartPort == "default" {
			config.SmartPortList = []string{"80"}
		}
		SmartMod(config)
		processLogln(fmt.Sprintf("[*] Found %d alive cidrs from CIDR %s", Alivesum-tmpalive, ip))
		_ = FileHandle.Sync()
	}
}
