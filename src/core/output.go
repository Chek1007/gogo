package core

import (
	"encoding/json"
	"fmt"
	"getitle/src/Utils"
	"io/ioutil"
	"sort"
)

var first = true

type portformat struct {
	Port      string `json:"p"`
	Stat      string `json:"s"`
	Title     string `json:"t"`
	Host      string `json:"h"`
	Midware   string `json:"m"`
	Language  string `json:"l"`
	Framework string `json:"f"`
	Vuln      string `json:"v"`
	Protocol  string `json:"r"`
}

func output(result *Utils.Result, outType string) string {
	var out string

	switch outType {
	case "color", "c":
		out = ColorOutput(result)
	case "json", "j":
		if FileHandle != nil {
			out = JsonFile(result)
		} else {
			out = JsonOutput(result)
		}
	//case "html":
	//	out = HtmlOutput(result)
	default:
		out = FullOutput(result)

	}
	return out

}

func ColorOutput(result *Utils.Result) string {
	s := fmt.Sprintf("[+] %s://%s:%s\t%s\t%s\t%s\t%s\t[%s] %s ", result.Protocol, result.Ip, result.Port, result.Midware, result.Language, blue(result.Framework), result.Host, yellow(result.HttpStat), blue(result.Title))
	s += red(vulnOutput(result.Vuln))
	s += "\n"
	return s
}

func FullOutput(result *Utils.Result) string {
	s := fmt.Sprintf("[+] %s://%s:%s%s\t%s\t%s\t%s\t%s\t[%s] %s ", result.Protocol, result.Ip, result.Port, result.Uri, result.Midware, result.Language, result.Framework, result.Host, result.HttpStat, result.Title)
	s += vulnOutput(result.Vuln)
	s += "\n"
	return s
}

func JsonOutput(result *Utils.Result) string {
	jsons, err := json.Marshal(result)
	if err != nil {
		return ""
	}
	return string(jsons) + "\n"
}

func JsonFile(result *Utils.Result) string {
	jsons, err := json.Marshal(result)
	if err != nil {
		return ""
	}
	if first {
		first = false
		return string(jsons)
	} else {
		return "," + string(jsons)
	}

}

//func HtmlOutput(result *Utils.Result) (s string) {
//	if strings.HasPrefix(result.Protocol, "http") {
//		s = fmt.Sprintf("[+] <a>%s://%s:%s</a>\t%s\t%s\t%s\t%s\t[%s] %s", result.Protocol, result.Ip, result.Port, result.Midware, result.Language, result.Framework, result.Host, result.HttpStat, result.Title)
//	} else {
//		s = fmt.Sprintf("[+] %s://%s:%s\t%s\t%s\t%s\t%s\t[%s] %s", result.Protocol, result.Ip, result.Port, result.Midware, result.Language, result.Framework, result.Host, result.HttpStat, result.Title)
//	}
//	vulnstr := vulnOutput(result.Vuln)
//	if vulnstr != "" {
//		s += "<b style=\"color:red;\">" + vulnstr + "</b>"
//	}
//	s += "\n"
//
//	return s
//
//}

func vulnOutput(vuln string) string {
	if vuln != "" {
		return fmt.Sprintf("[ Find Vuln: %s ]", vuln)
	}
	return ""

}

