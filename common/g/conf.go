package g

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"sync"

	"github.com/didi/falcon-log-agent/common/dlog"
	"github.com/didi/falcon-log-agent/common/utils"
)

type logConfig struct {
	LogPath       string `json:"log_path"`
	LogLevel      string `json:"log_level"`
	LogRotateSize int    `json:"log_rotate_size"`
	LogRotateNum  int    `json:"log_rotate_num"`
}

type httpConfig struct {
	HTTPPort int `json:"http_port"`
}

type loadConfig struct {
	UpdateDuration int `json:"update_duration"`
	DefaultDegree  int `json:"default_degree"`
}

type workerConfig struct {
	WorkerNum    int    `json:"worker_num"`
	QueueSize    int    `json:"queue_size"`
	PushInterval int    `json:"push_interval"`
	PushURL      string `json:"push_url"`
}

type Config struct {
	Log        logConfig    `json:"log"`
	Http       httpConfig   `json:"http"`
	Strategy   loadConfig   `json:"strategy"`
	Worker     workerConfig `json:"worker"`
	Endpoint   string       `json:"endpoint"`
	MaxCPURate float64      `json:"max_cpu_rate"`
	MaxCPUNum  int          `json:"max_cpu_num"`
	MaxMemRate float64      `json:"max_mem_rate"`
	MaxMemMB   int          `json:"max_mem_MB"`
}

func Conf() *Config {
	return config
}

var (
	cfg        = flag.String("c", "./cfg/dev.cfg", "specify config file")
	ConfigFile string
	config     *Config
	configLock = new(sync.RWMutex)
)

func InitConfig() {
	flag.Parse()
	cfgFile := *cfg
	if cfgFile == "" {
		dlog.Fatal("config file not specified: use -c $filename")
		os.Exit(1)
	}

	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		dlog.Fatalf("config file specified not found:%s\n", cfgFile)
		os.Exit(1)
	}

	ConfigFile = cfgFile
	dlog.Infof("use config file : %s", ConfigFile)

	if bs, err := ioutil.ReadFile(cfgFile); err != nil {
		dlog.Fatalf("read config file failed: %s\n", err.Error())
		os.Exit(1)
	} else {
		if err := json.Unmarshal(bs, &config); err != nil {
			dlog.Fatalf("decode config file failed: %s\n", err.Error())
			os.Exit(1)
		} else {
			dlog.Infof("load config success from %s\n", cfgFile)
		}
	}
	config.MaxMemMB = utils.CalculateMemLimit(config.MaxMemRate)
	config.MaxCPUNum = utils.GetCPULimitNum(config.MaxCPURate)

	dlog.Infof("config file content : %v", config)
	dlog.Infof("memory limit : %dMB", config.MaxMemMB)
}
