package ratelimit

import "time"

type LimiterOption func() Limiter

type Limiter interface {
	TryAcquire() LimitResult
	// todo StopLimiter
}

type LimitResult struct {
	Ok       bool
	WaitTime time.Duration
}

// 创建固定窗口限流器的Option
func WithFixedWindowLimiter(unitTime time.Duration, maxCount int) LimiterOption {
	return func() Limiter {
		limiter := NewFixedWindowLimiter(unitTime, maxCount)
		return limiter
	}
}

// 创建滑动窗口限流器的Option
func WithSlideWindowLimiter(unitTime time.Duration, smallUnitTime time.Duration, maxCount int) LimiterOption {
	return func() Limiter {
		limiter := NewSlideWindowLimiter(unitTime, smallUnitTime, maxCount)
		return limiter
	}
}

// 创建令牌桶限流器的Option
func WithTokenBucketLimiter(limitRate time.Duration, maxCount int, waitTime time.Duration) LimiterOption {
	return func() Limiter {
		limiter := NewTokenBucketLimiter(limitRate, maxCount, waitTime)
		return limiter
	}
}

func NewLimiter(option LimiterOption) Limiter {
	return option()
}

// LimiterRecord 限流情况全部记录下来。
// todo LimiterRecord 限流情况全部记录下来。
type LimiterRecord struct {
}
