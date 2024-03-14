package ratelimit

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// web 服务器使用ratelimit中间价
// 测试 性能;

type RateLimiterService struct {
	Limiters map[string]Limiter
}

var limiterSvr *RateLimiterService

type MiddleWire func(http.HandlerFunc) http.HandlerFunc

// 使用自定义限流器-slideLimiter
func RateLimiting(key string) MiddleWire {
	return func(f http.HandlerFunc) http.HandlerFunc {

		return func(w http.ResponseWriter, r *http.Request) {
			// 限流
			res := limiterSvr.Limiters[key].TryAcquire()
			if !res.Ok {
				fmt.Println("rejected")
				// 有些限流策略允许请求在 WaitTime后重试
				if res.WaitTime != 0 {
					fmt.Println(time.Now())
					time.Sleep(res.WaitTime)
					fmt.Println(time.Now())
					if res = limiterSvr.Limiters[key].TryAcquire(); !res.Ok {
						fmt.Println("rejected again")
						w.WriteHeader(http.StatusTooManyRequests)
						return
					}
				} else {
					w.WriteHeader(http.StatusTooManyRequests)
					return
				}
			}

			fmt.Println("accepted")
			// 调用下一个HandlerFunc
			f(w, r)

		}
	}
}

func sayHello(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ing")
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
	limiterSvr.Limiters[key1] = NewSlideWindowLimiter(time.Second*10, time.Second*5, 1)
	// 固定窗口算法 5s 只允许通过一个请求
	key2 := "slide2" //利用key值实现 某个接口的 自定义限流器
	limiterSvr.Limiters[key2] = NewFixedWindowLimiter(time.Second*5, 1)
	key3 := "token"
	// 5s 产生一个令牌，最多1个令牌 请求不到令牌阻塞2s
	limiterSvr.Limiters[key3] = NewLimiter(WithTokenBucketLimiter(time.Second*5, 1, 2*time.Second))
	http.HandleFunc("/1", ChaninFunc(sayHello, RateLimiting(key1)))
	http.HandleFunc("/2", ChaninFunc(sayHello, RateLimiting(key2)))
	http.HandleFunc("/token", ChaninFunc(sayHello, RateLimiting(key3)))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(" http.ListenAndServe Error: ")
		panic(err)
	}

}
func init() {
	limiterSvr = &RateLimiterService{
		Limiters: make(map[string]Limiter, 0),
	}
}
