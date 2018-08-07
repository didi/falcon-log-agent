package worker

import (
	"fmt"
	"math"
	"reflect"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/didi/falcon-log-agent/common/dlog"
	"github.com/didi/falcon-log-agent/common/scheme"
	"github.com/didi/falcon-log-agent/common/utils"
	"github.com/didi/falcon-log-agent/strategy"
)

// AnalysPoint to push to Calculate module
// 从worker往计算部分推的Point
type AnalysPoint struct {
	StrategyID int64
	Value      float64
	Tms        int64
	Tags       map[string]string
}

// PointCounter to analysis
//统计的实体
type PointCounter struct {
	sync.RWMutex
	Count int64
	Sum   float64
	Max   float64
	Min   float64
}

// PointsCounter to index the data
// 单策略下，单step的统计对象
// 以Sorted的tagkv的字符串来做索引
type PointsCounter struct {
	sync.RWMutex
	TagstringMap map[string]*PointCounter
}

// StrategyCounter to
// 单策略下的对象, 以step为索引, 索引每一个Step的统计
// 单step统计, 推送完则删
type StrategyCounter struct {
	sync.RWMutex
	Strategy  *scheme.Strategy         //Strategy结构体扔这里，以备不时之需
	TmsPoints map[int64]*PointsCounter //按照时间戳分类的分别的counter
}

// GlobalCounter to be as a global counter store
// 全局counter对象, 以key为索引，索引每个策略的统计
// key : Strategy ID
type GlobalCounter struct {
	sync.RWMutex
	StrategyCounts map[int64]*StrategyCounter
}

// GlobalCount to be as a global counter store
var GlobalCount *GlobalCounter

func init() {
	GlobalCount = new(GlobalCounter)
	GlobalCount.StrategyCounts = make(map[int64]*StrategyCounter)
}

// PushToCount to push to count module
// 提供给Worker用来Push计算后的信息
// 需保证线程安全
func PushToCount(Point *AnalysPoint) error {
	stCount, err := GlobalCount.GetStrategyCountByID(Point.StrategyID)

	// 更新strategyCounts
	if err != nil {
		strategy, err := strategy.GetByID(Point.StrategyID)
		if err != nil {
			dlog.Errorf("GetByID ERROR when count:[%v]", err)
			return err
		}

		GlobalCount.AddStrategyCount(strategy)

		stCount, err = GlobalCount.GetStrategyCountByID(Point.StrategyID)
		// 还拿不到，就出错返回吧
		if err != nil {
			dlog.Errorf("Get strategyCount Failed after addition: %v", err)
			return err
		}
	}

	// 拿到stCount，更新StepCounts
	stepTms := AlignStepTms(stCount.Strategy.Interval, Point.Tms)
	tmsCount, err := stCount.GetByTms(stepTms)
	if err != nil {
		err := stCount.AddTms(stepTms)
		if err != nil {
			dlog.Errorf("Add tms to strategy error: %v", err)
			return err
		}

		tmsCount, err = stCount.GetByTms(stepTms)
		// 还拿不到，就出错返回吧
		if err != nil {
			dlog.Errorf("Get tmsCount Failed By Twice Add: %v", err)
			return err
		}
	}

	//拿到tmsCount, 更新TagstringMap
	tagstring := utils.SortedTags(Point.Tags)
	return tmsCount.Update(tagstring, Point.Value)
}

// AlignStepTms to align the step
// 时间戳向前对齐
func AlignStepTms(step, tms int64) int64 {
	if step <= 0 {
		return tms
	}
	newTms := tms - (tms % step)
	return newTms
}

// GetBytagstring to get Counter structure by tagstring
func (pc *PointsCounter) GetBytagstring(tagstring string) (*PointCounter, error) {
	pc.RLock()
	point, ok := pc.TagstringMap[tagstring]
	pc.RUnlock()

	if !ok {
		return nil, fmt.Errorf("tagstring [%s] not exists", tagstring)
	}
	return point, nil
}

// UpdateCnt to update count
func (pc *PointCounter) UpdateCnt() {
	atomic.AddInt64(&pc.Count, 1)
}

// UpdateSum to update sum
func (pc *PointCounter) UpdateSum(value float64) {
	addFloat64(&pc.Sum, value)
}

// UpdateMaxMin to update max & min
func (pc *PointCounter) UpdateMaxMin(value float64) {
	// 这里要用到结构体的小锁
	// sum和cnt可以不用锁，但是最大最小没办法做到原子操作
	// 只能引入锁
	pc.RLock()
	if math.IsNaN(pc.Max) || value > pc.Max {
		pc.RUnlock()
		pc.Lock()
		if math.IsNaN(pc.Max) || value > pc.Max {
			pc.Max = value
		}
		pc.Unlock()
	} else {
		pc.RUnlock()
	}

	pc.RLock()
	if math.IsNaN(pc.Min) || value < pc.Min {
		pc.RUnlock()
		pc.Lock()
		if math.IsNaN(pc.Min) || value < pc.Min {
			pc.Min = value
		}
		pc.Unlock()
	} else {
		pc.RUnlock()
	}
}

