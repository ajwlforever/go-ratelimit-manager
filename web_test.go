package github.com/ajwlforever/go-ratelimit-manager

import (
	"testing"
	"time"
)

func TestWeb(t *testing.T) {
	StartWeb()
	for {
		time.Sleep(time.Hour)
	}
}
