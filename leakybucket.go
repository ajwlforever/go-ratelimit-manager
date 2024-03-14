package ratelimit

import "sync"

// LeakyBukect
// Cap 桶的容量
// AccessRate
// todo LeakyBucket
type LeakyBukect struct {
	Cap        int64
	AccessRate float64
	InRate     float64

	Mu sync.Mutex
}
