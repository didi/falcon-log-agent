package sample_log

import (
	"common/dlog"
	"sync"
	"time"
)

type SampleLog struct {
	sync.RWMutex
	Sample map[string]int64
}

var ErrorLog SampleLog

func init() {
	ErrorLog = SampleLog{}
	ErrorLog.Sample = make(map[string]int64, 0)
	dlog.Info("sample log start")
	SampleLoop()
}

func SampleLoop() {
	go func() {
		for {
			// 每1s强制落盘
			ErrorLog.ForceFlush()
			time.Sleep(1 * time.Second)
		}
	}()
}

func (s *SampleLog) Get(logContent string) int64 {
	s.RLock()
	ret, ok := s.Sample[logContent]
	s.RUnlock()
	if !ok {
		ret = 0
	}
	return ret
}

func (s *SampleLog) Input(logContent string) {
	num := s.Get(logContent)
	s.Lock()
	num, ok := s.Sample[logContent]
	if !ok {
		s.Sample[logContent] = 1
	} else {
		s.Sample[logContent] = num + 1
	}
	s.Unlock()
}

func (s *SampleLog) RemoveKey(logContent string) {
	s.Lock()
	delete(s.Sample, logContent)
	s.Unlock()
}

func (s *SampleLog) Keys() []string {
	ret := make([]string, 0)
	s.RLock()
	for k, _ := range s.Sample {
		ret = append(ret, k)
	}
	s.RUnlock()
	return ret
}

func (s *SampleLog) ForceFlushKey(logContent string) {
	num := s.Get(logContent)
	if num > 0 {
		dlog.Error(logContent, ". log_num : ", num)
		s.RemoveKey(logContent)
	}
}

func (s *SampleLog) ForceFlush() {
	s.Lock()
	for c, num := range s.Sample {
		dlog.Error(c, ". log_num : ", num)
	}
	s.Sample = make(map[string]int64, 0)
	s.Unlock()
}

func Error(content string) {
	ErrorLog.Input(content)
}
