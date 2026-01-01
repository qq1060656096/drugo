package log

import (
	"fmt"

	"go.uber.org/zap"
)

// Config 日志配置结构
type Config struct {
	Dir        string `yaml:"dir"`         // 日志目录
	Level      string `yaml:"level"`       // 日志级别: debug, info, warn, error
	Format     string `yaml:"format"`      // 日志格式: json, console, text(或standard)
	MaxSize    int    `yaml:"max_size"`    // 单个日志文件最大大小(MB)
	MaxBackups int    `yaml:"max_backups"` // 保留的旧日志文件数量
	MaxAge     int    `yaml:"max_age"`     // 保留旧日志的最大天数
	Compress   bool   `yaml:"compress"`    // 是否压缩旧日志文件
	Console    bool   `yaml:"console"`     // 是否同时输出到控制台
}

// Validate 验证配置的有效性
func (c Config) Validate() error {
	if c.Dir == "" {
		return ErrEmptyLogDir
	}
	if c.MaxSize < 0 || c.MaxBackups < 0 || c.MaxAge < 0 {
		return ErrInvalidConfigValue
	}
	// 验证日志级别
	if c.Level != "" {
		if _, err := zap.ParseAtomicLevel(c.Level); err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidLogLevel, c.Level)
		}
	}
	// 验证日志格式
	if c.Format != "" && c.Format != "json" && c.Format != "console" && c.Format != "text" && c.Format != "standard" {
		return fmt.Errorf("%w: %s", ErrInvalidLogFormat, c.Format)
	}
	return nil
}
