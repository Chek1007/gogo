package plugin

import (
	. "github.com/chainreactors/gogo/v2/pkg"
	"github.com/chainreactors/gogo/v2/pkg/fingers"
	"github.com/chainreactors/logs"
	"github.com/chainreactors/parsers"
	"net/http"
)

func httpFingerScan(result *Result) {
	passiveHttpMatch(result)
	if RunOpt.VersionLevel > 0 {
		activeHttpMatch(result)
	}
	return
}

func passiveHttpMatch(result *Result) {
	for _, f := range PassiveHttpFingers {
		frame, vuln, ok := f.Match(result.ContentMap(), 0, nil)
		if ok {
			if vuln != nil {
				result.AddVuln(vuln)
			}
			result.AddFramework(frame)
		} else {
			historyMatch(result, f)
		}
	}

}

func activeHttpMatch(result *Result) {
	var closureResp *http.Response
	sender := func(sendData []byte) ([]byte, bool) {
		conn := result.GetHttpConn(RunOpt.Delay)
		url := result.GetURL() + string(sendData)
		logs.Log.Debugf("active detect: %s", url)
		resp, err := conn.Get(url)
		if err == nil {
			closureResp = resp
			return parsers.ReadRaw(resp), true
		} else {
			return nil, false
		}
	}

	for _, f := range ActiveHttpFingers {
		// 当前gogo中最大指纹level为1, 因此如果调用了这个函数, 则认定为level1
		frame, vuln, ok := f.Match(result.ContentMap(), 1, sender)
		if ok {
			if vuln != nil {
				result.AddVuln(vuln)
			}
			result.AddFramework(frame)
			CollectHttpInfo(result, closureResp)
		} else {
			// 如果没有匹配到,则尝试使用history匹配
			historyMatch(result, f)
		}
	}
}

func historyMatch(result *Result, f *fingers.Finger) {
	for _, content := range result.Httpresp.History {
		frame, vuln, ok := f.Match(content.ContentMap(), 0, nil)
		if ok {
			if vuln != nil {
				result.AddVuln(vuln)
			}
			frame.From = 5
			result.AddFramework(frame)
		}
	}
}
