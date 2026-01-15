package log

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewZapLogger(cfg Config, bizName string) (*zap.Logger, zap.AtomicLevel, error) {
	level, err := zap.ParseAtomicLevel(cfg.Level)
	if err != nil {
		return nil, zap.AtomicLevel{}, fmt.Errorf("failed to parse log level for '%s' (%v): %w", bizName, err, ErrInvalidLogLevel)
	}

	// 文件输出 writer
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(cfg.Dir, bizName+".log"),
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	})

	// 根据配置选择编码器
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

	format := cfg.Format
	if format == "" {
		format = "json" // 默认使用 json 格式
	}

	// 创建文件编码器
	var fileEncoder zapcore.Encoder
	switch format {
	case "json":
		fileEncoder = zapcore.NewJSONEncoder(encoderConfig)
	case "console":
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // console 模式使用彩色输出
		fileEncoder = zapcore.NewConsoleEncoder(encoderConfig)
	case "text", "standard":
		// 标准应用日志格式：2024-01-10 13:55:36 [INFO] message key=value
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		encoderConfig.ConsoleSeparator = " "
		fileEncoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		return nil, zap.AtomicLevel{}, fmt.Errorf("unsupported log format '%s' for '%s': %w (supported formats: json, console, text)", format, bizName, ErrInvalidLogFormat)
	}

	// 创建文件 core
	fileCore := zapcore.NewCore(fileEncoder, fileWriter, level)

	// 如果启用控制台输出，创建多核心
	var core zapcore.Core
	if cfg.Console {
		// 控制台编码器配置（使用彩色输出）
		consoleEncoderConfig := encoderConfig
		consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		consoleEncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
		consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)

		// 控制台 core
		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			level,
		)

		// 合并文件和控制台 core
		core = zapcore.NewTee(fileCore, consoleCore)
	} else {
		core = fileCore
	}

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
