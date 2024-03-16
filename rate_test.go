package goratelimitmanager

import (
	"bufio"
	"context"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

var c int64
var rejcetCnt int64
var acCnt int64

func doA(limit Limiter) {
	c += 1
	ctx := context.Background()
	if !limit.TryAcquire(ctx).Ok {
		// log.Println("reject")
		rejcetCnt += 1
		return
	}
	acCnt += 1
	// log.Println("do")
}

func TestFixedWindow1(t *testing.T) {
	interval := time.Millisecond * 100 // 0.1s
	ticker := time.NewTicker(interval)
	// 1s 5个请求、
	limiter := NewFixedWindowLimiter("f1", time.Second, 5)
	cnt := 0
	for range ticker.C {
		doA(limiter)
		cnt += 1
		if cnt == 1000 {
			ticker.Stop()
			break
		}
	}

}

func TestFixedWindow2(t *testing.T) {
	http.HandleFunc("/h", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, "<h1>hello, world</h1>")
	})

	http.ListenAndServe("0.0.0.0:8080", nil)
	for {
		time.Sleep(time.Second)
	}
}

func TestTime(t *testing.T) {
	now := time.Now()
	time.Sleep(time.Second * 5)
	after := time.Now()

	log.Println(now)
	log.Println(after.Sub(now))
	// sub := after.Sub(now).Seconds()

	duration := time.Duration(after.Sub(now).Seconds() * float64(time.Second))
	log.Println("duration:", duration)
	// 判断两个Duration 是否相等，

}

func TestTimer(t *testing.T) {
	timer := time.NewTimer(time.Second)
	for {
		<-timer.C
		log.Println("timer")
		timer.Reset(time.Second)
	}
}

func TestTicker(t *testing.T) {
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		log.Println("timer")
		ticker.Reset(time.Second)
	}
}

func TestCalulateWindowCnt(t *testing.T) {
	s := calculateWindowCount(time.Hour, time.Millisecond*33)
	log.Println(s)
}

func TestSlidWindowLimiter(t *testing.T) {
	c = 0
	rejcetCnt = 0
	acCnt = 0
	limiter := NewSlideWindowLimiter("1", time.Second, time.Millisecond*100, 100)

	for i := 0; i < 20; i++ {
		go doACircu(limiter)
	}
	time.Sleep(time.Second * 10)
	log.Println("all count:", c)
	log.Println("rejectCount:", rejcetCnt)
	log.Println("accesscCount:", acCnt)
	log.Println("ac+rej:", acCnt+rejcetCnt)

}

func doACircu(limiter Limiter) {
	ticker := time.NewTicker(time.Microsecond * (100 + time.Duration(rand.Int31n(100))))
	for range ticker.C {
		doA(limiter)
	}
}

func TestSllep(t *testing.T) {
	log.Println(time.Now())
	time.Sleep(time.Second * 2)
	log.Println(time.Now())
}

func s(args ...interface{}) {
	log.Println(reflect.TypeOf(args))
	log.Println(args...)
	log.Println(args[0])
}
func TestDot(t *testing.T) {
	s([]string{"a", "b", "c", "d"})
	s(1, 23, 3, 4)
}

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

func Test123(t *testing.T) {
	// 输出当前目录下有多少行代码
	test()
}

func test() {
	totalLines, err := countLinesInDir(".")
	if err != nil {
		log.Printf("Error counting lines: %s\n", err)
		return
	}

	log.Printf("Total lines of code: %d\n", totalLines)
}

// countLinesInDir 返回指定目录及其所有子目录中所有Go文件的代码行数总和
func countLinesInDir(dirPath string) (int, error) {
	totalLines := 0

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".go" {
			lines, err := countLinesInFile(path)
			if err != nil {
				return err
			}
			totalLines += lines
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return totalLines, nil
}

// countLinesInFile 返回文件中的代码行数
func countLinesInFile(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := 0
	for scanner.Scan() {
		lines++
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return lines, nil
}

func TestWatchDog(t *testing.T) {
	svr, _ := NewRateLimitService("", NewRedisClient(), WithWatchDog(time.Second*5))
	log.Println(svr)
	time.Sleep(time.Hour * 5)
}