// Update to update value
func (pc *PointsCounter) Update(tagstring string, value float64) error {
	pointCount, err := pc.GetBytagstring(tagstring)
	if err != nil {
		pc.Lock()
		tmp := new(PointCounter)
		tmp.Count = 0
		tmp.Sum = 0
		tmp.Max = math.NaN()
		tmp.Min = math.NaN()
		pc.TagstringMap[tagstring] = tmp
		pc.Unlock()

		pointCount, err = pc.GetBytagstring(tagstring)
		// 如果还是拿不到，就出错返回吧
		if err != nil {
			return fmt.Errorf("when update, cannot get pointCount after add [tagstring:%s]", tagstring)
		}
	}

	pointCount.Lock()
	pointCount.Count = pointCount.Count + 1
	pointCount.Sum = pointCount.Sum + value
	if math.IsNaN(pointCount.Max) || value > pointCount.Max {
		pointCount.Max = value
	}
	if math.IsNaN(pointCount.Min) || value < pointCount.Min {
		pointCount.Min = value
	}
	pointCount.Unlock()

	return nil
}

func addFloat64(val *float64, delta float64) (new float64) {
	for {
		old := *val
		new = old + delta
		if atomic.CompareAndSwapUint64(
			(*uint64)(unsafe.Pointer(val)),
			math.Float64bits(old),
			math.Float64bits(new),
		) {
			break
		}
	}
	return
}

// GetTmsList to get tmslist cached
func (sc *StrategyCounter) GetTmsList() []int64 {
	tmsList := []int64{}
	sc.RLock()
	for tms := range sc.TmsPoints {
		tmsList = append(tmsList, tms)
	}
	sc.RUnlock()
	return tmsList
}

// DeleteTms to delete one tms
func (sc *StrategyCounter) DeleteTms(tms int64) {
	sc.Lock()
	delete(sc.TmsPoints, tms)
	sc.Unlock()
}

// GetByTms get cached counter by tms
func (sc *StrategyCounter) GetByTms(tms int64) (*PointsCounter, error) {
	sc.RLock()
	psCount, ok := sc.TmsPoints[tms]
	if !ok {
		sc.RUnlock()
		return nil, fmt.Errorf("no this tms:%v", tms)
	}
	sc.RUnlock()
	return psCount, nil
}

// AddTms to add Tms to counter
func (sc *StrategyCounter) AddTms(tms int64) error {
	sc.Lock()
	_, ok := sc.TmsPoints[tms]
	if !ok {
		tmp := new(PointsCounter)
		tmp.TagstringMap = make(map[string]*PointCounter, 0)
		sc.TmsPoints[tms] = tmp
	}
	sc.Unlock()
	return nil
}

// UpdateByStrategy to update counter by strategy
// 只做更新和删除，添加 由数据驱动
func (gc *GlobalCounter) UpdateByStrategy(globalStras map[int64]*scheme.Strategy) {
	dlog.Info("Updating global count")
	var delCount, upCount int
	// 先以count的ID为准，更新count
	// 若ID没有了, 那就删掉
	for _, id := range gc.GetIDs() {
		gc.RLock()
		sCount, ok := gc.StrategyCounts[id]
		gc.RUnlock()

		if !ok || sCount.Strategy == nil {
			//证明此策略无效，或已被删除
			//删一下
			delCount = delCount + 1
			gc.deleteByID(id)
			continue
		}

		newStrategy := globalStras[id]

		// 一个是sCount.Strategy, 一个是newStrategy
		// 开始比较
		if !countEqual(newStrategy, sCount.Strategy) {
			//需要清空缓存
			upCount = upCount + 1
			dlog.Infof("strategy [%d] changed, clean data", id)
			gc.cleanStrategyData(id)
			sCount.Strategy = newStrategy
		} else {
			gc.upStrategy(newStrategy)
		}
	}
	dlog.Infof("Update global count done, [del:%d][update:%d]", delCount, upCount)
}

// AddStrategyCount to add strategy to counter
func (gc *GlobalCounter) AddStrategyCount(st *scheme.Strategy) {
	gc.Lock()
	if _, ok := gc.StrategyCounts[st.ID]; !ok {
		tmp := new(StrategyCounter)
		tmp.Strategy = st
		tmp.TmsPoints = make(map[int64]*PointsCounter, 0)
		gc.StrategyCounts[st.ID] = tmp
	}
	gc.Unlock()
}

// GetStrategyCountByID get count by strategy id
func (gc *GlobalCounter) GetStrategyCountByID(id int64) (*StrategyCounter, error) {
	gc.RLock()
	stCount, ok := gc.StrategyCounts[id]
	if !ok {
		gc.RUnlock()
		return nil, fmt.Errorf("No this ID")
	}
	gc.RUnlock()
	return stCount, nil
}

// GetIDs get ids from counter
func (gc *GlobalCounter) GetIDs() []int64 {
	gc.RLock()
	rList := make([]int64, 0)
	for k := range gc.StrategyCounts {
		rList = append(rList, k)
	}
	gc.RUnlock()
	return rList
}

func (gc *GlobalCounter) deleteByID(id int64) {
	gc.Lock()
	delete(gc.StrategyCounts, id)
	gc.Unlock()
}

func (gc *GlobalCounter) cleanStrategyData(id int64) {
	gc.RLock()
	sCount, ok := gc.StrategyCounts[id]
	gc.RUnlock()
	if !ok || sCount == nil {
		return
	}
	sCount.TmsPoints = make(map[int64]*PointsCounter, 0)
	return
}

func (gc *GlobalCounter) upStrategy(st *scheme.Strategy) {
	gc.Lock()
	if _, ok := gc.StrategyCounts[st.ID]; ok {
		gc.StrategyCounts[st.ID].Strategy = st
	}
	gc.Unlock()
}

// countEqual意味着不会对统计的结构产生影响
func countEqual(A *scheme.Strategy, B *scheme.Strategy) bool {
	if A == nil || B == nil {
		return false
	}
	if A.Pattern == B.Pattern && A.Interval == B.Interval && A.Func == B.Func && reflect.DeepEqual(A.Tags, B.Tags) {
		return true
	}
	return false

}
