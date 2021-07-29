package core

import (
	"fmt"
	"getitle/src/Utils"
	"net"
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

type Config struct {
	IP            string
	IPlist        []string
	Ports         string
	Portlist      []string
	List          string
	Threads       int
	Mod           string
	SmartPort     string
	SmartPortList []string
	IpProbe       string
	IpProbeList   []uint
	Output        string
	Filename      string
	Spray         bool
}

func Init(config Config) Config {
	//println("*********  main 0.3.3 beta by Sangfor  *********")

	//if config.Mod != "default" && config.List != "" {
	//	println("[-] error Smart scan config")
	//	os.Exit(0)
	//}

	if config.Mod == "ss" && config.List != "" {
		fmt.Println("[-] error Smart scan config")
		os.Exit(0)
	}

	// 初始化启发式扫描的端口
	if config.SmartPort != "default" {
		config.SmartPortList = PortHandler(config.SmartPort)
	} else {
		if config.Mod == "s" {
			config.SmartPortList = []string{"80"}
		} else if config.Mod == "ss" || config.Mod == "f" {
			config.SmartPortList = []string{"icmp"}
		}
	}

	// 默认ss默认只探测ip为1的c段,可以通过-ipp参数指定,例如-ipp 1,254,253
	if config.IpProbe != "default" {
		config.IpProbeList = Utils.Str2uintlist(config.IpProbe)
	} else {
		config.IpProbeList = []uint{1}
	}

	// 初始化端口配置
	config.Portlist = PortHandler(config.Ports)

	// 如果从文件中读,初始化IP列表配置
	if config.List != "" {
		config.IPlist = ReadTargetFile(config.List)
	}

	//if config.Spray && config.Mod != "default" {
	//	println("[-] error Spray scan config")
	//	os.Exit(0)
	//}

	//windows系统默认协程数为2000
	OS := runtime.GOOS
	if config.Threads == 4000 && OS == "windows" {
		config.Threads = 2000
	}

	if config.IP == "" && config.List == "" && config.Mod != "a" {
		os.Exit(0)
	}
	// 存在文件输出则停止命令行输出
	if config.Filename != "" {
		Clean = !Clean
		// 创建filehandle
		FileHandle = initFileHandle(config.Filename)
		if FileOutput == "json" && !Noscan {
			_, _ = FileHandle.WriteString("[")
		}

	}

	_ = os.Remove(".sock.lock")
	LogFileHandle = initFileHandle(".sock.lock")
	initFile()
	// 进度文件,任务完成后自动删除

	return config
}

func RunTask(config Config) {
	var taskname string = ""
	if config.Mod == "a" {
		// 内网探测默认使用icmp扫描
		taskname = "Reserved interIP addresses"
	} else {
		config = IpInit(config)
		if config.IP != "" {
			taskname = config.IP
		} else if config.List != "" {
			taskname = config.List
		}
	}
	if taskname == "" {
		fmt.Println("[-] No Task")
		os.Exit(0)
	}

	// 输出任务的基本信息
	processLog(fmt.Sprintf("[*] Start scan task %s ,total ports: %d , mod: %s", taskname, len(config.Portlist), config.Mod))
	if len(config.Portlist) > 1000 {
		fmt.Println("[*] too much ports , only show top 1000 ports: " + strings.Join(config.Portlist[:1000], ",") + "......")
	} else {
		fmt.Println("[*] ports: " + strings.Join(config.Portlist, ","))
	}
	if config.Mod == "default" {
		processLog(fmt.Sprintf("[*] Estimated to take about %d seconds", guesstime(config)))
	}

	switch config.Mod {
	case "default":
		StraightMod(config)
	case "a", "auto":
		config.Mod = "ss"
		config.IP = "10.0.0.0/8"
		processLog("[*] Spraying : 10.0.0.0/8")
		if config.SmartPort == "default" {
			config.SmartPortList = []string{"icmp"}
		}
		SmartMod(config)

		processLog("[*] Spraying : 172.16.0.0/12")
		config.IP = "172.16.0.0/12"
		SmartMod(config)

		processLog("[*] Spraying : 192.168.0.0/16")
		if config.SmartPort == "default" {
			config.SmartPortList = []string{"80"}
		}
		config.IP = "192.168.0.0/16"
		//config.Mod = "s"
		SmartMod(config)

	case "s", "f", "ss":
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

func PortHandler(portstring string) []string {
	var ports []string
	portstring = strings.Replace(portstring, "\r", "", -1)

	postslist := strings.Split(portstring, ",")
	for _, portname := range postslist {
		ports = append(ports, choiceports(portname)...)
	}
	ports = Utils.Ports2PortSlice(ports)
	ports = Utils.SliceUnique(ports)
	return ports
}

// 端口预设
func choiceports(portname string) []string {
	var ports []string
	if portname == "all" {
		for p := range Utils.Portmap {
			ports = append(ports, p)
		}
		return ports
	}

	if Utils.Namemap[portname] != nil {
		ports = append(ports, Utils.Namemap[portname]...)
		return ports
	} else if Utils.Typemap[portname] != nil {
		ports = append(ports, Utils.Typemap[portname]...)
		return ports
	} else {
		return []string{portname}
	}
}

func Printportconfig() {
	fmt.Println("当前已有端口配置: (根据端口类型分类)")
	for k, v := range Utils.Namemap {
		fmt.Println("	", k, ": ", strings.Join(v, ","))
	}
	fmt.Println("当前已有端口配置: (根据服务分类)")
	for k, v := range Utils.Typemap {
		fmt.Println("	", k, ": ", strings.Join(v, ","))
	}
}

func IpInit(config Config) Config {
	// 如果输入的是文件,则格式化所有输入值.如果无有效ip
	if config.List != "" {
		var iplist []string
		for _, ip := range config.IPlist {
			tmpip := IpForamt(ip)
			if !strings.HasPrefix(tmpip, "err") {
				iplist = append(iplist, tmpip)
			} else {
				fmt.Println("[-] " + tmpip + " ip format error")
			}
		}
		config.IPlist = Utils.SliceUnique(iplist) // 去重
		if len(config.IPlist) == 0 {
			fmt.Println("[-] all IP error")
			os.Exit(0)
		}
	} else if config.IP != "" {
		config.IP = IpForamt(config.IP)
		if strings.HasPrefix(config.IP, "err") {
			fmt.Println("[-] IP format error")
			os.Exit(0)
		}
	}
	return config
}

func IpForamt(target string) string {
	target = strings.Replace(target, "http://", "", -1)
	target = strings.Replace(target, "https://", "", -1)
	target = strings.Trim(target, "/")
	if strings.Contains(target, "/") {
		ip := strings.Split(target, "/")[0]
		mask := strings.Split(target, "/")[1]
		if isIPv4(ip) {
			target = ip + "/" + mask
		} else {
			target = getIp(ip) + "/" + mask
		}
	} else {
		if isIPv4(target) {
			target = target + "/32"
		} else {
			target = getIp(target) + "/32"
		}
	}
	return target
}

func getIp(target string) string {
	iprecords, err := net.LookupIP(target)
	if err != nil {
		fmt.Println("[-] error IPv4 or bad domain:" + target + ". JUMPED!")
		return "err"
	}
	for _, ip := range iprecords {
		if ip.To4() != nil {
			fmt.Println("[*] parse domain SUCCESS, map " + target + " to " + ip.String())
			return ip.String()
		}
	}
	return "err"
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
