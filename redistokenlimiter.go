package ratelimit

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisTokenLimiter struct {
	rdb      *redis.Client
	limiters map[string]LimitProlicy
}
type LimitProlicy struct {
	resetInterval time.Duration
	maxCount      int
	initTokens    int
	rate          time.Duration // 令牌产生速度
	key           string
}

func (r *LimitProlicy) toParams() []string {
	res := make([]string, 0, 5)
	res = append(res, r.resetInterval.String())
	res = append(res, string(time.Now().UnixNano()))
	res = append(res, string(r.maxCount))
	res = append(res, string(r.initTokens))
	return res
}

func (r *RedisTokenLimiter) TryAcquire(key string) {

}

func NewRedisTokenLimiter(polices ...LimitProlicy) {
	limiter := &RedisTokenLimiter{
		rdb:      NewRedisClient(), // todo 替换成集群redis
		limiters: make(map[string]LimitProlicy),
	}
	for _, pol := range polices {
		limiter.limiters[pol.key] = pol
	}

}

func NewRedisClient() *redis.Client {
	// create redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return rdb
}

// todo ， 分布式redis+lua脚本实现分布式限流
type rediser interface {
	Set(ctx context.Context, string, value any, ttl time.Duration) *redis.StatusCmd
}
