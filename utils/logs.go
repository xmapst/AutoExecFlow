package utils

import (
	"github.com/robfig/cron/v3"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"path/filepath"
)

func LogOutput(dir, name string) io.Writer {
	out := &lumberjack.Logger{
		Filename:   filepath.Join(dir, name+".log"),
		MaxBackups: 7,
		MaxSize:    50,
		MaxAge:     7,
		Compress:   true, // disabled by default
		LocalTime:  true, // use local time zone
	}
	c := cron.New()
	_, _ = c.AddFunc("@daily", func() {
		_ = out.Rotate()
	})
	c.Start()
	return out
}
