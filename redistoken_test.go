package goratelimitmanager

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestRedisToken1(t *testing.T) {
	l := NewRedisTokenLimiter(
		NewRedisClient(),
		"test4",
		time.Second,
		time.Hour,
		1,
		100,
	)
	ctx := context.Background()
	timer := time.NewTicker(time.Second * 2)
	for range timer.C {
		fmt.Println(l.TryAcquire(ctx))
	}
	fmt.Println(l.TryAcquire(ctx))
}

func Test1(t *testing.T) {
	//td := time.Second
	fmt.Println(strconv.Itoa((100)))
}
