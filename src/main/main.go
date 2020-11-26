package main

import (
	"flag"
	"fmt"
	"getitle/src/Scan"
	"getitle/src/Utils"
	"getitle/src/core"
	"github.com/panjf2000/ants/v2"
	"os"
	"time"
)

func main() {
	defer ants.Release()

	//默认参数信息

	ports := flag.String("p", "top1", "")
	//ports := flag.String("port","80-8000","port")
	//key := flag.String("k", "", "")
	//ports := flag.String("port","21,22,23,25,443,444,445,464,465,468,487,488,496,500,512,513,514,515,517,518,519,520,521,525,526,530,531,532,533,535,538,540,543,544,546,547,548,554,556,563,565,587,610,611,612,616,631,636,674,694,749,750,751,752,754,760,765,767,808,871,873,901,953,992,993,994,995,1080,1109,1127,1178,1236,1300,1313,1433,1434,1494,1512,1524,1525,1529,1645,1646,1649,1701,1718,1719,1720,1758,1759,1789,1812,1813,1911,1985,1986,1997,2003,2049,2053,2102,2103,2104,2105,2150,2401,2430,2431,2432,2433,2600,2601,2602,2603,2604,2605,2606,2809,2988,3128,3130,3306,3346,3455,4011,4321,4444,4557,4559,5002,5232,5308,5354,5355,5432,5680,5999,6000,6010,6667,7000,7001,7002,7003,7004,7005,7006,7007,7008,7009,7100,7666,8008,8080,8081,9100,9359,9876,10080,10081,10082,10083,11371,11720,13720,13721,13722,13724,13782,13783,20011,20012,22273,22289,22305,22321,24554,26000,26208,27374,33434,60177,60179","max")
	list := flag.String("l", "", "")
	threads := flag.Int("t", 4000, "")
	IPaddress := flag.String("ip", "", "")
	mod := flag.String("m", "default", "")
	delay := flag.Int("d", 2, "")
	Output := flag.String("o", "full", "")
	Filename := flag.String("f", "", "")
	Exploit := flag.Bool("e", false, "")
	Version := flag.Bool("v", false, "")
	flag.Parse()

	if *IPaddress == "" && *list == "" {
		os.Exit(0)
	}

	starttime := time.Now()

	//初始化全局变量
	Scan.Delay = time.Duration(*delay)
	core.Threads = *threads
	core.Filename = *Filename
	core.OutputType = *Output
	Scan.Exploit = *Exploit
	Utils.Version = *Version
	core.Init()
	if *list != "" {
		targetList := core.ReadTargetFile(*list)
		for _, v := range targetList {
			CIDR, portlist, m := core.TargetHandler(v)
			core.RunTask(core.IpInit(CIDR), portlist, m)
		}
	} else {
		CIDR := core.IpInit(*IPaddress)
		portlist := core.PortHandler(*ports)
		core.RunTask(CIDR, portlist, *mod)
	}

	endtime := time.Since(starttime)

	//core.Datach <- sum
	if *Output == "json" {
		_, _ = core.FileHandle.WriteString(core.JsonOutput(*new(Utils.Result)) + "]")
	}
	time.Sleep(time.Microsecond * 500)
	println(fmt.Sprintf("[*] Alive sum: %d, Target sum : %d", Scan.Alivesum, Scan.Sum))
	println("[*] Totally run: " + endtime.String())

}
