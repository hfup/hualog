package hualog

import (
	"context"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	logger := NewLogger(context.TODO())
	//fileHandler := NewFileHandler("", LS_DAY)
	//logger.AddHandler(fileHandler)
	logger.SetLevel(L_DEBUG)

	logger.Debug("hello world")
	logger.InfoJson(LogField{"name": "hualong", "age": 18})
	logger.DebugN("hello world notice")

	time.Sleep(10 * time.Second)
}
