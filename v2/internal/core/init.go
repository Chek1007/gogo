package core

import (
	"fmt"
	. "github.com/chainreactors/gogo/v2/internal/plugin"
	. "github.com/chainreactors/gogo/v2/pkg"
	. "github.com/chainreactors/gogo/v2/pkg/utils"
	"github.com/chainreactors/ipcs"
	. "github.com/chainreactors/logs"
	"os"
	"strings"
)

var Opt = Options{
	AliveSum: 0,
	Noscan:   false,
}

func InitConfig(config *Config) (*Config, error) {
	err := config.Validate()
	if err != nil {
		return nil, err
	}
	// 初始化
	config.Exploit = RunOpt.Exploit
	config.VersionLevel = RunOpt.VersionLevel

	if config.Threads == 0 { // if 默认线程
		config.Threads = LinuxDefaultThreads
		if Win {
			//windows系统默认协程数为1000
			config.Threads = WindowsDefaultThreads
		} else {
			// linux系统判断fd限制, 如果-t 大于fd限制,则将-t 设置到fd-100
			if fdlimit := GetFdLimit(); config.Threads > fdlimit {
				Log.Warnf("System fd limit: %d , Please exec 'ulimit -n 65535'", fdlimit)
				Log.Warnf("Now set threads to %d", fdlimit-100)
				config.Threads = fdlimit - 100
			}
		}
	}

	var file *os.File
	if config.ListFile != "" {
		file = Open(config.ListFile)
	} else if config.JsonFile != "" {
		file = Open(config.JsonFile)
	} else if HasStdin() {
		file = os.Stdin
	}

	// 初始化文件操作
	err = config.InitFile()
	if err != nil {
		return nil, err
	}
	syncFile = func() {
		if config.File != nil {
			config.File.SafeSync()
		}
	}

	if config.ListFile != "" || config.IsListInput {
		// 如果从文件中读,初始化IP列表配置
		config.IPlist = strings.Split(string(LoadFile(file)), "\n")
	} else if config.JsonFile != "" || config.IsJsonInput {
		// 如果输入的json不为空,则从json中加载result,并返回结果
		data := LoadResultFile(file)
		switch data.(type) {
		case Results:
			config.Results = data.(Results)
		case *ResultsData:
			config.Results = data.(*ResultsData).Data
		case *SmartData:
			config.IPlist = data.(*SmartData).Data
		default:
			return nil, fmt.Errorf("not support result type, maybe use -l flag")
		}
	}

	err = config.InitIP()
	if err != nil {
		return nil, err
	}
	// 初始化端口配置
	config.PortList = ipcs.ParsePort(config.Ports)

	// 如果指定端口超过100,则自动启用spray
	if len(config.PortList) > 500 && !config.NoSpray {
		if config.CIDRs.Count() == 1 {
			config.PortSpray = false
		} else {
			config.PortSpray = true
		}
	}

	// 初始化启发式扫描的端口探针
	if config.PortProbe != Default {
		config.PortProbeList = ipcs.ParsePort(config.PortProbe)
		//} else {
		//	if config.Mod == SMART {
		//		config.PortProbeList = []string{DefaultSmartPortProbe}
		//	} else if SliceContains([]string{SUPERSMART, SUPERSMARTB}, config.Mod) {
		//		config.PortProbeList = []string{DefaultSuperSmartPortProbe}
		//	}
	}

	// 初始化ss模式ip探针,默认ss默认只探测ip为1的c段,可以通过-ipp参数指定,例如-ipp 1,254,253
	if config.IpProbe != Default {
		config.IpProbeList = Str2uintlist(config.IpProbe)
	} else {
		config.IpProbeList = Str2uintlist(DefaultIpProbe)
	}

	// 初始已完成,输出任务基本信息
	taskname := config.GetTargetName()
	// 输出任务的基本信息
	printTaskInfo(config, taskname)
	return config, nil
}

func printTaskInfo(config *Config, taskname string) {
	// 输出任务的基本信息
	Log.Importantf("Current goroutines: %d, Version Level: %d,Exploit: %s, PortSpray Scan: %t", config.Threads, RunOpt.VersionLevel, RunOpt.Exploit, config.PortSpray)
	if config.Results == nil {
		Log.Importantf("Starting task %s ,total ports: %d , mod: %s", taskname, len(config.PortList), config.Mod)
		// 输出端口信息
		if len(config.PortList) > 500 {
			Log.Important("too much ports , only show top 500 ports: " + strings.Join(config.PortList[:500], ",") + "......")
		} else {
			Log.Important("ports: " + strings.Join(config.PortList, ","))
		}
	} else {
		Log.Importantf("Starting results task: %s ,total target: %d", taskname, len(config.Results))
	}
}

func RunTask(config Config) {
	switch config.Mod {
	case Default:
		createDefaultScan(config)
	case SMART, SUPERSMART, SUPERSMARTB:
		if config.CIDRs != nil {
			for _, ip := range config.CIDRs {
				SmartMod(ip, config)
			}
		}
	default:
		createDefaultScan(config)
	}
}

func guessTime(targets interface{}, portcount, thread int) int {
	ipcount := 0

	switch targets.(type) {
	case ipcs.CIDRs:
		for _, cidr := range targets.(ipcs.CIDRs) {
			ipcount += int(cidr.Count())
		}
	case ipcs.CIDR:
		ipcount += int(targets.(ipcs.CIDR).Count())
	case Results:
		ipcount = len(targets.(Results))
		portcount = 1
	default:
	}

	return (portcount*ipcount/thread)*4 + 4
}

func guessSmartTime(cidr *ipcs.CIDR, config Config) int {
	var spc, ippc int
	var mask int
	spc = len(config.PortProbeList)
	if config.IsBSmart() {
		ippc = 1
	} else {
		ippc = len(config.IpProbeList)
	}
	mask = cidr.Mask

	var count int
	if config.Mod == SMART || config.Mod == SUPERSMARTC {
		count = 2 << uint((32-mask)-1)
	} else {
		count = 2 << uint((32-mask)-9)
	}

	return (spc*ippc*count)/(config.Threads)*2 + 2
}
