package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type zapWriter struct {
	l   *zap.Logger
	lvl zapcore.Level
}

func (w *zapWriter) Write(p []byte) (n int, err error) {
	if w == nil || w.l == nil {
		return len(p), nil
	}
	msg := strings.TrimSpace(string(p))
	if msg == "" {
		return len(p), nil
	}
	if ce := w.l.Check(w.lvl, msg); ce != nil {
		ce.Write()
	}
	return len(p), nil
}

// NewWriter 将 io.Writer 的写入桥接到 zap.Logger。
// 用于把第三方库（例如 gin.DefaultWriter / gin.DefaultErrorWriter）的输出重定向到 zap。
func NewWriter(l *zap.Logger, lvl zapcore.Level) io.Writer {
	return &zapWriter{l: l, lvl: lvl}
}

// NewManagerWriter 是 NewWriter 的便捷封装：从 Manager 中按 bizName 获取 logger。
func NewManagerWriter(m *Manager, bizName string, lvl zapcore.Level) (io.Writer, error) {
	if m == nil {
		return nil, ErrNilManager
	}
	l, err := m.Get(bizName)
	if err != nil {
		return nil, err
	}
	return NewWriter(l, lvl), nil
}

func NewZapLogger(cfg Config, bizName string) (*zap.Logger, zap.AtomicLevel, error) {
	levelText := cfg.Level
	if levelText == "" {
		levelText = "info"
	}

	level, err := zap.ParseAtomicLevel(levelText)
	if err != nil {
		return nil, zap.AtomicLevel{}, fmt.Errorf("failed to parse log level for '%s' (%v): %w", bizName, err, ErrInvalidLogLevel)
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}

	var cores []zapcore.Core
	for _, out := range cfg.Outputs {
		format := out.Format
		if format == "" {
			format = FormatText
		}

		var enc zapcore.Encoder
		switch format {
		case FormatJSON:
			enc = zapcore.NewJSONEncoder(encoderConfig)
		case FormatText:
			textCfg := encoderConfig
			textCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
			textCfg.EncodeLevel = zapcore.CapitalLevelEncoder
			textCfg.EncodeCaller = zapcore.ShortCallerEncoder
			textCfg.ConsoleSeparator = " "
			enc = zapcore.NewConsoleEncoder(textCfg)
		default:
			return nil, zap.AtomicLevel{}, fmt.Errorf("unsupported log format '%s' for '%s': %w (supported formats: %s, %s)", format, bizName, ErrInvalidLogFormat, FormatJSON, FormatText)
		}

		switch out.Type {
		case "file":
			if out.File == nil {
				return nil, zap.AtomicLevel{}, fmt.Errorf("file output config missing for '%s': %w", bizName, ErrInvalidConfigValue)
			}
			fileWriter := zapcore.AddSync(&lumberjack.Logger{
				Filename:   filepath.Join(out.File.Dir, bizName+".log"),
				MaxSize:    out.File.MaxSize,
				MaxBackups: out.File.MaxBackups,
				MaxAge:     out.File.MaxAge,
				Compress:   out.File.Compress,
			})
			cores = append(cores, zapcore.NewCore(enc, fileWriter, level))
		case "console":
			stdoutLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl < zapcore.ErrorLevel && lvl >= level.Level()
			})
			stderrLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl >= zapcore.ErrorLevel && lvl >= level.Level()
			})
			cores = append(cores,
				zapcore.NewCore(enc, zapcore.AddSync(os.Stdout), stdoutLevel),
				zapcore.NewCore(enc, zapcore.AddSync(os.Stderr), stderrLevel),
			)
		}
	}

	core := zapcore.NewTee(cores...)

	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(0),                   // 跳过一层调用栈，显示正确的调用位置
		zap.Fields(zap.String("biz", bizName)), // 添加业务名称字段
	)

	return logger, level, nil
}

// Data 返回一个zap.Field，用于记录任意类型的数据
// 这是一个便捷函数，等价于 zap.Any("data", x)
// x: 要记录的任意类型数据
// 返回: zap.Field，可用于各种日志方法
func Data(x interface{}) zap.Field {
	return zap.Any("data", x)
}
