package ratelimit

import (
	"fmt"
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
}

// NewFixedWindowLimiter
func NewFixedWindowLimiter(unitTime time.Duration, maxCount int) *FixedWindowLimiter {

	f := &FixedWindowLimiter{
		UnitTime: unitTime,
		Count:    0,
		MaxCount: maxCount,
	}
	go f.resetWindow()
	return f

}

func (f *FixedWindowLimiter) resetWindow() {
	defer func() {
		if x := recover(); x != nil {
			fmt.Printf("Failed to reset window: %v", x)
			go f.resetWindow()
		}
	}()
	ticker := time.NewTicker(f.UnitTime)
	fmt.Println("resetWindow")
	for range ticker.C {
		f.mu.Lock()
		//fmt.Println("reset window")
		f.Count = 0
		// f.LastReqTime = time.Now().Add(-f.UnitTime)
		f.mu.Unlock()

	}
}

func (limiter *FixedWindowLimiter) TryAcquire() (res LimitResult) {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	if limiter.Count < limiter.MaxCount {
		limiter.Count += 1
		res.Ok = true
		return
	}

	// curTime := time.Now()
	res.Ok = false
	return

}
