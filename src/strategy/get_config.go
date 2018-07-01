package strategy

import (
	"encoding/json"
	"io/ioutil"

	"common/dlog"
	"common/g"
	"common/scheme"
)

func getFileStrategy() ([]*scheme.Strategy, error) {
	var config []*scheme.Strategy
	if bs, err := ioutil.ReadFile(g.StrategyFile); err != nil {
		dlog.Errorf("read config file failed: %s\n", err.Error())
		return nil, err
	} else {
		if err := json.Unmarshal(bs, &config); err != nil {
			dlog.Errorf("decode config file failed: %s\n", err.Error())
			return nil, err
		} else {
			dlog.Infof("load config success from %s\n", g.StrategyFile)
		}
	}
	return config, nil

}
