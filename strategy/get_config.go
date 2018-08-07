package strategy

import (
	"encoding/json"
	"io/ioutil"

	"github.com/didi/falcon-log-agent/common/dlog"
	"github.com/didi/falcon-log-agent/common/g"
	"github.com/didi/falcon-log-agent/common/scheme"
)

func getFileStrategy() ([]*scheme.Strategy, error) {
	var config []*scheme.Strategy
	bs, err := ioutil.ReadFile(g.StrategyFile)
	if err != nil {
		dlog.Errorf("read config file failed: %s\n", err.Error())
		return nil, err
	}
	if err := json.Unmarshal(bs, &config); err != nil {
		dlog.Errorf("decode config file failed: %s\n", err.Error())
		return nil, err
	}
	dlog.Infof("load config success from %s\n", g.StrategyFile)
	return config, nil

}
