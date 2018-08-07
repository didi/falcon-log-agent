package worker

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CachedDuration cached时间周期
const CachedDuration = 60

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

func (c *counterCache) AddPoint(tms int64, value float64) {
	c.Lock()
	c.Points[tms] = value
	c.Unlock()
}

// PostToCache to post points to cache
func PostToCache(paramPoints []*FalconPoint) {
	for _, point := range paramPoints {
		globalPushPoints.AddPoint(point)
	}
}

// CleanLoop to Loop & clean old cache
func CleanLoop() {
	for {
		// 遍历，删掉过期的cache
		globalPushPoints.CleanOld()
		time.Sleep(5 * time.Second)
	}
}

// GetCachedAll to get all cache
func GetCachedAll() string {
	globalPushPoints.Lock()
	str, _ := json.Marshal(globalPushPoints)
	globalPushPoints.Unlock()
	return string(str)
}

// GetKeys
func (c *counterCache) GetKeys() []int64 {
	c.RLock()
	retList := make([]int64, 0)
	for k := range c.Points {
		retList = append(retList, k)
	}
	c.RUnlock()
	return retList
}

// RemoveTms
func (c *counterCache) RemoveTms(tms int64) {
	c.Lock()
	delete(c.Points, tms)
	c.Unlock()
}

// AddCounter
func (pc *pushPointsCache) AddCounter(counter string) {
	pc.Lock()
	tmp := new(counterCache)
	tmp.Points = make(map[int64]float64, 0)
	pc.Counters[counter] = tmp
	pc.Unlock()
}

// GetCounters
func (pc *pushPointsCache) GetCounters() []string {
	ret := make([]string, 0)
	pc.RLock()
	for k := range pc.Counters {
		ret = append(ret, k)
	}
	pc.RUnlock()
	return ret
}

// RemoveCounter
func (pc *pushPointsCache) RemoveCounter(counter string) {
	pc.Lock()
	delete(pc.Counters, counter)
	pc.Unlock()
}

// GetCounterObj
func (pc *pushPointsCache) GetCounterObj(key string) (*counterCache, bool) {
	pc.RLock()
	Points, ok := pc.Counters[key]
	pc.RUnlock()

	return Points, ok
}

// AddPoint
func (pc *pushPointsCache) AddPoint(point *FalconPoint) {
	counter := calcCounter(point)
	if _, ok := pc.GetCounterObj(counter); !ok {
		pc.AddCounter(counter)
	}
	counterPoints, _ := pc.GetCounterObj(counter)
	counterPoints.AddPoint(point.Timestamp, point.Value)
}

// CleanOld
func (pc *pushPointsCache) CleanOld() {
	counters := pc.GetCounters()
	for _, counter := range counters {
		counterObj, exists := pc.GetCounterObj(counter)
		if !exists {
			continue
		}
		tmsList := counterObj.GetKeys()

		//如果列表为空，清理掉这个counter
		if len(tmsList) == 0 {
			pc.RemoveCounter(counter)
		} else {
			for _, tms := range tmsList {
				if (time.Now().Unix() - tms) > CachedDuration {
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
