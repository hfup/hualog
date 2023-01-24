package hualog

import (
	"encoding/json"
	"runtime"
)

type LogField map[string]any

func (l LogField) ToJson() (string, error) {
	data, err := json.Marshal(l)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

type LogMsg struct {
	Id       int64  `gorm:"primary_key;auto_increment"`
	Level    string `gorm:"type:varchar(10);comment:'日志级别';" json:"level"`
	Created  int64  `gorm:"type:bigint;comment:'日志时间';" json:"created"`
	Message  string `gorm:"type:text;comment:'日志内容';" json:"message"`
	IsNotice bool   //是否通知
}

func (*LogMsg) TableName() string {
	return "log_msg"
}

func (l *LogMsg) Reset() {
	l.IsNotice = false
	l.Id = 0
	l.Level = ""
	l.Created = 0
	l.Message = ""
}

func GetFullPath(path string) string {
	if path == "" {
		return ""
	}
	if runtime.GOOS == "windows" {
		if path[len(path)-1:] == "\\" {
			return path
		}
		return path + "\\"
	}
	if path[len(path)-1:] == "/" {
		return path
	}
	return path + "/"
}
