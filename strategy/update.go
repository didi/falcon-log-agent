package strategy

import (
	"regexp"
	"strings"

	"github.com/didi/falcon-log-agent/common/dlog"
	"github.com/didi/falcon-log-agent/common/scheme"
	"github.com/didi/falcon-log-agent/common/utils"

	"time"
)

// PatternExcludePartition to separate pattern and exclude
const PatternExcludePartition = "```EXCLUDE```"

// Update to update strategy
func Update() error {
	markTms := time.Now().Unix()
	dlog.Infof("[%d]Update Strategy start", markTms)
	strategys, err := GetAllStrategies()
	parsePattern(strategys)
	updateRegs(strategys)

	if err != nil {
		dlog.Errorf("[%d]Get my Strategy error ! [msg:%v]", markTms, err)
		return err
	}
	dlog.Infof("[%d]Get my Strategy success, num : [%d]", markTms, len(strategys))

	err = UpdateGlobalStrategy(strategys)
	if err != nil {
		dlog.Errorf("[%d]Update Strategy cache error ! [msg:%v]", err)
		return err
	}
	dlog.Infof("[%d]Update Strategy end", markTms)
	return nil
}

func parsePattern(strategys []*scheme.Strategy) {
	for _, st := range strategys {
		patList := strings.Split(st.Pattern, PatternExcludePartition)

		if len(patList) == 1 {
			st.Pattern = strings.TrimSpace(st.Pattern)
			continue
		} else if len(patList) >= 2 {
			st.Pattern = strings.TrimSpace(patList[0])
			st.Exclude = strings.TrimSpace(patList[1])
			continue
		} else {
			dlog.Errorf("bad pattern to parse : [%s]", st.Pattern)
		}
	}
}

func updateRegs(strategys []*scheme.Strategy) {
	for _, st := range strategys {
		st.TagRegs = make(map[string]*regexp.Regexp, 0)
		st.ParseSucc = false

		//更新时间正则
		pat, _ := utils.GetPatAndTimeFormat(st.TimeFormat)
		reg, err := regexp.Compile(pat)
		if err != nil {
			dlog.Errorf("compile time regexp failed:[sid:%d][format:%s][pat:%s][err:%v]", st.ID, st.TimeFormat, pat, err)
			continue
		}
		st.TimeReg = reg

		if len(st.Pattern) == 0 && len(st.Exclude) == 0 {
			dlog.Errorf("pattern and exclude are all empty, sid:[%d]", st.ID)
			continue
		}

		//更新pattern
		if len(st.Pattern) != 0 {
			reg, err = regexp.Compile(st.Pattern)
			if err != nil {
				dlog.Errorf("compile pattern regexp failed:[sid:%d][pat:%s][err:%v]", st.ID, st.Pattern, err)
				continue
			}
			st.PatternReg = reg
		}

		//更新exclude
		if len(st.Exclude) != 0 {
			reg, err = regexp.Compile(st.Exclude)
			if err != nil {
				dlog.Errorf("compile exclude regexp failed:[sid:%d][pat:%s][err:%v]", st.ID, st.Exclude, err)
				continue
			}
			st.ExcludeReg = reg
		}

		//更新tags
		for tagk, tagv := range st.Tags {
			reg, err = regexp.Compile(tagv)
			if err != nil {
				dlog.Errorf("compile tag failed:[sid:%d][pat:%s][err:%v]", st.ID, st.Exclude, err)
				continue
			}
			st.TagRegs[tagk] = reg
		}
		st.ParseSucc = true
	}
}
