package worker

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"time"

	"github.com/didi/falcon-log-agent/strategy"

	"github.com/didi/falcon-log-agent/common/dlog"
	"github.com/didi/falcon-log-agent/common/g"
	"github.com/didi/falcon-log-agent/common/proc/metric"
	"github.com/didi/falcon-log-agent/common/sample_log"
	"github.com/didi/falcon-log-agent/common/scheme"
	"github.com/didi/falcon-log-agent/common/utils"
)

//单个worker对象
type Worker struct {
	FilePath  string
	Counter   int64
	LatestTms int64 //正在处理的单条日志时间
	Close     chan struct{}
	Stream    chan string
	Mark      string //标记该worker信息，方便打log及上报自监控指标, 追查问题
	Analyzing bool   //标记当前Worker状态是否在分析中,还是空闲状态
}

//worker组
type WorkerGroup struct {
	WorkerNum          int
	LatestTms          int64 //保留字段
	Workers            []*Worker
	TimeFormatStrategy string
}

func (this WorkerGroup) GetOldestTms() (tms int64, allFree bool) {
	allFree = true
	var analysingOldestTms int64
	var freeNewestTms int64

	for _, w := range this.Workers {
		if w.LatestTms > freeNewestTms {
			freeNewestTms = w.LatestTms
		}

		if w.LatestTms >= 0 && w.Analyzing == true {
			allFree = false
			if analysingOldestTms == 0 {
				analysingOldestTms = w.LatestTms
			} else if analysingOldestTms > w.LatestTms {
				analysingOldestTms = w.LatestTms
			}
		}
	}

	if allFree {
		tms = freeNewestTms
	} else {
		tms = analysingOldestTms
	}

	return tms, allFree
}

/*
 * filepath和stream依赖外部，其他的都自己创建
 */
func NewWorkerGroup(filePath string, stream chan string, st *scheme.Strategy) *WorkerGroup {

	wg := &WorkerGroup{
		WorkerNum: g.Conf().Worker.WorkerNum,
		Workers:   make([]*Worker, 0),
	}

	dlog.Infof("new worker group, [file:%s][worker_num:%d]", filePath, g.Conf().Worker.WorkerNum)

	for i := 0; i < wg.WorkerNum; i++ {
		mark := fmt.Sprintf("[worker][file:%s][num:%d][id:%d]", filePath, g.Conf().Worker.WorkerNum, i)
		w := Worker{}
		w.Close = make(chan struct{})
		// w.ParentGroup = wg
		w.FilePath = filePath
		w.Stream = stream
		w.Mark = mark
		w.Analyzing = false
		w.Counter = 0
		wg.Workers = append(wg.Workers, &w)
	}

	return wg
}

func (wg *WorkerGroup) Start() {
	for _, worker := range wg.Workers {
		worker.Start()
	}
}

func (wg *WorkerGroup) Stop() {
	for _, worker := range wg.Workers {
		worker.Stop()
	}
}

func (w *Worker) Start() {
	go func() {
		w.Work()
	}()
}

func (w *Worker) Stop() {
	close(w.Close)
}

func (w *Worker) Work() {
	defer func() {
		if reason := recover(); reason != nil {
			dlog.Infof("%s -- worker quit: panic reason: %v", w.Mark, reason)
		} else {
			dlog.Infof("%s -- worker quit: normally", w.Mark)
		}
	}()
	dlog.Infof("worker starting...[%s]", w.Mark)

	var anaCnt, anaSwp int64
	analysClose := make(chan int, 0)

	go func() {
		for {
			//休眠10s
			select {
			case <-analysClose:
				return
			case <-time.After(time.Second * 10):
			}
			a := anaCnt
			metric.MetricAnalysis(w.FilePath, a-anaSwp)
			anaSwp = a
		}
	}()

	for {
		select {
		case line := <-w.Stream:
			w.Analyzing = true
			anaCnt = anaCnt + 1
			w.analysis(line)
			w.Analyzing = false
		case <-w.Close:
			analysClose <- 0
			return
		}

	}
}

