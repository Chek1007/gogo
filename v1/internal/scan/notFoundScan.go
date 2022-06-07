package scan

import (
	"getitle/v1/pkg"
	"getitle/v1/pkg/fingers"
	"getitle/v1/pkg/utils"
)

func NotFoundScan(result *pkg.Result) {
	conn := result.GetHttpConn(RunOpt.Delay)
	url := result.GetURL() + pkg.RandomDir
	resp, err := conn.Get(url)

	if err != nil {
		pkg.Log.Debugf("request 404page %s %s", url, err.Error())
		return
	}
	pkg.Log.Debugf("request 404page %s %d", url, resp.StatusCode)
	if utils.ToString(resp.StatusCode) == result.HttpStat {
		return
	}
	content := string(pkg.GetBody(resp))
	if content == "" {
		return
	}

	for _, finger := range pkg.AllFingers {
		framework, _, ok := fingers.FingerMatcher(finger, content)
		if ok {
			framework.From = "404"
			result.AddFramework(framework)
		}
	}
}
