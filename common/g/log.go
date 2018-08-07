package g

import (
	"github.com/didi/falcon-log-agent/common/dlog"
)

func InitLog() error {
	backend, err := dlog.NewFileBackend(Conf().Log.LogPath)
	if err != nil {
		return err
	} else {
		dlog.SetLogging(Conf().Log.LogLevel, backend)
		// 日志rotate设置
		backend.Rotate(Conf().Log.LogRotateNum, uint64(1024*1024*Conf().Log.LogRotateSize))
		return nil
	}
}

func CloseLog() {
	dlog.Close()
}
