package utils

import (
	"math"
	"runtime"

	"github.com/didi/falcon-log-agent/common/dlog"

	"github.com/toolkits/nux"
)

func GetCPULimitNum(maxCPURate float64) int {
	return int(math.Ceil(float64(runtime.NumCPU()) * maxCPURate))
}

func CalculateMemLimit(maxMemRate float64) int {
	m, err := nux.MemInfo()
	var memTotal, memLimit int
	if err != nil {
		dlog.Error("failed to get mem.total:", err)
		memLimit = 1024
	} else {
		memTotal = int(m.MemTotal / (1024 * 1024))
		memLimit = int(float64(memTotal) * maxMemRate)
	}

	if memLimit < 1024 {
		memLimit = 1024
	}

	return memLimit
}
