package strategy

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/didi/falcon-log-agent/common/dlog"
	"github.com/didi/falcon-log-agent/common/scheme"
	"github.com/didi/falcon-log-agent/common/utils"

	"github.com/parnurzeal/gorequest"
)

func getHTTPStrategy(addr, uri string, timeout int) ([]*scheme.Strategy, error) {
	hostname, err := utils.LocalHostname()
	if err != nil {
		return nil, err
	}
	urlTemp := fmt.Sprintf("%s%s", addr, uri)

	url := fmt.Sprintf(urlTemp, hostname)
	dlog.Infof("URL in get strategy : [%s]", url)

	body, err := getRequest(url, timeout)

	if err != nil {
		return nil, err
	}

	var strategyResp []*scheme.Strategy
	err = json.Unmarshal([]byte(body), &strategyResp)
	if err != nil {
		dlog.Errorf("json decode error when update strategy : [%v]", err)
		return nil, err
	}
	return strategyResp, nil
}

func getRequest(url string, timeout int) (string, error) {
	request := gorequest.New().Timeout(time.Duration(timeout) * time.Second)
	resp, body, errs := request.Get(url).End()

	if errs == nil {
		if resp.StatusCode != 200 {
			dlog.Errorf("get HTTP Request Response: [code:%d][body:%s][errs:%v]", resp.StatusCode, body, errs)
			return body, fmt.Errorf("Code is not 200")
		} else {
			//err == nil  && code == 200
			dlog.Infof("get HTTP Request Response : [code:%d][body:%s]", 200, body)
			return body, nil
		}
	} else {
		dlog.Errorf("get HTTP Request Response: [body:%s][errs:%v]", body, errs)
		return body, fmt.Errorf("%v", errs)
	}
}
