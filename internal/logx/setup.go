package logx

import (
	"io"
	"os"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// levelController 日志输出基本控制器
	levelController zap.AtomicLevel
)

// initDefaultLogger 在没有外部调用Setup进行日志库设置的情况下，进行默认的日志库配置；
// 以便开发单独的小应用的使用时候；
func initDefaultLogger() {
	SetupLogger("")
}

// CloseLogger 系统运行结束时，将日志落盘；
func CloseLogger() {
	_ = rootLogger.Sync()
}

// SetupLogger 配置日志记录器
func SetupLogger(logfile string) {
	levelController = zap.NewAtomicLevelAt(zap.DebugLevel)

	// 将日志输出到屏幕
	config := zap.NewProductionEncoderConfig()
	config.TimeKey = "time"
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoder := zapcore.NewJSONEncoder(config)

	core := zapcore.NewCore(encoder, os.Stdout, levelController)
	// 将日志输出到滚动切割文件中
	if logfile != "" {
		lumberWriterSync := zapcore.AddSync(fileWriter(logfile))

		core = zapcore.NewCore(encoder, lumberWriterSync, levelController)
	}

	// 生产根logger，设置输出调度点(上跳2行），输出Fatal级别的堆栈信息，
	_zLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2), zap.AddStacktrace(zapcore.FatalLevel)) // 选择输出调用点,对于FatalLevel输出调用堆栈；

	rootLogger = newzLogger(_zLogger)
}

func SetLevel(l zapcore.Level) {
	levelController.SetLevel(l)
}

func fileWriter(path string) io.Writer {
	out := &lumberjack.Logger{
		Filename:   path,
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
