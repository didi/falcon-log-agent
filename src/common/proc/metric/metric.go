package metric

import (
	"fmt"
	"sync"
	"time"

	"common/dlog"
)

type MetricTags struct {
	sync.RWMutex
	Counters map[string]int64
}

func (m *MetricTags) HasKey(k string) bool {
	m.RLock()
	if _, ok := m.Counters[k]; ok {
		m.RUnlock()
		return true
	} else {
		m.RUnlock()
		return false
	}
}

func (m *MetricTags) AddCount(k string, v int64) {
	if _, ok := m.Counters[k]; !ok {
		m.Lock()
		if _, ok := m.Counters[k]; !ok {
			m.Counters[k] = 0
		}
		m.Unlock()
	}
	m.Counters[k] = m.Counters[k] + v
}

// 自监控结构体
type SelfMonitMetrics struct {
	MemUsedMB       int64       `json:"mem_used_mb"`
	ReadLineCnt     *MetricTags `json:"read_line_cnt"`
	DropLineCnt     *MetricTags `json:"drop_line_cnt"`
	AnalysisCnt     *MetricTags `json:"analysis_cnt"`
	AnalysisSuccCnt *MetricTags `json:"analysis_succ_cnt"`
	PushCnt         int64       `json:"push_cnt"`
	PushErrorCnt    int64       `json:"push_err_cnt"`
	PushLatency     int64       `json:"push_latency"`
	NewTms          int64
}

var (
	globalSelfMonit *SelfMonitMetrics = newSelfMonitMetrics()
)

func newSelfMonitMetrics() *SelfMonitMetrics {
	return &SelfMonitMetrics{
		MemUsedMB:       0,
		ReadLineCnt:     newMetricTags(),
		DropLineCnt:     newMetricTags(),
		AnalysisCnt:     newMetricTags(),
		AnalysisSuccCnt: newMetricTags(),
		PushCnt:         0,
		PushErrorCnt:    0,
		PushLatency:     0,
		NewTms:          time.Now().Unix(),
	}
}

func newMetricTags() *MetricTags {
	return &MetricTags{
		Counters: make(map[string]int64),
	}
}

func clearGlobalCnt() {
	globalSelfMonit = newSelfMonitMetrics()
}

// 将统计落实成一个个的监控点
// 此处只打印日志，若需要上报自监控指标，可以修改此方法
// TODO:此处只对齐至最近的时间点，统计并非十分精准
func HandleMetrics(step int64) {
	statSelfMonit := globalSelfMonit
	clearGlobalCnt()

	statTms := statSelfMonit.NewTms
	tms := statTms + (step - statTms%step)

	logFormat := fmt.Sprintf("self monit [metric:%%s][tms:%d][value:%%v]", tms)
	dlog.Debugf(logFormat, "log.agent.mem.used.mb", statSelfMonit.MemUsedMB)
	dlog.Debugf(logFormat, "log.agent.push.cnt", statSelfMonit.PushCnt)
	dlog.Debugf(logFormat, "log.agent.push.err.cnt", statSelfMonit.PushErrorCnt)
	dlog.Debugf(logFormat, "log.agent.read.line.cnt", statSelfMonit.ReadLineCnt)
	dlog.Debugf(logFormat, "log.agent.drop.line.cnt", statSelfMonit.DropLineCnt)
	dlog.Debugf(logFormat, "log.agent.analysis.cnt", statSelfMonit.AnalysisCnt)
	dlog.Debugf(logFormat, "log.agent.analysis.succ", statSelfMonit.AnalysisSuccCnt)

	if statSelfMonit.PushCnt != 0 {
		latency := statSelfMonit.PushLatency / statSelfMonit.PushCnt
		dlog.Debugf(logFormat, "log.agent.push.latency.avg", latency)
	}
}

func MetricMem(size int64) {
	globalSelfMonit.MemUsedMB = size
}

func MetricReadLine(file string, num int64) {
	globalSelfMonit.ReadLineCnt.AddCount(file, num)
}

func MetricDropLine(file string, num int64) {
	globalSelfMonit.DropLineCnt.AddCount(file, num)
}

func MetricAnalysis(file string, num int64) {
	globalSelfMonit.AnalysisCnt.AddCount(file, num)
}

func MetricAnalysisSucc(file string, num int64) {
	globalSelfMonit.AnalysisSuccCnt.AddCount(file, num)
}

func MetricPushCnt(num int64, succ bool) {
	globalSelfMonit.PushCnt = globalSelfMonit.PushCnt + num
	if !succ {
		globalSelfMonit.PushErrorCnt = globalSelfMonit.PushErrorCnt + num
	}
}

func MetricPushLatency(latency int64) {
	globalSelfMonit.PushLatency = globalSelfMonit.PushLatency + latency
}

func MetricLoop(step int64) {
	for {
		HandleMetrics(step)
		time.Sleep(time.Duration(step) * time.Second)
	}
}
