package hualog

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

type LogSplitType uint8

const (
	LS_DEFAULT LogSplitType = 0
	// LS_LEVEL 按照日志级别分开输出
	LS_LEVEL LogSplitType = 1
	// LS_DAY 按照天数分开输出
	LS_DAY LogSplitType = 2
)

type LogHandlerInf interface {
	Write(msg *LogMsg) error //写入日志
}

type LogNoticeHandlerInf interface {
	Notice(msg *LogMsg) error //通知日志
}

type FileHandler struct {
	mu              sync.Mutex          //保证写入文件的原子性
	splitType       LogSplitType        //分割类型
	stdOutMap       map[string]*os.File //文件输出位置
	filePath        string              //文件路径
	currentDayLabel string              //当前日期标签
}

func NewFileHandler(filePath string, splitType LogSplitType) *FileHandler {
	if filePath == "" {
		runPath, _ := os.Getwd()
		filePath = runPath + "/log"
	}
	return &FileHandler{
		splitType: splitType,
		filePath:  filePath,
		stdOutMap: make(map[string]*os.File),
	}
}

func (f *FileHandler) Write(msg *LogMsg) error {
	std, err := f.getFileStream(msg.Level)
	if err != nil {
		return err
	}
	msgStr := fmt.Sprintf("%s %s %s", msg.Level, time.Unix(msg.Created, 0).Format("2006-01-02 15:04:05"), msg.Message)
	_, err = std.WriteString(msgStr + "\n")
	return err
}

func (f *FileHandler) getFileStream(level string) (*os.File, error) {
	f.filePath = GetFullPath(f.filePath)
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.splitType == LS_DEFAULT {
		std, ok := f.stdOutMap["default"]
		if ok {
			return std, nil
		}
		// 直接创建
		std, err := os.OpenFile(f.filePath+"default.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			return nil, err
		}
		f.stdOutMap["default"] = std
		return std, nil
	}
	// 按照日志级别分开输出
	if f.splitType == LS_LEVEL {
		std, ok := f.stdOutMap[level]
		if ok {
			return std, nil
		}
		// 直接创建
		std, err := os.OpenFile(f.filePath+level+".log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			return nil, err
		}
		f.stdOutMap[level] = std
		return std, nil
	}
	// 按照天分开输出
	if f.splitType == LS_DAY {
		skey := time.Now().Format("2006-01-02")
		std, ok := f.stdOutMap["default"]
		if ok && f.currentDayLabel == skey {
			return std, nil
		}
		if ok {
			std.Close() // 关闭之前的文件
		}
		// 直接创建
		std, err := os.OpenFile(f.filePath+skey+".log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			return nil, err
		}
		f.stdOutMap["default"] = std
		f.currentDayLabel = skey
		return std, nil
	}
	return nil, errors.New("not support split type")
}
