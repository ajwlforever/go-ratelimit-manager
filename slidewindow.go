package goratelimitmanager

import (
	"context"
	"log"
	"sync"
	"time"
)

type SlideWindowLimiter struct {
	UnitTime      time.Duration // 窗口时间
	SmallUnitTime time.Duration // 小窗口时间
	Cnts          []int         //  每个小窗口的请求数量 - 固定大小- 模拟循环队列
	Index         int           // 目前在循环队列的index
	Count         int           // 实际的请求数量
	MaxCount      int           // number 窗口期允许请求的数量
	Mu            sync.Mutex    //
	Record        *LimitRecord
	Key           string //
}

func NewSlideWindowLimiter(key string, unitTime time.Duration, smallUnitTime time.Duration, maxCount int) *SlideWindowLimiter {

	windowCount := calculateWindowCount(unitTime, smallUnitTime)
	s := &SlideWindowLimiter{
		Key:           key,
		UnitTime:      unitTime,
		SmallUnitTime: smallUnitTime,
		MaxCount:      maxCount,
		Cnts:          make([]int, windowCount),
		Index:         0,
		Record:        NewLimitRecord(),
	}

	go s.slideWindow()
	return s
}

func (s *SlideWindowLimiter) slideWindow() {
	defer func() {
		log.Printf("Failed to slide window")
		if x := recover(); x != nil {
			log.Printf("Failed to slide window: %v", x)
			go s.slideWindow()
		}
	}()
	ticker := time.NewTicker(s.SmallUnitTime) // 每个小窗口时间，就滑动！
	for range ticker.C {
		s.Mu.Lock()
		// 滑动
		s.Count -= s.Cnts[s.Index]
		s.Cnts[s.Index] = 0
		s.Index++
		//  log.Println(s.Count)
		if s.Index >= len(s.Cnts) {
			s.Index = 0
		}
		s.Mu.Unlock()

	}
}
func (s *SlideWindowLimiter) TryAcquire(ctx context.Context) (res LimitResult) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	if s.Count < s.MaxCount {
		s.Count += 1
		s.Cnts[s.Index] += 1
		res.Ok = true
		return
	}
	res.Ok = false
	s.record(res)
	return
}

// record 记录尝试请求的最终结果
func (s *SlideWindowLimiter) record(res LimitResult) {
	item := &Item{
		Timestamp: time.Now(),
		Key:       s.Key,
		Allowed:   res.Ok,
		Reason:    "SlideWindowLimiter rejected",
	}
	s.Record.Save(item, DETAIL_LEVEL_1)
}

// calculateWindowCount 计算 unitTime 被 smallUnitTime划分为几份
func calculateWindowCount(unitTime time.Duration, smallUnitTime time.Duration) int {
	return int(unitTime / smallUnitTime)
}
