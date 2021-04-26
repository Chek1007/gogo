package core

import (
	"fmt"
	"getitle/src/Scan"
	"getitle/src/Utils"
	"github.com/panjf2000/ants/v2"
	"net"
	"sync"
	"time"
)

type TargetConfig struct {
	ip   string
	port string
}

//直接扫描
func StraightMod(config Config) {
	var wgs sync.WaitGroup

	targetChannel := generator(config)

	// Use the pool with a function,
	// set 10 to the capacity of goroutine pool and 1 second for expired duration.
	scanPool, _ := ants.NewPoolWithFunc(config.Threads, func(i interface{}) {
		defaultScan(i.(TargetConfig))
		wgs.Done()
	})
	defer scanPool.Release()

	for t := range targetChannel {
		wgs.Add(1)
		_ = scanPool.Invoke(t)
	}

	wgs.Wait()

}

func defaultScan(tc TargetConfig) {
	//fmt.Println(ip)
	var result = new(Utils.Result)
	result.Ip = tc.ip
	result.Port = tc.port
	Scan.Dispatch(result)
	//res := Scan.SystemHttp(ip)

	if result.Stat != "" {
		if !Clean {
			fmt.Print(output(result, Output))
		}
		if FileHandle != nil {
			Datach <- output(result, FileOutput)
		}

	}
}

func SmartBMod(config Config) {
	var wg sync.WaitGroup
	var temp sync.Map

	//go safeMap(&temp, aliveC)
	var ipChannel chan string
	ipChannel = ipGenerator(config, &temp)
	// 选择ip生成器

	var tcChannel chan TargetConfig

	if config.Typ == "icmp" || config.Typ == "i" {
		fmt.Println("[*] current Protocol: ICMP")
		tcChannel = tcGenerator(ipChannel, []string{"icmp"})
	} else {
		fmt.Println("[*] current Protocol: Socket")
		tcChannel = tcGenerator(ipChannel, []string{"80"})
	}

	scanPool, _ := ants.NewPoolWithFunc(config.Threads, func(i interface{}) {
		tc := i.(TargetConfig)
		smartScan(tc, &temp, false)
		wg.Done()
	})

	defer scanPool.Release()
	for t := range tcChannel {
		wg.Add(1)
		_ = scanPool.Invoke(t)
	}
	wg.Wait()
	time.Sleep(2 * time.Second)

	if Noscan {
		return
	}
	config.Mod = "default"
	var iplist []string
	temp.Range(func(key, value interface{}) bool {
		iplist = append(iplist, key.(string)+"/24")
		return true
	})
	config.IPlist = iplist
	StraightMod(config)

}

func SmartAMod(config Config) {
	var wg sync.WaitGroup
	var temp sync.Map

	bIpChannel := aIpGenerator(config.IP, &temp)
	tcChannel := tcGenerator(bIpChannel, []string{"80"})
	scanPool, _ := ants.NewPoolWithFunc(config.Threads, func(i interface{}) {
		tc := i.(TargetConfig)
		smartScan(tc, &temp, true)
		wg.Done()
	})

	defer scanPool.Release()
	for t := range tcChannel {
		wg.Add(1)
		_ = scanPool.Invoke(t)
	}
	wg.Wait()

	var iplist []string
	temp.Range(func(key, value interface{}) bool {
		iplist = append(iplist, key.(string)+"/16")
		return true
	})
	for _, ip := range iplist {
		config.IP = ip
		fmt.Println("[*] " + Utils.GetCurtime() + " Spraying B class IP:" + ip)
		SmartBMod(config)
	}
	//for i := range btargetChannel {
	//	fmt.Println("[*]" + Utils.GetCurtime() + "Spraying B class IP:" + i)
	//	var tmpconfig = config
	//	tmpconfig.IP = i + "/16"
	//	SmartBMod(tmpconfig)
	//}
}

func c_alived(ip string, temp *sync.Map) {
	s2ip := net.ParseIP(ip).To4()
	s2ip[3] = 1
	aliveC := s2ip.String()
	_, ok := temp.Load(aliveC)

	if !ok {
		temp.Store(aliveC, 1)
		fmt.Println("[*] Find " + aliveC + "/24")
		if FileHandle != nil && Noscan {
			Datach <- aliveC + "/24\n"
		}
	}
}

func b_alived(ip string, temp *sync.Map) {
	s2ip := net.ParseIP(ip).To4()
	s2ip[3] = 1
	s2ip[2] = 1
	aliveB := s2ip.String()

	_, ok := temp.Load(aliveB)
	if !ok {
		temp.Store(aliveB, 1)
		fmt.Println("[*] Find " + aliveB + "/16")
	}
}

func smartScan(tc TargetConfig, temp *sync.Map, isB bool) {
	var result = new(Utils.Result)
	result.Ip = tc.ip
	result.Port = tc.port
	result.HttpStat = "s"

	Scan.Dispatch(result)

	if result.Stat == "OPEN" {
		if isB {
			b_alived(result.Ip, temp)
		} else {
			c_alived(result.Ip, temp)
		}
	}
}
