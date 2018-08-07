package worker

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// cached时间周期
const CACHED_DURATION = 60

type counterCache struct {
	sync.RWMutex
	Points map[int64]float64 `json:"points"`
}

type pushPointsCache struct {
	sync.RWMutex
	Counters map[string]*counterCache `json:"counters"`
}

var globalPushPoints = pushPointsCache{Counters: make(map[string]*counterCache, 0)}

func init() {
	go CleanLoop()
}

func (this *counterCache) AddPoint(tms int64, value float64) {
	this.Lock()
	this.Points[tms] = value
	this.Unlock()
}

func PostToCache(paramPoints []*FalconPoint) {
	for _, point := range paramPoints {
		globalPushPoints.AddPoint(point)
	}
}

func CleanLoop() {
	for {
		// 遍历，删掉过期的cache
		globalPushPoints.CleanOld()
		time.Sleep(5 * time.Second)
	}
}

func GetCachedAll() string {
	globalPushPoints.Lock()
	str, _ := json.Marshal(globalPushPoints)
	globalPushPoints.Unlock()
	return string(str)
}

func (this *counterCache) GetKeys() []int64 {
	this.RLock()
	retList := make([]int64, 0)
	for k, _ := range this.Points {
		retList = append(retList, k)
	}
	this.RUnlock()
	return retList
}

func (this *counterCache) RemoveTms(tms int64) {
	this.Lock()
	delete(this.Points, tms)
	this.Unlock()
}

func (this *pushPointsCache) AddCounter(counter string) {
	this.Lock()
	tmp := new(counterCache)
	tmp.Points = make(map[int64]float64, 0)
	this.Counters[counter] = tmp
	this.Unlock()
}

func (this *pushPointsCache) GetCounters() []string {
	ret := make([]string, 0)
	this.RLock()
	for k, _ := range this.Counters {
		ret = append(ret, k)
	}
	this.RUnlock()
	return ret
}

func (this *pushPointsCache) RemoveCounter(counter string) {
	this.Lock()
	delete(this.Counters, counter)
	this.Unlock()
}

func (this *pushPointsCache) GetCounterObj(key string) (*counterCache, bool) {
	this.RLock()
	Points, ok := this.Counters[key]
	this.RUnlock()

	return Points, ok
}

func (this *pushPointsCache) AddPoint(point *FalconPoint) {
	counter := calcCounter(point)
	if _, ok := this.GetCounterObj(counter); !ok {
		this.AddCounter(counter)
	}
	counterPoints, _ := this.GetCounterObj(counter)
	counterPoints.AddPoint(point.Timestamp, point.Value)
}

func (this *pushPointsCache) CleanOld() {
	counters := this.GetCounters()
	for _, counter := range counters {
		counterObj, exists := this.GetCounterObj(counter)
		if !exists {
			continue
		}
		tmsList := counterObj.GetKeys()

		//如果列表为空，清理掉这个counter
		if len(tmsList) == 0 {
			this.RemoveCounter(counter)
		} else {
			for _, tms := range tmsList {
				if (time.Now().Unix() - tms) > CACHED_DURATION {
					counterObj.RemoveTms(tms)
				}
			}
		}
	}
}

func calcCounter(point *FalconPoint) string {
	counter := fmt.Sprintf("%s/%s", point.Metric, point.Tags)
	return counter
}
