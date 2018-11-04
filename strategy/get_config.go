package strategy

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"fmt"

	"github.com/didi/falcon-log-agent/common/dlog"
	"github.com/didi/falcon-log-agent/common/g"
	"github.com/didi/falcon-log-agent/common/scheme"
)

// 如果有folder，将屏蔽单个配置文件
func GetAllStrategies() ([]*scheme.Strategy, error) {
	if g.StrategyFolder != "" {
		sts, err := getFolderStrategy()
		return sts, err
	}

	if g.StrategyFile != "" {
		sts, err := getFileStrategy()
		return sts, err
	}

	return nil, fmt.Errorf("[-s | -sf] is all empty, please review !")
}

func getFileStrategy() ([]*scheme.Strategy, error) {
	ret, err := _getStrategyFromFile(g.StrategyFile)
	return ret, err
}

func getFolderStrategy() ([]*scheme.Strategy, error) {
	files, err := _getFolderFileList(g.StrategyFolder)
	if err != nil {
		dlog.Errorf("get folder [%s] strategies failed : [%s]", g.StrategyFolder, err.Error())
		return nil, err
	}

	strategyMap := make(map[int64]*scheme.Strategy, 0)
	for _, f := range files {
		sts, err := _getStrategyFromFile(f)
		if err != nil {
			dlog.Errorf("get strategy from file [%s] err: [%s]", f, err.Error())
		}

		for _, st := range sts {
			if _, ok := strategyMap[st.ID]; ok {
				dlog.Error("reduplicated strategy ID : [%d], will drop :[%+v]", st.ID, st)
				continue
			}
			strategyMap[st.ID] = st
		}
	}

	strategyList := make([]*scheme.Strategy, 0)
	for _, st := range strategyMap {
		strategyList = append(strategyList, st)
	}

	return strategyList, nil
}

func _getStrategyFromFile(file string) ([]*scheme.Strategy, error) {
	var config []*scheme.Strategy
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		dlog.Errorf("read config file failed: %s\n", err.Error())
		return nil, err
	}
	if err := json.Unmarshal(bs, &config); err != nil {
		dlog.Errorf("decode config file failed: %s\n", err.Error())
		return nil, err
	}
	dlog.Infof("load config success from %s\n", file)
	return config, nil
}

func _getFolderFileList(path string) ([]string, error) {
	fileList := make([]string, 0)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		dlog.Errorf("cannot get strategy config folder [%s]'s file:%s", path, err.Error())
	}
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	for _, f := range files {
		fileName := fmt.Sprintf("%s%s", path, f.Name())
		fileList = append(fileList, fileName)
	}
	return fileList, nil
}
