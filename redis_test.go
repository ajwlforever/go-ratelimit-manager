package goratelimitmanager

import (
	"testing"
	"time"
)

func TestRedis1(t *testing.T) {

	limiter := NewRedisTokenLimiter("test", time.Hour, 100, 10)

	_ = limiter
}
