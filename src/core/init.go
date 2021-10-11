package core

import (
	"fmt"
	"getitle/src/scan"
	. "getitle/src/structutils"
	. "getitle/src/utils"
	"os"
	"runtime"
	"strconv"
	"strings"
)

//文件输出
var Datach = make(chan string, 100)
var FileHandle *os.File // 输出文件 handle

var Output string     // 命令行输出格式
var FileOutput string // 文件输出格式

//进度tmp文件
var LogDetach = make(chan string, 100)
var LogFileHandle *os.File

var Clean bool
var Noscan bool

var InterConfig = map[string][]string{
	"10.0.0.0/8":     {"ss", "icmp", "1"},
	"172.16.0.0/12":  {"ss", "icmp", "1"},
	"192.168.0.0/16": {"s", "80", "all"},
	"100.100.0.0/16": {"s", "icmp", "all"},
	"200.200.0.0/16": {"s", "icmp", "all"},
	"169.254.0.0/16": {"s", "icmp", "all"},
	"168.254.0.0/16": {"s", "icmp", "all"},
}

func Init(config Config) Config {
	//println("*********  main 0.9.4 beta by Sangfor  *********")

	//if config.Mod != "default" && config.ListFile != "" {
	//	println("[-] error Smart scan config")
	//	os.Exit(0)
	//}

	// check命令行参数
	checkCommand(config)

	// 初始化

	//windows系统默认协程数为2000
	OS := runtime.GOOS
	if config.Threads == 4000 { // if 默认线程
		if OS == "windows" {
			config.Threads = 1000
		} else if config.JsonFile != "" {
			config.Threads = 1000
		}
	}

	// 初始化文件操作
	initFile(config)

	// 如果输入的json不为空,则从json中加载result,并返回结果
	if config.JsonFile != "" {
		config.Results = LoadResult(config.JsonFile).Data
		return config
	}

	// 初始化启发式扫描的端口探针
	if config.SmartPort != "default" {
		config.SmartPortList = portHandler(config.SmartPort)
	} else {
		if config.Mod == "s" {
			config.SmartPortList = []string{"80"}
		} else if SliceContains([]string{"ss", "sc", "f"}, config.Mod) {
			config.SmartPortList = []string{"icmp"}
		}
	}

	// 初始化ss模式ip探针,默认ss默认只探测ip为1的c段,可以通过-ipp参数指定,例如-ipp 1,254,253
	if config.IpProbe != "default" {
		config.IpProbeList = Str2uintlist(config.IpProbe)
	} else {
		config.IpProbeList = []uint{1}
	}

	// 初始化端口配置
	config.Portlist = portHandler(config.Ports)
	// 如果从文件中读,初始化IP列表配置
	if config.ListFile != "" {
		config.IPlist = readTargetFile(config.ListFile)
	}

	//if config.Spray && config.Mod != "default" {
	//	println("[-] error Spray scan config")
	//	os.Exit(0)
	//}
	// 文件操作

	return config
}

func checkCommand(config Config) {
	// 一些命令行参数错误处理,如果check没过直接退出程序或输出警告
	if config.Mod == "ss" && config.ListFile != "" {
		fmt.Println("[-] error Smart scan can not use File input")
		os.Exit(0)
	}
	if config.JsonFile != "" {
		if config.Ports != "top1" {
			fmt.Println("[warn] json input can not config ports")
		}
		if config.Mod != "default" {
			fmt.Println("[warn] json input can not config scan Mod,default scanning")
		}
	}
	if config.IP == "" && config.ListFile == "" && config.Mod != "a" { // 一些导致报错的参数组合
		fmt.Println("[-] mod AUTO can not define IP or IPlist")
		os.Exit(0)
	}
}

func printTaskInfo(config Config, taskname string) {
	// 输出任务的基本信息

	fmt.Printf("[*] Current goroutines: %d, Version Level: %d,Exploit Target: %s, Spray Scan: %t\n", config.Threads, scan.VersionLevel, scan.Exploit, config.Spray)
	if config.JsonFile == "" {
		processLogln(fmt.Sprintf("[*] Start scan task %s ,total ports: %d , mod: %s", taskname, len(config.Portlist), config.Mod))
		if len(config.Portlist) > 500 {
			fmt.Println("[*] too much ports , only show top 500 ports: " + strings.Join(config.Portlist[:500], ",") + "......")
		} else {
			fmt.Println("[*] ports: " + strings.Join(config.Portlist, ","))
		}
		if config.Mod == "default" {
			processLogln(fmt.Sprintf("[*] Estimated to take about %d seconds", guesstime(config)))
		}
	} else {
		processLogln(fmt.Sprintf("[*] Start scan task %s ,total target: %d", taskname, len(config.Results)))
		processLogln(fmt.Sprintf("[*] Estimated to take about %d seconds", (len(config.Results)/config.Threads)*4+4))
	}
}

func RunTask(config Config) {
	var taskname string
	if config.Mod == "a" {
		// 内网探测默认使用icmp扫描
		taskname = "Reserved interIP addresses"
	} else {
		config = ipInit(config)
		if config.IP != "" {
			taskname = config.IP
		} else if config.ListFile != "" {
			taskname = config.ListFile
		} else if config.JsonFile != "" {
			taskname = config.JsonFile
		}
	}
	if taskname == "" {
		fmt.Println("[-] No Task")
		os.Exit(0)
	}

	// 如果指定端口超过100,则自动启用spray
	if len(config.Portlist) > 150 && !config.NoSpray {
		config.Spray = true
	}
	// 输出任务的基本信息
	printTaskInfo(config, taskname)

	switch config.Mod {
	case "default":
		StraightMod(config)
	case "a", "auto":
		autoScan(config)
	case "s", "f", "ss", "sc":
		mask := getMask(config.IP)
		if mask >= 24 {
			config.Mod = "default"
			StraightMod(config)
		} else {
			SmartMod(config)
		}
	default:
		StraightMod(config)
	}
}

func guesstime(config Config) int {
	ipcount := 0
	portcount := len(config.Portlist)
	if config.IPlist != nil {
		for _, ip := range config.IPlist {
			ipcount += countip(ip)
		}
	} else {
		ipcount = countip(config.IP)
	}
	return (portcount*ipcount/config.Threads)*4 + 4
}

func countip(ip string) int {
	count := 0
	c, _ := strconv.Atoi(strings.Split(ip, "/")[1])
	if c == 32 {
		count++
	} else {
		count += 2 << (31 - uint(c))
	}
	return count
}

func autoScan(config Config) {
	for cidr, st := range InterConfig {
		processLogln("[*] Spraying : " + cidr)
		SmartMod(createSmartTask(config, cidr, st))
	}
}
func createSmartTask(config Config, cidr string, c []string) Config {
	config.IP = cidr
	config.SmartPortList = portHandler(c[1])
	config.Mod = c[0]
	if c[2] != "all" {
		config.IpProbe = c[2]
		config.IpProbeList = Str2uintlist(c[2])
	}
	return config
}
