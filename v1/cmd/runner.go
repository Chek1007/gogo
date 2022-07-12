package cmd

import (
	"fmt"
	. "getitle/v1/internal/core"
	"getitle/v1/internal/scan"
	. "getitle/v1/pkg"
	. "getitle/v1/pkg/utils"
	. "github.com/chainreactors/logs"
	"net"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

func NewRunner() *Runner {
	return &Runner{
		config: Config{},
	}
}

type Runner struct {
	Ports             string
	Version           bool // version level1
	Version2          bool // version level2
	Exploit           bool // 启用漏洞扫描
	NoUpload          bool // 关闭文件回传
	Compress          bool // 启用压缩
	Clean             bool // 是否开启命令行输出扫描结果
	Quiet             bool // 是否开启命令行输出日志
	AutoFile          bool // 自动生成格式化文件名
	HiddenFile        bool // 启用自动隐藏文件
	Ping              bool
	Arp               bool
	FileOutput        string // 输出格式
	FilenameFormat    string // 文件名格式, clear, auto or hidden
	FormatterFilename string // 待格式化文件名
	filters           arrayFlags
	payloads          arrayFlags
	extract           arrayFlags
	extracts          string
	ExploitName       string // 指定漏扫poc名字
	ExploitFile       string // 指定漏扫文件
	Printer           string // 输出特定的预设
	UploadFile        string // 上传特定的文件名
	WorkFlowName      string
	Ver               bool // 输出版本号
	NoScan            bool
	IsWorkFlow        bool
	Debug             bool
	iface             string
	start             time.Time
	config            Config
}

func (r *Runner) preInit() bool {
	// 初始化日志工具"
	Log = NewLogger(r.Quiet, r.Debug)
	Log.LogFileName = ".sock.lock"
	Log.Init()
	legalFormat := []string{"url", "ip", "port", "frameworks", "framework", "vuln", "vulns", "protocol", "title", "target", "hash", "language", "host", "color", "c", "json", "j", "full", "jsonlines", "jl", "zombie"}
	if r.FileOutput != "default" {
		for _, form := range strings.Split(r.FileOutput, ",") {
			if !SliceContains(legalFormat, form) {
				Log.Warnf("illegal file output format: %s, Please use one or more of the following formats: %s", form, strings.Join(legalFormat, ", "))
			}
		}
	}

	if Opt.Output != "full" {
		for _, form := range strings.Split(Opt.Output, ",") {
			if !SliceContains(legalFormat, form) {
				Log.Warnf("illegal output format: %s, Please use one or more of the following formats: %s", form, strings.Join(legalFormat, ", "))
			}
		}
	}

	if r.AutoFile {
		r.FilenameFormat = "auto"
	} else if r.HiddenFile {
		r.FilenameFormat = "hidden"
	}

	// 一些特殊的分支, 这些分支将会直接退出程序
	if r.Ver {
		fmt.Println(ver)
		return false
	}

	if r.FormatterFilename != "" {
		FormatOutput(r.FormatterFilename, r.config.Filename, r.AutoFile, r.filters)
		return false
	}
	// 输出 config
	if r.Printer != "" {
		printConfigs(r.Printer)
		return false
	}

	//if r.UploadFile != "" {
	//	// 指定上传文件
	//	uploadfiles(strings.Split(r.UploadFile, ","))
	//	return false
	//}
	return true
}

func (r *Runner) init() {
	// 初始化各种全局变量
	// 初始化指纹优先级
	if r.Version {
		scan.RunOpt.VersionLevel = 1
	} else if r.Version2 {
		scan.RunOpt.VersionLevel = 2
	} else {
		scan.RunOpt.VersionLevel = 0
	}

	// 初始化漏洞
	if r.Exploit {
		scan.RunOpt.Exploit = "auto"
	} else {
		scan.RunOpt.Exploit = r.ExploitName
	}

	if r.NoScan {
		Opt.Noscan = r.NoScan
	}

	if r.Compress {
		Opt.Compress = !Opt.Compress
	}

	if r.Clean {
		Log.Clean = !Log.Clean
	}

	if !Win {
		if r.iface == "eth0" {
			Log.Warn("no interface name input, use default interface name: eth0")
		}
		var err error
		scan.RunOpt.Interface, err = net.InterfaceByName(r.iface)
		if err != nil {
			Log.Warn("interface error, " + err.Error())
			//Log.Warn("interface error, " + err.Error())
			//Log.Warn("interface error, " + err.Error())
		}
	}

	if r.extracts != "" {
		exts := strings.Split(r.extracts, ",")
		for _, extract := range exts {
			if reg, ok := PresetExtracts[extract]; ok {
				Extractors[extract] = reg
			}
		}
	}
	for _, extract := range r.extract {
		if reg, ok := PresetExtracts[extract]; ok {
			Extractors[extract] = reg
		} else {
			Extractors[extract] = CompileRegexp(extract)
		}
	}

	// 加载配置文件中的全局变量
	configLoader()
	nucleiLoader(r.ExploitFile, r.payloads)
	r.start = time.Now()
}

func (r *Runner) prepareConfig(config Config) *Config {
	if r.Ports == "" {
		config.Ports = "top1"
	} else {
		config.Ports = r.Ports
	}

	if r.Arp {
		config.AliveSprayMod = append(config.AliveSprayMod, "arp")
	}
	if r.Ping {
		config.AliveSprayMod = append(config.AliveSprayMod, "icmp")
	}

	if r.FileOutput == "default" {
		Opt.FileOutput = "json"
	} else {
		Opt.FileOutput = r.FileOutput
	}

	if config.Mod == "sc" {
		Opt.FileOutput = "raw"
	}

	if config.Filename == "" {
		config.Filename = GetFilename(&config, r.FilenameFormat, Opt.FilePath, Opt.FileOutput)
	} else {
		config.Filename = path.Join(Opt.FilePath, config.Filename)
	}

	if config.IsSmart() {
		if r.NoScan && !r.AutoFile && !r.HiddenFile {
			config.SmartFilename = config.Filename
		} else {
			config.SmartFilename = GetFilename(&config, r.FilenameFormat, Opt.FilePath, "cidr")
		}
	}

	if config.HasAlivedScan() {
		config.AlivedFilename = GetFilename(&config, r.FormatterFilename, Opt.FilePath, "alived")
	}
	return &config
}

func (r *Runner) run() {
	if r.WorkFlowName == "" && !r.IsWorkFlow {
		r.runWithCMD()
	} else {
		var workflowMap = WorkflowMap{}
		if r.IsWorkFlow {
			workflowMap["tmp"] = ParseWorkflowsFromInput(LoadFile(os.Stdin))
			r.WorkFlowName = "tmp"
		} else if IsExist(r.WorkFlowName) {
			workflowMap["tmp"] = ParseWorkflowsFromInput(LoadFile(Open(r.WorkFlowName)))
			r.WorkFlowName = "tmp"
		} else {
			workflowMap = LoadWorkFlow()
		}
		r.runWithWorkFlow(workflowMap)
	}
}

func (r *Runner) runWithCMD() {
	config := r.prepareConfig(r.config)
	RunTask(*InitConfig(config)) // 运行
	r.close(config)
}

func (r *Runner) runWithWorkFlow(workflowMap WorkflowMap) {
	if workflows := workflowMap.Choice(r.WorkFlowName); len(workflows) > 0 {
		for _, workflow := range workflows {
			Log.Important("workflow " + workflow.Name + " starting")

			if Opt.FilePath != "" {
				workflow.Path = Opt.FilePath
			}

			if workflow.Mod == "sc" {
				Opt.FileOutput = "raw"
			}

			// 一些workflow的参数, 允许被命令行参数覆盖
			if r.config.IP != "" {
				workflow.IP = r.config.IP
			}

			config := workflow.PrepareConfig()

			if r.config.Filename != "" {
				config.Filename = r.config.Filename
			} else if r.AutoFile {
				config.Filename = GetFilename(config, "auto", workflow.Path, "json")
				if config.IsSmartScan() {
					config.SmartFilename = GetFilename(config, "auto", workflow.Path, "cidr")
				}
				if config.HasAlivedScan() {
					config.AlivedFilename = GetFilename(config, "auto", workflow.Path, "alived")
				}
			} else if r.HiddenFile {
				workflow.File = GetFilename(config, "hidden", workflow.Path, "json")
				if config.IsSmartScan() {
					config.SmartFilename = GetFilename(config, "hidden", workflow.Path, "cidr")
				}
				if config.HasAlivedScan() {
					config.AlivedFilename = GetFilename(config, "hidden", workflow.Path, "alived")
				}
			}

			// 一些workflow的参数, 允许被命令行参数覆盖
			if r.config.IP != "" {
				config.IP = r.config.IP
			}

			if r.config.ListFile != "" {
				config.ListFile = r.config.ListFile
			}

			if r.Ports != "" {
				config.Ports = r.Ports
			}

			if r.config.Threads != 0 {
				config.Threads = r.config.Threads
			}

			if r.config.SmartPort != "default" {
				config.SmartPort = r.config.SmartPort
			}

			if r.config.IpProbe != "default" {
				config.IpProbe = r.config.IpProbe
			}

			// 全局变量的处理
			if !r.NoScan {
				Opt.Noscan = workflow.NoScan
			}

			if r.FileOutput == "default" && config.Mod != "sc" {
				Opt.FileOutput = "json"
			} else {
				Opt.FileOutput = r.FileOutput
			}

			if r.Version {
				scan.RunOpt.VersionLevel = 1
			} else {
				scan.RunOpt.VersionLevel = workflow.Version
			}

			if scan.RunOpt.Exploit != "none" {
				if r.Exploit {
					scan.RunOpt.Exploit = "auto"
				} else {
					scan.RunOpt.Exploit = r.ExploitName
				}
			} else {
				scan.RunOpt.Exploit = workflow.Exploit
			}

			config = InitConfig(config)
			RunTask(*config) // 运行
			r.close(config)
			r.resetGlobals()
		}
	} else {
		Fatal("not fount workflow " + r.WorkFlowName)
	}
}

func (r *Runner) close(config *Config) {
	Opt.Close() // 关闭result与extract写入管道
	if r.HiddenFile {
		Chtime(config.Filename)
		if config.SmartFilename != "" {
			Chtime(config.SmartFilename)
		}
	}

	// 任务统计
	Log.Importantf("Alive sum: %d, Target sum : %d", Opt.AliveSum, scan.RunOpt.Sum)
	Log.Important("Totally run: " + time.Since(r.start).String())

	// 输出文件名
	if Opt.File != nil && Opt.File.FileHandler != nil {
		Log.Importantf("Results filename: " + config.Filename)
	}
	if Opt.SmartFile != nil && Opt.SmartFile.FileHandler != nil {
		Log.Important("Smartscan result filename: " + config.SmartFilename)
	}
	if Opt.AliveFile != nil && Opt.AliveFile.FileHandler != nil {
		Log.Important("Pingscan result filename: " + config.AlivedFilename)
	}
	if IsExist(config.Filename + "_extract") {
		Log.Important("extractor result filename: " + config.Filename + "_extract")
	}

	// 扫描结果文件自动上传
	//if connected && !r.NoUpload { // 如果出网则自动上传结果到云服务器
	//	uploadfiles([]string{config.Filename, config.SmartFilename})
	//}

	Log.Close(true)
	time.Sleep(time.Microsecond * 1000) // 因为是异步的, 等待文件最后处理完成
}

func (r *Runner) resetGlobals() {
	Opt.Noscan = false
	scan.RunOpt.Exploit = "none"
	scan.RunOpt.VersionLevel = 0
}

func printConfigs(t string) {
	if t == "port" {
		TagMap, NameMap, PortMap = LoadPortConfig()
		Printportconfig()
	} else if t == "nuclei" {
		nucleiLoader("", arrayFlags{})
		PrintNucleiPoc()
	} else if t == "workflow" {
		PrintWorkflow()
	} else if t == "extract" {
		PrintExtract()
	} else {
		fmt.Println("choice port|nuclei|workflow|extract")
	}
}

func nucleiLoader(pocfile string, payloads arrayFlags) {
	ExecuterOptions = ParserCmdPayload(payloads)
	TemplateMap = LoadNuclei(pocfile)
}

func configLoader() {
	TagMap, NameMap, PortMap = LoadPortConfig()
	AllFingers = LoadFinger("http")
	Mmh3Fingers, Md5Fingers = LoadHashFinger(AllFingers)
	TcpFingers = LoadFinger("tcp").GroupByPort()
	HttpFingers = AllFingers.GroupByPort()
	CommonCompiled = map[string]*regexp.Regexp{
		"title":     CompileRegexp("(?Uis)<title>(.*)</title>"),
		"server":    CompileRegexp("(?i)Server: ([\x20-\x7e]+)"),
		"xpb":       CompileRegexp("(?i)X-Powered-By: ([\x20-\x7e]+)"),
		"sessionid": CompileRegexp("(?i) (.*SESS.*?ID)"),
	}
}