func FormatOutput(filename string, outputfile string) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		print(err.Error())
	}
	var results []Utils.Result
	err = json.Unmarshal(content, &results)
	if err != nil {
		fmt.Println("json error")
	}
	pfs := make(map[string]map[string]portformat)
	//ipfs := make(map[string]ipformat)
	for _, result := range results {
		pf := portformat{
			Port:      result.Port,
			Stat:      result.HttpStat,
			Title:     result.Title,
			Host:      result.Host,
			Midware:   result.Midware,
			Language:  result.Language,
			Framework: result.Framework,
			Vuln:      result.Vuln,
			Protocol:  result.Protocol,
		}
		if pfs[result.Ip] == nil {
			pfs[result.Ip] = make(map[string]portformat)
		}
		pfs[result.Ip][result.Port] = pf

	}

	// 排序
	var keys []int
	for ip, _ := range pfs {
		keys = append(keys, int(ip2int(ip)))
	}
	sort.Ints(keys)

	for _, ipi := range keys {
		ip := int2ip(uint(ipi))

		//if len(pfs[ip]) == 1 {
		//	for pint, p := range pfs[ip] {
		//		fmt.Printf("[+] %s://%s:%s\t%s\t%s\t%s\t%s\t[%s] %s\n", p.protocol, ip, pint, p.midware, p.language, p.framework, p.host, p.stat, p.title)
		//		continue
		//	}
		//}
		var hostname, network, netbiosstat string

		if _, k := pfs[ip]["135"]; k {
			hostname = pfs[ip]["135"].Host
			network = pfs[ip]["135"].Title
		}
		if _, k := pfs[ip]["137"]; k {
			hostname = pfs[ip]["137"].Host
			netbiosstat = pfs[ip]["137"].Stat
		}
		s := fmt.Sprintf("[+] %s %s %s %s\n", ip, hostname, netbiosstat, network)
		for pint, p := range pfs[ip] {
			// 跳过OXID与NetBois
			if !(p.Port == "135" || p.Port == "137" || p.Port == "icmp") {
				if Output == "c" {
					// 颜色输出
					s += fmt.Sprintf("\t%s://%s:%s\t%s\t%s\t%s\t%s\t[%s] %s", p.Protocol, ip, pint, p.Midware, p.Language, blue(p.Framework), p.Host, yellow(p.Stat), blue(p.Title))
					s += red(vulnOutput(p.Vuln))
				} else {
					s += fmt.Sprintf("\t%s://%s:%s\t%s\t%s\t%s\t%s\t[%s] %s", p.Protocol, ip, pint, p.Midware, p.Language, p.Framework, p.Host, p.Stat, p.Title)
					s += vulnOutput(p.Vuln)
				}
				s += "\n"
			}
		}
		fmt.Print(s)
	}
	//fmt.Println(string(content))
}

func processLog(s string) {
	s = s + " , " + Utils.GetCurtime() + "\n"
	fmt.Print(s)
	if LogFileHandle != nil {
		LogDetach <- s
	}
}

func Banner() {
	//fmt.Println(
	//	"Usage of ./main:" +
	//		"\n  -d int			超时,默认2s (default 2)  " +
	//		"\n  -ip string		IP地址 like 192.168.1.1/24" +
	//		"\n  -m string        扫描模式：default or s(smart)" +
	//		"\n  -p string        ports (default \"top1\")" +
	//		"\n     ports preset:   top1(default) 80,81,88,443,8080,7001,9001,8081,8000,8443" +
	//		"\n                     top2 80-90,443,7000-7009,9000-9009,8080-8090,8000-8009,8443,7080,8070,9080,8888,7777,9999,9090,800,801,808,5555,10080" +
	//		"\n                     db 3306,1433,1521,5432,6379,11211,27017" +
	//		"\n                     rce 1090,1098,1099,4444,11099,47001,47002,10999,45000,45001,8686,9012,50500,4848,11111,4445,4786,5555,5556" +
	//		"\n                     win 53,88,135,139,389,445,3389,5985" +
	//		"\n                     brute 21,22,389,445,1433,1521,3306,3389,5901,5432,6379,11211,27017" +
	//		"\n                     all 21,22,23,25,53,69,80,81-89,110,135,139,143,443,445,465,993,995,1080,1158,1433,1521,1863,2100,3128,3306,3389,7001,8080,8081-8088,8888,9080,9090,5900,1090,1099,7002,8161,9043,50000,50070,389,5432,5984,9200,11211,27017,161,873,1833,2049,2181,2375,6000,6666,6667,7777,6868,9000,9001,12345,5632,9081,3700,4848,1352,8069,9300" +
	//		"\n  -t int        threads (default 4000)" +
	//		"\n  -o string     输出格式:clean,full(default) or json\n" +
	//		"\n     example:           ./main -ip 192.168.1.1 -p top2" +
	//		"\n     smart mod example: ./main -ip 192.168.1.1/8 -p top2 -m s",
	//)

}
func red(s string) string {
	return "\033[1;31m" + s + "\033[0m"
}

func green(s string) string {
	return "\033[1;32m" + s + "\033[0m"
}

func yellow(s string) string {
	return "\033[4;33m" + s + "\033[0m"
}

func blue(s string) string {
	return "\033[1;34m" + s + "\033[0m"
}
