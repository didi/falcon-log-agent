package g

import (
	"flag"
	"os"

	"github.com/didi/falcon-log-agent/common/dlog"
)

var (
	strategyCfg       = flag.String("s", "", "specify strategy json file")
	strategyFolderCfg = flag.String("sf", "", "specify strategy folder, [-s] will be disables")
	StrategyFile      string
	StrategyFolder    string
)

func InitStrategyFile() {
	flag.Parse()
	cfgFile := *strategyCfg
	cfgFolder := *strategyFolderCfg

	if cfgFile == "" && cfgFolder == "" {
		dlog.Fatal("strategy file/folder not specified: use [-s | -sf] $target")
		os.Exit(1)
	}

	if cfgFile != "" {
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			dlog.Fatalf("strategy file specified not found:%s\n", cfgFile)
			os.Exit(1)
		} else {
			StrategyFile = cfgFile
		}
	}

	if cfgFolder != "" {
		if _, err := os.Stat(cfgFolder); os.IsNotExist(err) {
			dlog.Fatalf("strategy folder specified not found:%s\n", cfgFile)
			os.Exit(1)
		} else {
			StrategyFolder = cfgFolder
		}
	}

	dlog.Infof("use strategy file : %s", StrategyFile)
}
