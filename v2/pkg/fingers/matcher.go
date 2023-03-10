package fingers

import (
	"github.com/chainreactors/parsers"
	"regexp"
	"strings"
)

func compiledMatch(reg *regexp.Regexp, s string) (string, bool) {
	matched := reg.FindStringSubmatch(s)
	if matched == nil {
		return "", false
	}
	if len(matched) == 1 {
		return "", true
	} else {
		return strings.TrimSpace(matched[1]), true
	}
}

func compiledAllMatch(reg *regexp.Regexp, s string) ([]string, bool) {
	matchedes := reg.FindAllString(s, -1)
	if matchedes == nil {
		return nil, false
	}
	return matchedes, true
}

func FingerMatcher(finger *Finger, content map[string]string, level int, sender func([]byte) (string, bool)) (*parsers.Framework, *parsers.Vuln, bool) {
	return finger.Match(content, level, sender)
}

func RuleMatcher(rule *Rule, content map[string]string, ishttp bool) (bool, bool, string) {
	var hasFrame, hasVuln bool
	var version string
	if rule.Regexps == nil {
		return false, false, ""
	}

	hasFrame, hasVuln, version = rule.Match(content["content"], ishttp)
	if hasFrame || !ishttp {
		return hasFrame, hasVuln, version
	}

	hasFrame = rule.MatchCert(content["cert"])
	return hasFrame, hasVuln, version
}
