package core

import (
	"encoding/json"
	"fmt"
	"getitle/src/structutils"
	. "getitle/src/utils"
	"strings"
)

func output(result *Result, outType string) string {
	var out string

	switch outType {
	case "color", "c":
		out = colorOutput(result)
	case "json", "j":
		out = jsonOutput(result)
	//case "html":
	//	out = HtmlOutput(result)
	default:
		out = fullOutput(result)

	}
	return out
}

func colorOutput(result *Result) string {
	s := fmt.Sprintf("[+] %s://%s:%s\t%s\t%s\t%s\t%s\t%s [%s] %s %s\n", result.Protocol, result.Ip, result.Port, result.Midware, result.Language, Blue(result.Frameworks.ToString()), result.Host, result.Hash, Yellow(result.HttpStat), Blue(result.Title), Red(result.Vulns.ToString()))
	return s
}

func fullOutput(result *Result) string {
	s := fmt.Sprintf("[+] %s://%s:%s%s\t%s\t%s\t%s\t%s\t%s [%s] %s %s\n", result.Protocol, result.Ip, result.Port, result.Uri, result.Midware, result.Language, result.Frameworks.ToString(), result.Host, result.Hash, result.HttpStat, result.Title, result.Vulns.ToString())
	return s
}

func jsonOutput(result *Result) string {
	jsons, _ := json.Marshal(result)
	return string(jsons)
}

func FormatOutput(filename string, outputfile string, autofile bool) {
	var outfunc func(s string)
	var iscolor bool
	var resultsdata ResultsData
	var smartdata SmartData

	data := LoadResultFile(filename)
	switch data.(type) {
	case ResultsData:
		resultsdata = data.(ResultsData)
		fmt.Println(resultsdata.ToConfig())
		if outputfile == "" {
			outputfile = GetFilename(resultsdata.Config, autofile, false, Output)
		}
	case SmartData:
		smartdata = data.(SmartData)
		if outputfile == "" {
			outputfile = GetFilename(smartdata.Config, autofile, false, "cidr")
		}
	default:
		return
	}

	if outputfile != "" {
		fileHandle := InitFileHandle(outputfile)
		fmt.Println("[*] Output filename: " + outputfile)
		defer fileHandle.Close()
		outfunc = func(s string) {
			_, _ = fileHandle.WriteString(s)
		}
	} else {
		outfunc = func(s string) {
			fmt.Print(s)
		}
	}

	if smartdata.Data != nil {
		outfunc(strings.Join(smartdata.Data, "\n"))
		return
	}

	if Output == "c" {
		iscolor = true
	}

	if Output == "cs" {
		outfunc(resultsdata.ToCobaltStrike())
		return
	} else if Output == "zombie" {
		outfunc(resultsdata.ToZombie())
		return
	} else if Output == "c" || Output == "full" {
		outfunc(resultsdata.ToConfig())
		outfunc(resultsdata.ToFormat(iscolor))
	} else {
		outfunc(resultsdata.GetValue(Output))
	}
}

func progressLogln(s string) {
	s = s + " , " + structutils.GetCurtime() + "\n"
	fmt.Print(s)
	if LogFileHandle != nil {
		LogDetach <- s
	}
}

func consoleLog(s string) {

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

func Printportconfig() {
	fmt.Println("当前已有端口配置: (根据端口类型分类)")
	for k, v := range Namemap {
		fmt.Println("	", k, ": ", strings.Join(v, ","))
	}
	fmt.Println("当前已有端口配置: (根据服务分类)")
	for k, v := range Tagmap {
		fmt.Println("	", k, ": ", strings.Join(v, ","))
	}
}

func PrintNucleiPoc() {
	fmt.Println("Nuclei Pocs")
	for k, v := range TemplateMap {
		fmt.Println(k + ":")
		for _, t := range v {
			fmt.Println("\t" + t.Info.Name)
		}

	}
}

func PrintInterConfig() {
	fmt.Println("Auto internet smart scan config")
	fmt.Println("CIDR\t\tMOD\tPortProbe\tIpProbe")
	for k, v := range InterConfig {
		fmt.Printf("%s\t\t%s\n", k, strings.Join(v, "\t"))
	}
}
