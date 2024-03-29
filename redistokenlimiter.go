package goratelimitmanager

import (
	"context"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisTokenLimiter struct {
	rdb                 *redis.Client
	intervalPerPermit   time.Duration // 令牌产生速度
	resetBucketInterval time.Duration // 令牌桶刷新间隔
	MaxCount            int
	initTokens          int
	key                 string
	Record              *LimitRecord
}

func (r *RedisTokenLimiter) toParams() []any {
	res := make([]any, 0, 5)
	res = append(res, int64(r.intervalPerPermit/time.Millisecond))   // 转换成以ms为单位  生成令牌的间隔(ms)
	res = append(res, time.Now().UnixMilli())                        //当前时间
	res = append(res, string(strconv.Itoa(r.initTokens)))            // 令牌桶初始化的令牌数
	res = append(res, string(strconv.Itoa(r.MaxCount)))              // 令牌桶的上限
	res = append(res, int64(r.resetBucketInterval/time.Millisecond)) // 重置桶内令牌的时间间隔

	return res
}

func (r *RedisTokenLimiter) TryAcquire(ctx context.Context) (res LimitResult) {
	params := r.toParams()

	luaPath := "tokenbucket.lua"
	file, _ := os.Open(luaPath)
	luas, _ := io.ReadAll(file)
	log.Println(params...)
	tokenScript := redis.NewScript(string(luas))
	n, err := tokenScript.Eval(ctx, *r.rdb, []string{r.key}, params...).Result()
	if err != nil {
		panic("failed to exec lua script: " + err.Error())
	}
	log.Println("remaining tokens: ", n)
	if n.(int64) <= 0 {
		res.Ok = false
	} else {
		res.Ok = true
	}
	r.record(res)
	return
}

func (limiter *RedisTokenLimiter) GetRecord() *LimitRecord {
	return limiter.Record
}

// record 记录尝试请求的最终结果
func (s *RedisTokenLimiter) record(res LimitResult) {
	item := &Item{
		Timestamp: time.Now(),
		Key:       s.key,
		Allowed:   res.Ok,
		Reason:    "RedisTokenLimiter rejected",
	}
	s.Record.Save(item, DETAIL_LEVEL_1)
	log.Println(item.String())
	log.Println("rejectCnt: ", s.Record.rejectCnt)
	log.Println("accessCnt: ", s.Record.allowCnt)
}

func NewRedisTokenLimiter(rdb *redis.Client, key string, intervalPerPermit time.Duration, resetBucketInterval time.Duration,
	initToken int, MaxCount int) *RedisTokenLimiter {

	limiter := &RedisTokenLimiter{
		rdb:                 rdb, // todo 替换成可自定义配置redis
		key:                 key,
		resetBucketInterval: resetBucketInterval,
		initTokens:          initToken,
		MaxCount:            MaxCount,
		intervalPerPermit:   intervalPerPermit,
		Record:              NewLimitRecord(),
	}
	return limiter

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
