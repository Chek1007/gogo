package core

import (
	"encoding/json"
	"fmt"
	. "getitle/src/utils"
	"os"
	"strings"
)

func output(result *Result, outType string) string {
	var out string

	switch outType {
	case "color", "c":
		out = colorOutput(result)
	case "json", "j":
		out = jsonOutput(result)
	case "full":
		out = fullOutput(result)
	default:
		out = valuesOutput(result, outType)

	}
	return out
}

func valuesOutput(result *Result, outType string) string {
	outs := strings.Split(outType, ",")
	for i, out := range outs {
		outs[i] = result.Get(out)
	}
	return strings.Join(outs, "\t") + "\n"
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

func FormatOutput(filename string, outputfile string, autofile bool, filters []string) {
	var outfunc func(s string)
	var iscolor bool
	var resultsdata ResultsData
	var smartdata SmartData
	var textdata string
	var file *os.File
	if filename == "stdin" {
		file = os.Stdin
	} else {
		file = Open(filename)
	}

	data := LoadResultFile(file)
	switch data.(type) {
	case ResultsData:
		resultsdata = data.(ResultsData)
		ConsoleLog(resultsdata.ToConfig())
		if outputfile == "" {
			outputfile = GetFilename(resultsdata.Config, autofile, false, Opt.Output)
		}
	case SmartData:
		smartdata = data.(SmartData)
		if outputfile == "" {
			outputfile = GetFilename(smartdata.Config, autofile, false, "cidr")
		}
	case []byte:
		textdata = string(data.([]byte))
	default:
		return
	}

	if outputfile != "" {
		fileHandle, err := NewFile(outputfile, Opt.Compress)
		if err != nil {
			fmt.Println("[-] " + err.Error())
			os.Exit(0)
		}
		fmt.Println("[*] Output filename: " + outputfile)
		defer fileHandle.close()
		outfunc = func(s string) {
			fileHandle.write(s)
		}
	} else {
		outfunc = func(s string) {
			fmt.Print(s)
		}
	}

	if Opt.Output == "c" {
		iscolor = true
	}

	if smartdata.Data != nil {
		outfunc(strings.Join(smartdata.Data, "\n"))
		return
	}

	if resultsdata.Data != nil {
		for _, filter := range filters {
			if strings.Contains(filter, "::") {
				kv := strings.Split(filter, "::")
				resultsdata.Data = resultsdata.Data.Filter(kv[0], kv[1], "::")
			} else if strings.Contains(filter, "==") {
				kv := strings.Split(filter, "==")
				resultsdata.Data = resultsdata.Data.Filter(kv[0], kv[1], "==")
			}
		}

		if Opt.Output == "cs" {
			outfunc(resultsdata.ToCobaltStrike())
		} else if Opt.Output == "zombie" {
			outfunc(resultsdata.ToZombie())
		} else if Opt.Output == "c" || Opt.Output == "full" {
			outfunc(resultsdata.ToFormat(iscolor))
		} else if Opt.Output == "json" {
			content, err := json.Marshal(resultsdata)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			outfunc(string(content))
		} else {
			outfunc(resultsdata.ToValues(Opt.Output))
		}
	}
	if textdata != "" {
		outfunc(textdata)
	}
}

func progressLogln(s string) {
	s = fmt.Sprintf("%s , %s", s, GetCurtime())
	if !Opt.Quiet {
		// 如果指定了-q参数,则不在命令行输出进度
		fmt.Println(s)
		return
	}

	if Opt.logFile != nil {
		Opt.LogDataCh <- s
	}
}

func ConsoleLog(s string) {
	if !Opt.Quiet {
		fmt.Println(s)
	}
}

func Banner() {
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
