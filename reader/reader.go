package reader

import (
	"os"
	"time"

	"github.com/didi/falcon-log-agent/common/proc/metric"

	"github.com/hpcloud/tail"
)

// Reader to read file
type Reader struct {
	FilePath    string //配置的路径 正则路径
	t           *tail.Tail
	Stream      chan string
	CurrentPath string //当前的路径
	Close       chan struct{}
}

// NewReader to create a reader
func NewReader(filepath string, stream chan string) (*Reader, error) {
	r := &Reader{
		FilePath: filepath,
		Stream:   stream,
		Close:    make(chan struct{}),
	}
	path := GetCurrentPath(filepath)
	err := r.openFile(os.SEEK_END, path) //默认打开seek_end

	return r, err
}

func (r *Reader) openFile(whence int, filepath string) error {
	seekinfo := &tail.SeekInfo{
		Offset: 0,
		Whence: whence,
	}
	config := tail.Config{
		Location: seekinfo,
		ReOpen:   true,
		Poll:     true,
		Follow:   true,
	}

	t, err := tail.TailFile(filepath, config)
	if err != nil {
		return err
	}
	r.t = t
	r.CurrentPath = filepath
	return nil
}

// StartRead to start to read
func (r *Reader) StartRead() {
	var readCnt, readSwp int64
	var dropCnt, dropSwp int64

	analysClose := make(chan int, 0)
	go func() {
		for {
			// 十秒钟统计一次
			select {
			case <-analysClose:
				return
			case <-time.After(time.Second * 10):
			}
			// 统计时间戳可以不准，但是不能漏
			a := readCnt
			b := dropCnt
			metric.MetricReadLine(r.FilePath, a-readSwp)
			metric.MetricDropLine(r.FilePath, b-dropSwp)
			readSwp = a
			dropSwp = b
		}
	}()

	for line := range r.t.Lines {
		readCnt = readCnt + 1
		select {
		case r.Stream <- line.Text:
		default:
			dropCnt = dropCnt + 1
			//TODO 数据丢失处理，从现时间戳开始截断上报5周期
			// 是否真的要做？
			// 首先，5 周期也是拍脑袋的，只能拍脑袋丢数据，并不能保证准确性
			// 其次，是当前时间推五周期，并不知道日志是什么时候，这个地方有待斟酌
			// 结论，暂且不做，后人注意
		}
	}
	analysClose <- 0
}

// StopRead to stop a read instance
func (r *Reader) StopRead() error {
	return r.t.Stop()
}

// Stop to stop a reader
func (r *Reader) Stop() {
	r.StopRead()
	close(r.Close)

}

// Start a reader
func (r *Reader) Start() {
	go r.StartRead()
	for {
		select {
		case <-time.After(time.Second):
			r.check()
		case <-r.Close:
			close(r.Stream)
			return
		}
	}

}

func (r *Reader) check() {
	nextpath := GetNowPath(r.FilePath)
	if r.CurrentPath != nextpath {
		if _, err := os.Stat(nextpath); err != nil {
			return
		}
		r.t.StopAtEOF()
		if err := r.openFile(os.SEEK_SET, nextpath); err == nil { //从文件开始打开
			go r.StartRead()
		}
	}
}
