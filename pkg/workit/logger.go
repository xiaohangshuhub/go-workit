package workit

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config 日志配置
type Config struct {
	Level      string // 日志级别 debug, info, warn, error
	Filename   string // 日志文件路径
	MaxSize    int    // 每个日志文件最大尺寸,单位MB
	MaxBackups int    // 保留的旧日志文件最大数量
	MaxAge     int    // 保留的旧日志文件最大天数
	Compress   bool   // 是否压缩旧日志文件
	Console    bool   // 是否输出到控制台
}

// newLogger 初始化日志
func newLogger(conf *Config) *zap.Logger {
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

	// 2. 配置日志格式
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if conf.Console {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // 控制台彩色
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder // 文件 JSON
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 3. 日志级别
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(conf.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(writers...),
		level,
	)

	return zap.New(core, zap.AddCaller())
}

// customTimeEncoder 自定义时间格式
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}
