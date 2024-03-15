package github.com/ajwlforever/go-ratelimit-manager

import (
	"context"
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/go-redis/redis/v8"
)

var (
	ConfigPath = "ratelimit_config.toml"
)

type RateLimitConfig struct {
	Limiters []struct {
		Type                string `toml:"Type"`
		Key                 string `toml:"Key"`
		LimitRate           string `toml:"LimitRate,omitempty"`
		WaitTime            string `toml:"WaitTime,omitempty"`
		MaxCount            int    `toml:"MaxCount"`
		UnitTime            string `toml:"UnitTime,omitempty"`
		SmallUnitTime       string `toml:"SmallUnitTime,omitempty"`
		IntervalPerPermit   string `toml:"IntervalPerPermit,omitempty"`
		ResetBucketInterval string `toml:"ResetBucketInterval,omitempty"`
		InitTokens          int    `toml:"InitTokens,omitempty"`
	} `toml:"Limiter"`
}
type LimiterOption func() Limiter

type RateLimitService struct {
	Limiters map[string]Limiter
}
type Limiter interface {
	TryAcquire(ctx context.Context) LimitResult
	// 有需要用key值来获取分布式令牌的
	// todo StopLimiter
}

type LimitResult struct {
	Ok       bool
	WaitTime time.Duration
}

func NewRateLimitService(path string, rdb *redis.Client) (svr *RateLimitService, err error) {
	if path != "" {
		ConfigPath = path
	}
	var config RateLimitConfig
	// 读取配置文件
	if _, err = toml.DecodeFile(ConfigPath, &config); err != nil {
		fmt.Println(err)
		return
	}
	svr = &RateLimitService{
		Limiters: make(map[string]Limiter),
	}

	// todo fmt change to log
	// Limiters 注入 svr
	for idx, c := range config.Limiters {
		switch c.Type {
		case "TokenBucketLimiter":
			if paramsCheck(c.LimitRate, c.WaitTime, c.Key) && c.MaxCount > 0 {
				// 根据配置初始化TokenBucketLimiter
				fmt.Println("Initializing TokenBucketLimiter:", c.Key)
				var lr, wt time.Duration
				lr, err = time.ParseDuration(c.LimitRate)
				wt, err = time.ParseDuration(c.WaitTime)
				if err != nil {
					panicInitRLConfig(idx)
				}
				svr.Limiters[c.Key] = NewLimiter(WithTokenBucketLimiter(
					lr,
					c.MaxCount,
					wt,
				))
			} else {
				panicInitRLConfig(idx)
			}
		case "SlideWindowLimiter":
			if paramsCheck(c.Key, c.UnitTime, c.SmallUnitTime) && c.MaxCount > 0 {
				// 根据配置初始化SlideWindowLimiter
				fmt.Println("Initializing SlideWindowLimiter:", c.Key)
				var ut, st time.Duration
				ut, err = time.ParseDuration(c.UnitTime)
				st, err = time.ParseDuration(c.SmallUnitTime)
				if err != nil {
					panicInitRLConfig(idx)
				}
				svr.Limiters[c.Key] = NewLimiter(WithSlideWindowLimiter(
					ut, st, c.MaxCount,
				))
			} else {
				panicInitRLConfig(idx)
			}
		case "FixedWindowLimiter":
			if paramsCheck(c.Key, c.UnitTime) && c.MaxCount > 0 {
				fmt.Println("Initializing FixedWindowLimiter:", c.Key)
				var ut time.Duration
				ut, err = time.ParseDuration(c.UnitTime)
				if err != nil {
					panicInitRLConfig(idx)
				}
				svr.Limiters[c.Key] = NewLimiter(WithFixedWindowLimiter(
					ut, c.MaxCount,
				))
			} else {
				panicInitRLConfig(idx)
			}
		case "RedisTokenLimiter":
			if paramsCheck(c.Key, c.IntervalPerPermit, c.ResetBucketInterval) && c.MaxCount > 0 && c.InitTokens <= c.MaxCount {
				// 根据配置初始化RedisTokenLimiter
				fmt.Println("Initializing RedisTokenLimiter:", c.Key)
				// 确保传递rdb到RedisTokenLimiter的构造函数中
				var interval, reset time.Duration
				interval, err = time.ParseDuration(c.IntervalPerPermit)
				reset, err = time.ParseDuration(c.ResetBucketInterval)
				if err != nil {
					panicInitRLConfig(idx)
				}
				svr.Limiters[c.Key] = NewLimiter(WithRedisTokenLimiter(
					rdb, c.Key, interval, reset, c.InitTokens, c.MaxCount,
				))
			} else {
				panicInitRLConfig(idx)
			}
		default:
			fmt.Println("Unknown Limiter Type:", c.Type)
		}
	}

	return

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

// 创建Redis分布式限流器的option
func WithRedisTokenLimiter(rdb *redis.Client, key string, intervalPerPermit time.Duration, resetBucketInterval time.Duration,
	initToken int, MaxCount int) LimiterOption {
	return func() Limiter {
		limiter := NewRedisTokenLimiter(rdb, key, intervalPerPermit, resetBucketInterval,
			initToken, MaxCount)
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

func paramsCheck(ss ...string) bool {
	for _, s := range ss {
		if s == "" {
			return false
		}
	}
	return true
}

func panicInitRLConfig(idx int) {
	fmt.Printf("init RateLimitConfiguration error, look at the %v limiter", idx+1)
	panic("init RateLimitConfiguration error")
}
