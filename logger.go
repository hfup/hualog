package hualog

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

type LogLevel uint8

func (l LogLevel) ToString() string {
	switch l {
	case L_DEBUG:
		return "DEBUG"
	case L_INFO:
		return "INFO"
	case L_WARN:
		return "WARN"
	case L_ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

const (
	L_DEBUG LogLevel = 20
	L_INFO  LogLevel = 40
	L_WARN  LogLevel = 60
	L_ERROR LogLevel = 80
)

var (
	logMsgPool = sync.Pool{
		New: func() interface{} {
			return &LogMsg{}
		},
	}

	defaultMaxMsgChanSize = 1000 //默认最大消息队列长度 1000 超过1000条日志将丢弃
)

type LoggerInf interface {
	Debug(msg ...string)
	Info(msg ...string)
	Warn(msg ...string)
	Error(msg ...string)
	DebugN(msg ...string) //消息通知
	InfoN(msg ...string)
	WarnN(msg ...string)
	ErrorN(msg ...string)

	DebugJson(msg LogField) //json格式
	InfoJson(msg LogField)
	WarnJson(msg LogField)
	ErrorJson(msg LogField)

	DebugJsonN(msg LogField) //json格式
	InfoJsonN(msg LogField)
	WarnJsonN(msg LogField)
	ErrorJsonN(msg LogField)
}

type Logger struct {
	Level         LogLevel
	msgChan       chan *LogMsg
	handlers      []LogHandlerInf       //日志处理器
	noticeHandler []LogNoticeHandlerInf //通知处理器

}

func NewLogger(ctx context.Context) *Logger {
	l := &Logger{
		Level:   L_DEBUG,
		msgChan: make(chan *LogMsg, defaultMaxMsgChanSize),
	}
	go l.Run(ctx)
	return l
}

func (l *Logger) Debug(msg ...string) {
	if l.Level <= L_DEBUG {
		l.print(L_DEBUG, false, msg...)
	}
}

func (l *Logger) Info(msg ...string) {
	if l.Level <= L_INFO {
		l.print(L_INFO, false, msg...)
	}
}

func (l *Logger) Warn(msg ...string) {
	if l.Level <= L_WARN {
		l.print(L_WARN, false, msg...)
	}
}

func (l *Logger) Error(msg ...string) {
	if l.Level <= L_ERROR {
		l.print(L_ERROR, false, msg...)
	}
}

func (l *Logger) SetLevel(level LogLevel) {
	l.Level = level
}

func (l *Logger) DebugN(msg ...string) {
	if l.Level <= L_DEBUG {
		l.print(L_DEBUG, true, msg...)
	}
}

func (l *Logger) InfoN(msg ...string) {
	if l.Level <= L_INFO {
		l.print(L_INFO, true, msg...)
	}
}

func (l *Logger) WarnN(msg ...string) {
	if l.Level <= L_WARN {
		l.print(L_WARN, true, msg...)
	}
}

func (l *Logger) ErrorN(msg ...string) {
	if l.Level <= L_ERROR {
		l.print(L_ERROR, true, msg...)
	}
}

func (l *Logger) DebugJson(msg LogField) {
	if l.Level <= L_DEBUG {
		l.printJson(L_DEBUG, false, msg)
	}
}

func (l *Logger) InfoJson(msg LogField) {
	if l.Level <= L_INFO {
		l.printJson(L_INFO, false, msg)
	}
}

func (l *Logger) WarnJson(msg LogField) {
	if l.Level <= L_WARN {
		l.printJson(L_WARN, false, msg)
	}
}

func (l *Logger) ErrorJson(msg LogField) {
	if l.Level <= L_ERROR {
		l.printJson(L_ERROR, false, msg)
	}
}

func (l *Logger) DebugJsonN(msg LogField) {
	if l.Level <= L_DEBUG {
		l.printJson(L_DEBUG, true, msg)
	}
}

func (l *Logger) InfoJsonN(msg LogField) {
	if l.Level <= L_INFO {
		l.printJson(L_INFO, true, msg)
	}
}

func (l *Logger) WarnJsonN(msg LogField) {
	if l.Level <= L_WARN {
		l.printJson(L_WARN, true, msg)
	}
}

func (l *Logger) ErrorJsonN(msg LogField) {
	if l.Level <= L_ERROR {
		l.printJson(L_ERROR, true, msg)
	}
}

func (l *Logger) AddHandler(handler LogHandlerInf) {
	if l.handlers == nil {
		l.handlers = make([]LogHandlerInf, 0, 1)
	}
	l.handlers = append(l.handlers, handler)
}

func (l *Logger) print(level LogLevel, isNotice bool, msg ...string) {
	if len(l.msgChan) >= defaultMaxMsgChanSize {
		return
	}
	msgInfo := logMsgPool.Get().(*LogMsg)
	msgInfo.Reset() //重置
	msgInfo.Level = level.ToString()
	msgInfo.Message = strings.Join(msg, " ")
	msgInfo.Created = time.Now().Unix()
	msgInfo.IsNotice = isNotice
	l.msgChan <- msgInfo
}

func (l *Logger) printJson(level LogLevel, isNotice bool, msg LogField) {
	if len(l.msgChan) >= defaultMaxMsgChanSize {
		return
	}
	if msg == nil {
		return
	}
	msgStr, err := msg.ToJson()
	if err != nil {
		fmt.Println("logger printJson err:", err.Error())
		return
	}
	msgInfo := logMsgPool.Get().(*LogMsg)
	msgInfo.Reset() //重置

	msgInfo.Level = level.ToString()
	msgInfo.Created = time.Now().Unix()
	msgInfo.IsNotice = isNotice
	msgInfo.Message = msgStr

	l.msgChan <- msgInfo
}

func (l *Logger) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			println("logger exit")
			return
		case msg := <-l.msgChan:
			if msg.IsNotice {
				//信息拷贝一份
				msgCopy := logMsgPool.Get().(*LogMsg)
				msgCopy.Reset() //重置
				msgCopy.Level = msg.Level
				msgCopy.Message = msg.Message
				msgCopy.Created = msg.Created
				msgCopy.IsNotice = msg.IsNotice

				go func() {
					defer func() {
						if err := recover(); err != nil {
							fmt.Println("logger dealNoticeHandler err:", err)
						}
					}()
					l.dealNoticeHandler(msgCopy)
				}()
			}
			go func() {
				defer func() {
					if err := recover(); err != nil {
						fmt.Println("logger handlerDeal err:", err)
					}
				}()
				l.handlerDeal(msg)
			}()
		}
	}
}

func (l *Logger) handlerDeal(msg *LogMsg) {
	if l.handlers == nil || len(l.handlers) == 0 {
		printStr := fmt.Sprintf("%s %s %s", msg.Level, time.Unix(msg.Created, 0).Format("2006-01-02 15:04:05"), msg.Message)
		fmt.Println(printStr)
	} else {
		for _, handler := range l.handlers {
			h := handler
			go func() {
				_ = h.Write(msg)
				logMsgPool.Put(msg)
			}()
		}
	}
}

// 通知处理的消息对象 拷贝一份 传递给通知处理器
func (l *Logger) dealNoticeHandler(msg *LogMsg) {
	if l.handlers == nil || len(l.handlers) == 0 {
		return
	}
	for _, handler := range l.noticeHandler {
		h := handler
		go func() {
			_ = h.Notice(msg)
			logMsgPool.Put(msg)
		}()
	}
}
