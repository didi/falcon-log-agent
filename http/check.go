package http

import (
	"regexp"

	"github.com/didi/falcon-log-agent/common/scheme"
	"github.com/didi/falcon-log-agent/common/utils"
	"github.com/didi/falcon-log-agent/strategy"
)

type MatchBody struct {
	Strategy *scheme.Strategy  `json:"strategy"`
	Detail   map[string]string `json:"detail"`
}
type CheckRet struct {
	Matched bool         `json:"matched"`
	Body    []*MatchBody `json:"body"`
}

func NewCheckRet() *CheckRet {
	return &CheckRet{Body: make([]*MatchBody, 0)}
}

func CheckLogByStrategy(content string) *CheckRet {
	var ret = NewCheckRet()

	sts := strategy.GetListAll()
	for idx, st := range sts {
		matched, detail := matchedStrategy(content, st)
		if matched {
			ret.Matched = true
			tmp := &MatchBody{
				Strategy: sts[idx],
				Detail:   detail,
			}
			ret.Body = append(ret.Body, tmp)
		}
	}
	return ret
}

func matchedStrategy(content string, strategy *scheme.Strategy) (bool, map[string]string) {
	var detail = make(map[string]string, 0)
	valid, patMap := getRegsFromOneStrategy(strategy)
	if !valid {
		return false, map[string]string{}
	}

	for key, pat := range patMap {
		reg, err := regexp.Compile(pat)
		if err != nil {
			return false, map[string]string{}
		}
		l := reg.FindStringSubmatch(content)
		if len(l) == 0 {
			if key != "exclude_" {
				return false, map[string]string{}
			}
			detail[key] = ""
			continue
		}
		detail[key] = l[0]
	}

	return true, detail
}

func getRegsFromOneStrategy(st *scheme.Strategy) (stValid bool, regs map[string]string) {
	var ret = make(map[string]string, 0)

	// pattern、exclude、time后带下划线，用来与重名tag区分
	// pattern字段必须有
	if st.Pattern == "" {
		return false, map[string]string{}
	}
	ret["pattern_"] = st.Pattern

	// exclude可以缺省
	if st.Exclude != "" {
		ret["exclude_"] = st.Exclude
	}

	// timeFormat必须有,且必须匹配到pat
	if st.TimeFormat == "" {
		return false, map[string]string{}
	} else {
		pat, _ := utils.GetPatAndTimeFormat(st.TimeFormat)
		if pat == "" {
			return false, map[string]string{}
		} else {
			ret["time_"] = pat
		}
	}

	// tags可以缺省
	for k, v := range st.Tags {
		ret[k] = v
	}

	return true, ret
}
