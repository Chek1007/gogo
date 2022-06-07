package scan

import (
	"getitle/v1/pkg/fingers"
	"github.com/alouca/gosnmp"
	"strings"

	//"encoding/hex"
	"getitle/v1/pkg"
)

func snmpScan(result *pkg.Result) {
	var err error
	result.Port = "161"
	s, err := gosnmp.NewGoSNMP(result.GetTarget(), "public", gosnmp.Version2c, int64(RunOpt.Delay+2))
	if err != nil {
		//log.Fatal(err)
		return
	}

	resp, err := s.Get(".1.3.6.1.2.1.1.1.0")
	if err != nil {
		return
	}

	result.Open = true
	result.Protocol = "snmp"
	result.HttpStat = "snmp"
	if len(resp.Variables) > 0 {
		result.AddVuln(&fingers.Vuln{Name: "snmp_default_auth", Payload: map[string]interface{}{"auth": "public"}, Severity: "info"})
		result.Title = strings.Split(resp.Variables[0].Value.(string), "#")[0]
	}
}
