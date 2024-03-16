package goratelimitmanager

import (
	"context"
	"log"
	"sync"
	"time"
)

// fixedWindow 限流算法

// FixedWindowLimiter
type FixedWindowLimiter struct {
	UnitTime time.Duration // 窗口时间
	Count    int           // 实际的请求数量
	MaxCount int           // number 窗口期允许请求的数量
	mu       sync.Mutex
	Key      string //
	Record   *LimitRecord
}

// NewFixedWindowLimiter
func NewFixedWindowLimiter(key string, unitTime time.Duration, maxCount int) *FixedWindowLimiter {

	f := &FixedWindowLimiter{
		Record:   NewLimitRecord(),
		UnitTime: unitTime,
		Count:    0,
		MaxCount: maxCount,
		Key:      key,
	}
	go f.resetWindow()
	return f

}

func (f *FixedWindowLimiter) resetWindow() {
	defer func() {
		if x := recover(); x != nil {
			log.Printf("Failed to reset window: %v", x)
			go f.resetWindow()
		}
	}()
	ticker := time.NewTicker(f.UnitTime)
	// log.Println("resetWindow")
	for range ticker.C {
		f.mu.Lock()
		// log.Println("reset window")
		f.Count = 0
		// f.LastReqTime = time.Now().Add(-f.UnitTime)
		f.mu.Unlock()

	}
}

func (limiter *FixedWindowLimiter) TryAcquire(ctx context.Context) (res LimitResult) {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	if limiter.Count < limiter.MaxCount {
		limiter.Count += 1
		res.Ok = true
		limiter.record(res)
		return
	}

	// curTime := time.Now()
	res.Ok = false
	limiter.record(res)
	return

}
func (limiter *FixedWindowLimiter) GetRecord() *LimitRecord {
	return limiter.Record
}

func (l *FixedWindowLimiter) record(res LimitResult) {
	item := &Item{
		Timestamp: time.Now(),
		Key:       l.Key,
		Allowed:   res.Ok,
		Reason:    "FixedWindowLimiter rejected",
	}
	l.Record.Save(item, DETAIL_LEVEL_1)
	log.Println(item.String())
	log.Println("rejectCnt: ", l.Record.rejectCnt)
	log.Println("accessCnt: ", l.Record.allowCnt)
}
