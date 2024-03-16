package goratelimitmanager

import (
	"sync"
	"time"
)

var (
	DETAIL_LEVEL_1 = 0 // 简易记录
	DETAIL_LEVEL_2 = 1 // 复杂记录
)

// 设计 简易记录/复杂记录
// LimitRecord 限流情况全部记录下来。
// todo LimitrRecord 限流情况全部记录下来。
// LimitRecord 嵌入到每个限流器结构体内
type LimitRecord struct {
	allows    []Item
	rejects   []Item
	allowCnt  int
	rejectCnt int
	mu        sync.Mutex // 锁住
}

func NewLimitRecord() *LimitRecord {
	return &LimitRecord{
		allows:  make([]Item, 0, 1000),
		rejects: make([]Item, 0, 1000),
	}
}

type Item struct {
	Timestamp time.Time `json:"timestamp"`
	Key       string    `json:"Key"`
	Allowed   bool      `json:"allowed"`
	Reason    string    `json:"reason,omitempty"` // 如果Reason为空，则在JSON中省略这个字段
}

type Record interface {
	Save(item *Item, detailLevel int) // item是一条请求，detailLevel是这个条请求记录的级别
}

// Save
func (record *LimitRecord) Save(item *Item, detailLevel int) {
	switch {
	case detailLevel == DETAIL_LEVEL_1:
		record.easySave(item)
	case detailLevel == DETAIL_LEVEL_2:
		record.allSave(item)
	}
}

// easySave 只统计请求数量变化
func (record *LimitRecord) easySave(item *Item) {
	record.mu.Lock()
	defer record.mu.Unlock()

	if !item.Allowed {
		record.rejectCnt += 1
	} else {
		record.allowCnt += 1
	}
}

// allSave item整个存入record
func (record *LimitRecord) allSave(item *Item) {
	record.mu.Lock()
	defer record.mu.Unlock()

	if !item.Allowed {
		record.rejectCnt += 1
		record.rejects = append(record.rejects, *item)
	} else {
		record.allowCnt += 1
		record.rejects = append(record.rejects, *item)
	}
}
