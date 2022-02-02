package cmd

import (
	"flag"
	"fmt"
	. "getitle/src/core"
	. "getitle/src/scan"
	"github.com/panjf2000/ants/v2"
	"os"
	"strings"
)

var ver = ""
var k = "ybb"

func CMD() {
	defer ants.Release()
	connected = checkconn()
	if !strings.Contains(strings.Join(os.Args, ""), k) {
		inforev()
	}
	runner := NewRunner()
	//默认参数信息
	// INPUT
	flag.StringVar(&runner.config.IP, "ip", "", "")
	flag.StringVar(&runner.config.Ports, "p", "top1", "")
	flag.StringVar(&runner.config.ListFile, "l", "", "")
	flag.StringVar(&runner.config.JsonFile, "j", "", "")
	flag.BoolVar(&runner.config.IsListInput, "L", false, "")
	flag.BoolVar(&runner.config.IsJsonInput, "J", false, "")

	// SMART
	flag.StringVar(&runner.config.SmartPort, "sp", "default", "")
	flag.StringVar(&runner.config.IpProbe, "ipp", "default", "")
	flag.BoolVar(&runner.config.NoSpray, "ns", false, "")
	flag.BoolVar(&Opt.Noscan, "no", false, "")

	// OUTPUT
	flag.StringVar(&runner.config.Filename, "f", "", "")
	flag.StringVar(&Opt.FilePath, "path", "", "")
	flag.StringVar(&runner.config.ExcludeIPs, "eip", "", "")
	flag.StringVar(&Opt.Output, "o", "full", "")
	flag.BoolVar(&runner.Clean, "c", false, "")
	flag.StringVar(&Opt.FileOutput, "O", "json", "")
	flag.BoolVar(&runner.Quiet, "q", false, "")
	flag.Var(&runner.filters, "filter", "")
	flag.StringVar(&runner.FormatOutput, "F", "", "")
	flag.BoolVar(&runner.AutoFile, "af", false, "")
	flag.BoolVar(&runner.HiddenFile, "hf", false, "")
	flag.BoolVar(&runner.Compress, "C", false, "")

	// CONFIG
	flag.IntVar(&runner.config.Threads, "t", 0, "")
	flag.StringVar(&runner.config.Mod, "m", "default", "")
	flag.BoolVar(&runner.config.Spray, "s", false, "")
	flag.BoolVar(&runner.config.Ping, "ping", false, "")
	flag.StringVar(&runner.iface, "iface", "eth0", "")
	flag.BoolVar(&Opt.Debug, "debug", false, "")
	flag.IntVar(&RunOpt.Delay, "d", 2, "")
	flag.IntVar(&RunOpt.HttpsDelay, "D", 2, "")
	flag.StringVar(&RunOpt.Payloadstr, "suffix", "", "")
	flag.Var(&runner.payloads, "payload", "")
	flag.Var(&runner.extract, "extract", "")
	flag.StringVar(&runner.extracts, "extracts", "", "")
	flag.BoolVar(&runner.Version, "v", false, "")
	flag.BoolVar(&runner.Version2, "vv", false, "")
	flag.BoolVar(&runner.Exploit, "e", false, "")
	flag.StringVar(&runner.ExploitName, "E", "none", "")
	flag.StringVar(&runner.ExploitFile, "ef", "", "")

	// OTHER
	key := flag.String("k", "", "")
	flag.StringVar(&runner.Printer, "P", "", "")
	flag.BoolVar(&runner.NoUpload, "nu", false, "")
	flag.StringVar(&runner.UploadFile, "uf", "", "")
	flag.BoolVar(&runner.Ver, "version", false, "")

	flag.Usage = func() { exit() }
	flag.Parse()
	// 密钥
	if *key != k {
		//rev()
		exit()
	}

	ok := runner.preInit()
	if !ok {
		os.Exit(0)
	}
	runner.init()

	// 初始化任务
	Log.InitFile() // 在真正运行前再初始化进度文件
	runner.config = InitConfig(runner.config)
	RunTask(runner.config) // 运行

	runner.close()
}

type Value interface {
	String() string
	Set(string) error
}

type arrayFlags []string

// Value ...
func (i *arrayFlags) String() string {
	return fmt.Sprint(*i)
}

// Set 方法是flag.Value接口, 设置flag Value的方法.
// 通过多个flag指定的值， 所以我们追加到最终的数组上.
func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
