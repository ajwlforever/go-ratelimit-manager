package goratelimitmanager

import (
	"context"
	"log"
	"sync"
	"time"
)

// TokenBucketLimiter
type TokenBucketLimiter struct {
	LimitRate time.Duration // 一个令牌产生的时间
	TokenChan chan struct{} // 令牌通道，可以理解为桶
	WaitTime  time.Duration // 没有令牌请求等待时间
	MaxCount  int           // 令牌桶的容量
<<<<<<< Updated upstream

	Mu     *sync.Mutex // 令牌桶锁，保证线程安全
	Stop   bool        // 停止标记，结束令牌桶
	Key    string
	Record *LimitRecord
=======
	Mu        *sync.Mutex   // 令牌桶锁，保证线程安全
	Stop      bool          // 停止标记，结束令牌桶
>>>>>>> Stashed changes
}

// NewTokenBucketLimiter
func NewTokenBucketLimiter(key string, limitRate time.Duration, maxCount int, waitTime time.Duration) *TokenBucketLimiter {
	if maxCount < 1 {
		panic("token bucket cap must be large 1")
	}
	l := &TokenBucketLimiter{
		Record:    NewLimitRecord(),
		Key:       key,
		LimitRate: limitRate,
		TokenChan: make(chan struct{}, maxCount),
		WaitTime:  waitTime,
		MaxCount:  maxCount,
		Mu:        &sync.Mutex{},
		Stop:      false,
	}
	go l.Start()
	return l
}

// Start 开启限流器
func (b *TokenBucketLimiter) Start() {
	go b.produceToken()
	// todo rate动态变化
}

// produceToken
func (b *TokenBucketLimiter) produceToken() {
	// 以固定速率生产令牌
	ticker := time.NewTicker(b.LimitRate)
	for range ticker.C {
		b.Mu.Lock()
		if b.Stop {
			b.Mu.Unlock()
			return
		}
		// log.Println(time.Now())
		if cap(b.TokenChan) == len(b.TokenChan) {
			// log.Println("桶满了！")
		} else {
			b.TokenChan <- struct{}{}
		}

		b.Mu.Unlock()
	}
}

func (b *TokenBucketLimiter) TryAcquire(ctx context.Context) (res LimitResult) {
	//  log.Println(time.Now())
	select {
	case <-b.TokenChan:
		res.Ok = true
		b.record(res)
		return
	default:
		// tuichu
		res.Ok = false
		res.WaitTime = b.WaitTime
		b.record(res)
		return
	}
}

func (limiter *TokenBucketLimiter) GetRecord() *LimitRecord {
	return limiter.Record
}

func (s *TokenBucketLimiter) record(res LimitResult) {
	item := &Item{
		Timestamp: time.Now(),
		Key:       s.Key,
		Allowed:   res.Ok,
		Reason:    "TokenBucketLimiter rejected",
	}
	s.Record.Save(item, DETAIL_LEVEL_1)
	log.Println(item.String())
	log.Println("rejectCnt: ", s.Record.rejectCnt)
	log.Println("accessCnt: ", s.Record.allowCnt)
}

// todo StopLimiter
func (b *TokenBucketLimiter) StopLimiter() {
	close(b.TokenChan)
}
