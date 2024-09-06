package utils

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"time"
)

var Log = NewLogger()

type CustomTextFormatter struct {
	logrus.TextFormatter
}

func (f *CustomTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timestamp := entry.Time.Format(f.TimestampFormat)

	//HasCaller()为true才会有调用信息
	if entry.HasCaller() {
		fName := filepath.Base(entry.Caller.File)
		fmt.Fprintf(b, "time=[%s]\tlevel=[%s]\tfile=[%s:%d %s]\tmsg=[%s]\t",
			timestamp, entry.Level, fName, entry.Caller.Line, entry.Caller.Function, entry.Message)
	} else {
		fmt.Fprintf(b, "time=[%s]\tlevel=[%s]\tmsg=[%s]\t", timestamp, entry.Level, entry.Message)
	}

	// 添加其他字段
	for key, value := range entry.Data {
		fmt.Fprintf(b, "%s=[%v]\t", key, value)
	}

	// 添加换行符
	b.WriteByte('\n')
	return b.Bytes(), nil
}

func NewLogger() *logrus.Logger {
	log := logrus.New()
	log.SetFormatter( // 自定义格式化器
		&CustomTextFormatter{
			logrus.TextFormatter{
				TimestampFormat: time.RFC3339, // 使用 RFC3339 时间格式
				FullTimestamp:   true,
			},
		},
	)
	return log
}

func SetLogLevel(level string) {
	switch level {
	case "debug":
		Log.SetLevel(logrus.DebugLevel)
	case "info":
		Log.SetLevel(logrus.InfoLevel)
	case "error":
		Log.SetLevel(logrus.ErrorLevel)
	default:
		Log.SetLevel(logrus.InfoLevel)
	}
}
