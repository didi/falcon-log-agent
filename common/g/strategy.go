package g

import (
	"flag"
	"os"

	"github.com/didi/falcon-log-agent/common/dlog"
)

var (
	strategyCfg  = flag.String("s", "cfg/strategy.dev.json", "specify strategy json file")
	StrategyFile string
)

func InitStrategyFile() {
	flag.Parse()
	cfgFile := *strategyCfg
	if cfgFile == "" {
		dlog.Fatal("strategy file not specified: use -c $filename")
		os.Exit(1)
	}

	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		dlog.Fatalf("strategy file specified not found:%s\n", cfgFile)
		os.Exit(1)
	}

	StrategyFile = cfgFile
	dlog.Infof("use strategy file : %s", StrategyFile)
}
