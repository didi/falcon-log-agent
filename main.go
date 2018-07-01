package main

import (
	"common/dlog"
	"common/g"
	"common/proc/metric"
	"common/proc/patrol"
	"common/utils"
	"http"

	"runtime"
	"worker"
)

func main() {
	g.InitAll()
	defer g.CloseLog()

	maxCoreNum := utils.GetCPULimitNum(g.Conf().MaxCPURate)
	dlog.Infof("bind [%d] cpu core", maxCoreNum)
	runtime.GOMAXPROCS(maxCoreNum)

	go metric.MetricLoop(60)
	go worker.UpdateConfigsLoop()
	go patrol.PatrolLoop()
	go worker.PusherStart()

	http.HttpStart()
}
