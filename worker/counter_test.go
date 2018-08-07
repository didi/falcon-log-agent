package worker

import (
	"fmt"
	"testing"
	"time"

	"github.com/didi/falcon-log-agent/common/scheme"
	"github.com/didi/falcon-log-agent/strategy"
)

func TestCounterStart(a *testing.T) {
	go strategy.Updater()
	for {
		globalCount.UpdateByStrategy(strategy.GetAllDeepCopy())
		for id, st := range globalCount.StrategyCounts {
			fmt.Printf("%d  ----\n", id)
			fmt.Printf("     %v\n", *st.Strategy)
		}
		time.Sleep(1 * time.Second)
	}
	time.Sleep(20 * time.Second)
}

func TestPushToCount(a *testing.T) {
	go strategy.Updater()
	time.Sleep(2 * time.Second)

	go func() {
		for {
			time.Sleep(time.Second)
			a, err := globalCount.GetStrategyCountByID(369944)
			if err != nil {
				fmt.Printf("error : %v\n", err)
				continue
			}
			a.RLock()
			for k, v := range a.TmsPoints {
				fmt.Println(k, *v, v.TagstringMap["host=hahaha"])
			}
			fmt.Println("-----")
			a.RUnlock()
		}
	}()

	for i := 0; i < 100; i++ {
		tmp := scheme.AnalysPoint{
			StrategyID: 369944,
			Value:      1.11111,
			Tms:        time.Now().Unix(),
			Tags:       map[string]string{"host": "hahaha"},
		}
		PushToCount(tmp)
		time.Sleep(1 * time.Second)
	}

}

func mockupsStrategy() *scheme.StrategyCache {
	return &scheme.StrategyCache{
		Updated: time.Now().Unix(),
		Strategys: map[int64]*scheme.Strategy{
			1: &scheme.Strategy{
				ID:              1,
				Name:            "alarmer.not.found.conf",
				FilePath:        "/tmp/memeda.log",
				TimeFormat:      "yyyy-mm-dd HH:MM:SS",
				Pattern:         "ERROR alarmer/alarmer.go:115 not found conf",
				MeasurementType: "LOG",
				Interval:        10,
				Tags:            map[string]string{},
				Func:            "cnt",
				Degree:          0,
				Unit:            "s",
				Comment:         "hahahaha",
				Creator:         "gaojiasheng",
				NS:              []string{"collect.gz01.alarm.old.monitor.odin.op.didi.com"},
				SrvUpdated:      "--",
				LocalUpdated:    time.Now().Unix(),
			},
		},
	}
}
