package scheme

import "regexp"

/*
Name		- 监控策略名
FilePath	- 文件路径
TimeFormat	- 时间格式
Pattern		- 表达式
Exclude     - 排除表达式
Interval	- 采集周期
Tags		- Tags
Func		- 采集方式（max/min/avg/cnt）
Degree		- 精度位数
Comment		- 备注
*/

type Strategy struct {
	ID         int64                     `json:"id"`
	Name       string                    `json:"name"`
	FilePath   string                    `json:"file_path"`
	TimeFormat string                    `json:"time_format"`
	Pattern    string                    `json:"pattern"`
	Exclude    string                    `json:"exclude"`
	Interval   int64                     `json:"step"`
	Tags       map[string]string         `json:"tags"`
	Func       string                    `json:"func"`
	Degree     int64                     `json:"degree"`
	Comment    string                    `json:"comment"`
	TimeReg    *regexp.Regexp            `json:"-"`
	PatternReg *regexp.Regexp            `json:"-"`
	ExcludeReg *regexp.Regexp            `json:"-"`
	TagRegs    map[string]*regexp.Regexp `json:"-"`
	ParseSucc  bool                      `json:"parse_succ"`
}

type LimitResp struct {
	CpuNum int `json:"cpu_num"`
	MemMB  int `json:"mem_mb"`
}

func DeepCopyStrategy(p *Strategy) *Strategy {
	s := Strategy{}
	s.ID = p.ID
	s.Name = p.Name
	s.FilePath = p.FilePath
	s.TimeFormat = p.TimeFormat
	s.Pattern = p.Pattern
	s.Interval = p.Interval
	s.Tags = DeepCopyStringMap(p.Tags)
	s.Func = p.Func
	s.Degree = p.Degree
	s.Comment = p.Comment

	return &s
}

func DeepCopyStringMap(p map[string]string) map[string]string {
	r := make(map[string]string, len(p))
	for k, v := range p {
		r[k] = v
	}
	return r
}

func DeepCopyStringSlice(p []string) []string {
	r := make([]string, len(p))
	for i, v := range p {
		r[i] = v
	}
	return r
}
