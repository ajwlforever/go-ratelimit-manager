package goratelimitmanager

import (
	"fmt"
	"time"
)

var DefaultWatchDogTimeout = 30 * time.Second

// 你想有一个后台运行的机制，定期检查或执行一些任务，
// 比如监控限流器的状态、自动调整限流参数、记录日志、清理过期的限流器实例等。
type watchDog struct {
	ticker *time.Ticker
	stopCh chan struct{}
}

func newWatchDog(d time.Duration) *watchDog {
	return &watchDog{
		ticker: time.NewTicker(d),
		stopCh: make(chan struct{}),
	}
}

func (wd *watchDog) Start() {
	go wd.watch()
}

func (wd *watchDog) watch() {
	defer func() {
		fmt.Printf("Failed to WatchDog\n")
		if x := recover(); x != nil {
			fmt.Printf("Restart WatchDog: %v\n", x)
			go wd.Start()
		}
	}()

	for {
		select {
		case <-wd.ticker.C:
			//todo watchDog 在这里执行你的周期性任务
			fmt.Println("watchDog tick")
		case <-wd.stopCh:
			return
		}
	}
}

// Stop stops the watchDog
func (wd *watchDog) Stop() {
	close(wd.stopCh)
	wd.ticker.Stop()
}
