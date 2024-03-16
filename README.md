# go-ratelimit-manager
使用go实现单机式/分布式限流方案
## Features 

1. 支持多种限流策略，固定窗口/滑动窗口/漏桶/令牌桶/，支持自定义扩展
2. 支持多粒度限流，针对不同的API范围实现多粒度限流。
3. 支持自定义配置，配置文件直接配置。
4. 支持单机式/分布式限流，go原生实现单机式，redis+lua脚本实现分布式限流。
5. 支持WatchDog，监控系统保证限流组件的高可用
6. todo 高可用
7. todo 监控系统
8. todo 平滑限流


## Easy Use 
``` go
go get github.com/ajwlforever/go-ratelimit-manager@latest
```

``` go
import (
	rate "github.com/ajwlforever/go-ratelimit-manager"
)

func main() {
	svr, _ := rate.NewRateLimitService("configs\\ratelimit_config.toml", rate.NewRedisClient())
	// 使用具体的限流器
	res := svr.Limiters["api_ai"].TryAcquire(context.Background())
	if res.Ok {
		 log.Println("allow")
	} else {
		 log.Println("reject")
	}
}

```
### 1. Use
创建xxx限流器的方式:
`NewxxxLimiter` 或者 `NewLimiter(WithxxxLimiter())`
#### (1). FixedWindowLimiter
``` go
// NewFixedWindowLimiter
func NewFixedWindowLimiter(unitTime time.Duration, maxCount int) *FixedWindowLimiter
```
参数：
``` go
UnitTime time.Duration // 窗口时间
MaxCount int           // number 窗口期允许请求的数量
```
新建：
``` go
limiter =  NewFixedWindowLimiter(time.Second*5, 1)

```
或者
``` go
limiter =   NewLimiter(WithFixedWindowLimiter(time.Second*5, 1))
```
#### (2). SlideWindowLimiter
``` go 
type SlideWindowLimiter struct {
	UnitTime      time.Duration // 窗口时间
	SmallUnitTime time.Duration // 小窗口时间
	Cnts          []int         //  每个小窗口的请求数量 - 固定大小- 模拟循环队列
	Index         int           // 目前在循环队列的index
	Count         int           // 实际的请求数量
	MaxCount      int           // number 窗口期允许请求的数量
	Mu            sync.Mutex    //
}
func NewSlideWindowLimiter(unitTime time.Duration, smallUnitTime time.Duration, maxCount int)

slide = NewSlideWindowLimiter(time.Second*10, time.Second*5, 1)
slide = NewLimiter(WithSlideWindowLimiter(time.Second*10, time.Second*5, 1))
```
#### (3). TokenBucketLimiter
``` go 
func WithTokenBucketLimiter(limitRate time.Duration, maxCount int, waitTime time.Duration) 
limiter = NewLimiter(WithTokenBucketLimiter(time.Second*5, 1, 2*time.Second))
```
#### (4). RedisTokenLimiter
``` go
func WithRedisTokenLimiter(rdb *redis.Client, key string, intervalPerPermit time.Duration, resetBucketInterval time.Duration,
	initToken int, MaxCount int)
```
### 2.Configuration
``` toml  
# RateLimiterService配置文件

# TokenBucketLimiter
[[Limiter]]
    Type = "TokenBucketLimiter"
    Key = "api_ai"
    LimitRate = "1s"    # 每秒产生一个令牌
    WaitTime = "500ms" # 最大等待时间500毫秒
    MaxCount = 100 # 令牌桶最大容量

# SlideWindowLimiter
[[Limiter]]
    Type = "SlideWindowLimiter"
    Key = "user_login"
    UnitTime = "60s" # 窗口时间60秒
    SmallUnitTime = "1s" # 小窗口时间1秒
    MaxCount = 5 # 窗口期允许最大请求数量

[[Limiter]]
    Type = "FixedWindowLimiter"
    Key = "filedown"
    UnitTime = "1s" # 窗口时间1秒钟
    MaxCount = 10 # 窗口期允许最大请求数量

[[Limiter]]
    Type = "RedisTokenLimiter"
    Key = "global_rate_limiter"  # redis key
    IntervalPerPermit = "200ms" # 令牌产生速度
    ResetBucketInterval = "1h" # 令牌桶刷新间隔
    MaxCount = 1000 # 令牌桶最大容量
    InitTokens = 500 # 初始化令牌数量
 
```
`func NewRateLimitService(path string, rdb *redis.Client)` 中`path`是配置路径地址，配置文件参考上述，每一个限流器必须要有Type和Key，其余的参数根据不同的种类有不同的参数，参数是必须填写。rdb是redis客户端； 

``` go
func TestConfiguration(t *testing.T) {
	svr, _ := NewRateLimitService("", NewRedisClient())
	// 使用具体的限流器
	res := svr.Limiters["api_ai"].TryAcquire(context.Background())
	if res.Ok {
		 log.Println("allow")
	} else {
		 log.Println("reject")
	}
}
```