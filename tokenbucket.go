package ratelimit

import (
	"sync"
	"time"
)

// TokenBucketLimiter
type TokenBucketLimiter struct {
	LimitRate time.Duration // 一个令牌产生的时间
	TokenChan chan struct{} // 令牌通道，可以理解为桶
	WaitTime  time.Duration // 没有令牌请求等待时间
	MaxCount  int           // 令牌桶的容量

	Mu   *sync.Mutex // 令牌桶锁，保证线程安全
	Stop bool        // 停止标记，结束令牌桶
}

// NewTokenBucketLimiter
func NewTokenBucketLimiter(limitRate time.Duration, maxCount int, waitTime time.Duration) *TokenBucketLimiter {
	if maxCount < 1 {
		panic("token bucket cap must be large 1")
	}
	l := &TokenBucketLimiter{
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
		//fmt.Println(time.Now())
		if cap(b.TokenChan) == len(b.TokenChan) {
			//fmt.Println("桶满了！")
		} else {
			b.TokenChan <- struct{}{}
		}

		b.Mu.Unlock()
	}
}

func (b *TokenBucketLimiter) TryAcquire() (res LimitResult) {
	// fmt.Println(time.Now())
	select {
	case <-b.TokenChan:
		res.Ok = true
		return
	default:
		// tuichu
		res.Ok = false
		res.WaitTime = b.WaitTime
		return
	}
}

// todo StopLimiter
func (b *TokenBucketLimiter) StopLimiter() {
	close(b.TokenChan)
}
