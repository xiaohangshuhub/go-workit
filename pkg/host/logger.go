package host

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Level      string // 日志级别 debug, info, warn, error
	Filename   string // 日志文件路径
	MaxSize    int    // 每个日志文件最大尺寸,单位MB
	MaxBackups int    // 保留的旧日志文件最大数量
	MaxAge     int    // 保留的旧日志文件最大天数
	Compress   bool   // 是否压缩旧日志文件
	Console    bool   // 是否输出到控制台
}

// 初始化日志
func newLogger(conf *Config) *zap.Logger {
	// 设置默认值
	if conf.MaxSize == 0 {
		conf.MaxSize = 100
	}
	if conf.MaxBackups == 0 {
		conf.MaxBackups = 3
	}
	if conf.MaxAge == 0 {
		conf.MaxAge = 7
	}

	// 确保日志目录存在
	if conf.Filename != "" {
		err := os.MkdirAll(filepath.Dir(conf.Filename), 0744)
		if err != nil {
			panic(fmt.Sprintf("create log directory failed: %v", err))
		}
	}

	// 设置日志编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 配置日志输出
	var writers []zapcore.WriteSyncer
	if conf.Filename != "" {
		writers = append(writers, zapcore.AddSync(&lumberjack.Logger{
			Filename:   conf.Filename,
			MaxSize:    conf.MaxSize,
			MaxBackups: conf.MaxBackups,
			MaxAge:     conf.MaxAge,
			Compress:   conf.Compress,
		}))
	}
	if conf.Console {
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}

	var encoder zapcore.Encoder

	if conf.Console {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 创建核心
	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(writers...),
		getLogLevel(conf.Level),
	)

	// 创建日志记录器
	return zap.New(core, zap.AddCaller())
}

// 自定义时间格式
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// 获取日志级别
func getLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
