package goratelimitmanager

import (
	"log"
	"os"
	"testing"
	"time"
)

func TestWeb(t *testing.T) {
	// 打开一个文件用于写入日志，如果文件不存在则创建，如果存在则追加内容
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()

	// 将日志输出设置为之前打开的文件
	log.SetOutput(logFile)
	StartWebConfigurationAndWatchDog()
	for {
		time.Sleep(time.Hour)
	}
}
