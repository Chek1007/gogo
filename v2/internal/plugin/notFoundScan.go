package plugin

import (
	"github.com/chainreactors/gogo/v2/pkg"
	"github.com/chainreactors/gogo/v2/pkg/fingers"
	"github.com/chainreactors/logs"
	"github.com/chainreactors/parsers"
	"github.com/chainreactors/parsers/iutils"
	"strings"
)

func NotFoundScan(result *pkg.Result) {
	conn := result.GetHttpConn(RunOpt.Delay)
	url := result.GetURL() + pkg.RandomDir
	resp, err := conn.Get(url)

	if err != nil {
		logs.Log.Debugf("request 404page %s %s", url, err.Error())
		return
	}

	logs.Log.Debugf("request 404page %s %d", url, resp.StatusCode)
	if iutils.ToString(resp.StatusCode) == result.Status {
		return
	}
	content := string(parsers.ReadRaw(resp))
	if content == "" {
		return
	}

	for _, finger := range pkg.AllHttpFingers {
		framework, _, ok := fingers.FingerMatcher(finger, map[string]string{"content": strings.ToLower(content)}, 0, nil)
		if ok {
			framework.From = fingers.NOTFOUND
			result.AddFramework(framework)
		}
	}
}
