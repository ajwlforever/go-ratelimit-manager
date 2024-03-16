package goratelimitmanager

import (
	"context"
	"log"
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
		log.Println(l.TryAcquire(ctx))
	}
	log.Println(l.TryAcquire(ctx))
}

func Test1(t *testing.T) {
	//td := time.Second
	log.Println(strconv.Itoa((100)))
}
