package strategy

import (
	"fmt"

	"github.com/didi/falcon-log-agent/common/g"
	"github.com/didi/falcon-log-agent/common/scheme"
	"github.com/didi/falcon-log-agent/common/utils"
)

// 后续开发者切记 : 没有锁，不要修改globalStrategy，更新的时候直接替换，否则会panic
var (
	globalStrategy map[int64]*scheme.Strategy
)

func init() {
	globalStrategy = make(map[int64]*scheme.Strategy, 0)
}

// UpdateGlobalStrategy to update strategy
func UpdateGlobalStrategy(sts []*scheme.Strategy) error {
	tmpStrategyMap := make(map[int64]*scheme.Strategy, 0)
	for _, st := range sts {
		if st.Degree == 0 {
			st.Degree = int64(g.Conf().Strategy.DefaultDegree)
		}
		tmpStrategyMap[st.ID] = st
	}
	globalStrategy = tmpStrategyMap
	return nil
}

// GetListAll to get all strategy
func GetListAll() []*scheme.Strategy {
	stmap := GetDeepCopyAll()
	var ret []*scheme.Strategy
	for _, st := range stmap {
		ret = append(ret, st)
	}
	return ret
}

// GetDeepCopyAll to get all strategy deep copy
func GetDeepCopyAll() map[int64]*scheme.Strategy {
	ret := make(map[int64]*scheme.Strategy, len(globalStrategy))
	for k, v := range globalStrategy {
		ret[k] = utils.DeepCopyStrategy(v)
	}
	return ret
}

// GetAll to get all strategy
func GetAll() map[int64]*scheme.Strategy {
	return globalStrategy
}

// GetByID to get strategy by id
func GetByID(id int64) (*scheme.Strategy, error) {
	st, ok := globalStrategy[id]

	if !ok {
		return nil, fmt.Errorf("ID : %d is not exists in global Cache", id)
	}
	return st, nil

}
