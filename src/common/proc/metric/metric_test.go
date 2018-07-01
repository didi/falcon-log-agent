package metric

import (
	"testing"
	"time"
)

func Test(t *testing.T) {
	go MetricLoop(10)

	tests := []struct {
		Name  string
		Tags  string
		Value int64
	}{
		{"mem", "", 100},
		{"read", "a", 1},
		{"read", "b", 2},
		{"read", "c", 3},
		{"drop", "a", 1},
		{"drop", "b", 2},
		{"drop", "c", 3},
		{"analysis", "a", 1},
		{"analysis", "b", 2},
		{"analysis", "c", 3},
		{"analySucc", "a", 1},
		{"analySucc", "b", 2},
		{"analySucc", "c", 3},
		{"pushCnt", "", 10},
		{"pushCnt", "", 20},
		{"pushCnt", "", 30},
		{"pushLatency", "", 10},
		{"pushLatency", "", 20},
		{"pushLatency", "", 30},
	}

	for i := 0; i < 10; i++ {
		for _, test := range tests {
			switch test.Name {
			case "mem":
				MetricMem(test.Value)
			case "read":
				MetricReadLine(test.Tags, test.Value)
			case "drop":
				MetricDropLine(test.Tags, test.Value)
			case "analysis":
				MetricAnalysis(test.Tags, test.Value)
			case "analySucc":
				MetricAnalysisSucc(test.Tags, test.Value)
			case "pushCnt":
				MetricPushCnt(test.Value, false)
				MetricPushCnt(test.Value, true)
			case "pushLatency":
				MetricPushLatency(test.Value)
			}
		}
		time.Sleep(1 * time.Second)
	}
}
