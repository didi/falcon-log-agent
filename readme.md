# falcon-log-agent

## Synopsis

Falcon-Log-Agent is an open-source log collection tool designed to crawl and log feature information from streaming logs.

The acquired feature information is associated with the open source Open-Falcon monitoring system. It can be used for measuring business indicators and for building stability.

## Highlights

- Powerful and flexible
- Calculating with high efficiency
- Friendly strategy configuration supporting secondary development
- Support for a variety of computing methods（cnt、avg、sum、max、min etc.）
- Development language: entire agent written in Golang

## Data Model

The data model is similar to the data model in Open-Falcon: metric and endpoint with a couple of key value tags. Here are two examples:

```
{
    metric: load.1min,
    endpoint: open-falcon-host,
    tags: srv=falcon,idc=aws-sgp,group=az1,
    value: 1.5,
    timestamp: `date +%s`,
    counterType: GAUGE,
    step: 60
}
{
    metric: net.port.listen,
    endpoint: open-falcon-host,
    tags: port=3306,
    value: 1,
    timestamp: `date +%s`,
    counterType: GAUGE,
    step: 60
}
```

## Geting Started

clone & build
```
git clone https://github.com/didichuxing/falcon-log-agent.git && cd falcon-log-agent && sh build.sh
```
change configs
```
# base config
cp cfg/dev.cfg cfg/cfg.json
vim cfg/cfg.json

# strategy config
cp cfg/strategy.dev.json cfg/strategy.json
vim cfg/strategy.json
```

start & stop & status service

```
 # start
./control start

# stop
./control stop

# status
./control status
```

## BasicConfig
```
{
	# configs about log
    "log" : {
        "log_path" : "/var/log/log-agent",
        "log_level" : "INFO",
        "log_rotate_size" : 200,
        "log_rotate_num" : 10
    },

    # configs about service
    "http" : {
        "http_port" : 8003
    },

    # configs about strategy & update
    "strategy" : {
        "update_duration" : 60,
        "default_degree" : 6
    },

    # configs about worker
    "worker" : {
        "worker_num" : 10,
        "queue_size" : 1024000,
        "push_interval" : 1,
        "push_url" : "http://127.0.0.1:1988/v1/push"
    },

    # configs about resource self-limited
    "max_cpu_rate": 0.2,
    "max_mem_rate": 0.05
}
```

## StrategyConfig
```
[
    {
        "id":1,
        "name":"流量500错误数",
        "file_path":"/home/work/myService/log/access.log",
        "time_format":"dd/mmm/yyyy:HH:MM:SS",
        "pattern":"service error 500, num=(\\d+)",
        "exclude":"unimport-request",
        "step":10,
        "tags":{
            "provice" : "province=(\\s+)"
        },
        "func":"cnt",
        "degree":6,
        "comment":"i'm comment"
    }
]
```

## Support TimeFormat List

- dd/mmm/yyyy:HH:MM:SS
- dd/mmm/yyyy HH:MM:SS
- yyyy-mm-ddTHH:MM:SS
- dd-mmm-yyyy HH:MM:SS
- yyyy-mm-dd HH:MM:SS
- yyyy/mm/dd HH:MM:SS
- yyyymmdd HH:MM:SS
- mmm dd HH:MM:SS

And you can also add timeFormat in [src/common/utils/util.go](https://github.com/didichuxing/falcon-log-agent/blob/master/src/common/utils/util.go)