//内部的分析方法
//轮全局的规则列表
//单次遍历
func (w *Worker) analysis(line string) {
	defer func() {
		if err := recover(); err != nil {
			dlog.Infof("%s[analysis panic] : %v", w.Mark, err)
		}
	}()

	sts := strategy.GetAll()
	for _, strategy := range sts {
		if strategy.FilePath == w.FilePath && strategy.ParseSucc {
			analyspoint, err := w.producer(line, strategy)

			if err != nil {
				log := fmt.Sprintf("%s[producer error][sid:%d] : %v", w.Mark, strategy.ID, err)
				sample_log.Error(log)
				continue
			} else {
				if analyspoint != nil {
					metric.MetricAnalysisSucc(w.FilePath, 1)
					toCounter(analyspoint, w.Mark)
				}
			}
		}
	}
}

func (w *Worker) producer(line string, strategy *scheme.Strategy) (*AnalysPoint, error) {
	defer func() {
		if err := recover(); err != nil {
			dlog.Errorf("%s[producer panic] : %v", w.Mark, err)
		}
	}()

	var reg *regexp.Regexp
	_, timeFormat := utils.GetPatAndTimeFormat(strategy.TimeFormat)

	reg = strategy.TimeReg

	t := reg.FindString(line)
	if len(t) <= 0 {
		return nil, errors.New(fmt.Sprintf("cannot get timestamp:[sname:%s][sid:%d][timeFormat:%v]", strategy.Name, strategy.ID, timeFormat))
	}

	// 如果没有年，需添加当前年
	// 需干掉内部的多于空格, 如Dec  7,有的有一个空格，有的有两个，这里统一替换成一个
	if timeFormat == "Jan 2 15:04:05" {
		timeFormat = fmt.Sprintf("2006 %s", timeFormat)
		t = fmt.Sprintf("%d %s", time.Now().Year(), t)
		reg := regexp.MustCompile(`\s+`)
		rep := " "
		t = reg.ReplaceAllString(t, rep)
	}

	// [风险]统一使用东八区
	loc, err := time.LoadLocation("Asia/Shanghai")
	tms, err := time.ParseInLocation(timeFormat, t, loc)
	if err != nil {
		return nil, err
	}
	//
	w.LatestTms = tms.Unix()

	//处理用户正则
	var patternReg, excludeReg *regexp.Regexp
	var value float64
	patternReg = strategy.PatternReg
	if patternReg != nil {
		v := patternReg.FindStringSubmatch(line)
		var vString string
		if v != nil && len(v) != 0 {
			if len(v) > 1 {
				vString = v[1]
			} else {
				vString = ""
			}
			value, err = strconv.ParseFloat(vString, 64)
			if err != nil {
				value = math.NaN()
			}
		} else {
			//外边匹配err之后，要确保返回值不是nil再推送至counter
			//正则有表达式，没匹配到，直接返回
			return nil, nil
		}

	} else {
		value = math.NaN()
	}

	//处理exclude
	excludeReg = strategy.ExcludeReg
	if excludeReg != nil {
		v := excludeReg.FindStringSubmatch(line)
		if v != nil && len(v) != 0 {
			//匹配到exclude了，需要返回
			return nil, nil
		}
	}

	//处理tag 正则
	tag := map[string]string{}
	for tagk, tagv := range strategy.Tags {
		var regTag *regexp.Regexp
		regTag, ok := strategy.TagRegs[tagk]
		if !ok {
			dlog.Errorf("%s[get tag reg error][sid:%d][tagk:%s][tagv:%s]", w.Mark, strategy.ID, tagk, tagv)
			return nil, nil
		}
		t := regTag.FindStringSubmatch(line)
		if t != nil && len(t) > 1 {
			tag[tagk] = t[1]
		} else {
			return nil, nil
		}
	}

	ret := &AnalysPoint{
		StrategyID: strategy.ID,
		Value:      value,
		Tms:        tms.Unix(),
		Tags:       tag,
	}
	return ret, nil
}

//将解析数据给counter
func toCounter(analyspoint *AnalysPoint, mark string) {
	if err := PushToCount(analyspoint); err != nil {
		dlog.Errorf("%s push to counter error: %v", mark, err)
	}
}
