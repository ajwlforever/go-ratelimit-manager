package goratelimitmanager

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"
)

// web 服务器使用ratelimit中间价
// 测试 性能;

var limiterSvr *RateLimitService

type MiddleWire func(http.HandlerFunc) http.HandlerFunc

// 使用自定义限流器-slideLimiter
func RateLimiting(key string) MiddleWire {
	return func(f http.HandlerFunc) http.HandlerFunc {

		return func(w http.ResponseWriter, r *http.Request) {
			// 限流
			ctx := context.Background()
			res := limiterSvr.Limiters[key].TryAcquire(ctx)
			if !res.Ok {
				log.Println("rejected")
				// 有些限流策略允许请求在 WaitTime后重试
				if res.WaitTime != 0 {
					time.Sleep(res.WaitTime)
					if res = limiterSvr.Limiters[key].TryAcquire(ctx); !res.Ok {
						log.Println("rejected again")
						w.WriteHeader(http.StatusTooManyRequests)
						return
					}
				} else {
					w.WriteHeader(http.StatusTooManyRequests)
					return
				}
			}

			log.Println("accepted")
			// 调用下一个HandlerFunc
			f(w, r)

		}
	}
}

func sayHello(w http.ResponseWriter, r *http.Request) {
	// log.Println("ing")
	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, "<h1>hello, world</h1>")
	return
}

// 一层层的中间件，按顺序包围f
func ChaninFunc(f http.HandlerFunc, middleWires ...MiddleWire) http.HandlerFunc {
	for _, mw := range middleWires {
		f = mw(f)
	}

	return f
}

func StartWeb() {
	// 滑动窗口算法 1s为大窗口 0.1s 为小窗口
	key1 := "slide1"
	limiterSvr.Limiters[key1] = NewSlideWindowLimiter(key1, time.Second*10, time.Second*1, 100)
	// 固定窗口算法 5s 只允许通过一个请求
	key2 := "fixed" //利用key值实现 某个接口的 自定义限流器
	limiterSvr.Limiters[key2] = NewLimiter(WithFixedWindowLimiter(key2, time.Second*5, 1))
	key3 := "token"
	// 5s 产生一个令牌，最多1个令牌 请求不到令牌阻塞2s
	limiterSvr.Limiters[key3] = NewLimiter(WithTokenBucketLimiter(key3, time.Second*5, 1, 2*time.Second))
	//
	key4 := "redis"
	limiterSvr.Limiters[key4] = NewRedisTokenLimiter(
		NewRedisClient(),
		key4,
		time.Second,
		time.Hour,
		1,
		100,
	)
	http.HandleFunc("/slide1", ChaninFunc(sayHello, RateLimiting(key1)))
	http.HandleFunc("/fixed", ChaninFunc(sayHello, RateLimiting(key2)))
	http.HandleFunc("/token", ChaninFunc(sayHello, RateLimiting(key3)))
	http.HandleFunc("/redis", ChaninFunc(sayHello, RateLimiting(key4)))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println(" http.ListenAndServe Error: ")
		panic(err)
	}

}

func StartWebConfiguration() {

}
func init() {
	limiterSvr = &RateLimitService{
		Limiters: make(map[string]Limiter, 0),
	}
}
