package log

import (
	"fmt"

	"go.uber.org/zap"
)

const (
	OutputTypeConsole = "console"
	OutputTypeFile    = "file"
)

var validOutputTypes = map[string]struct{}{
	OutputTypeConsole: {},
	OutputTypeFile:    {},
}

var validOutputFormats = map[string]struct{}{
	FormatJSON: {},
	FormatText: {},
}

const (
	FormatJSON = "json"
	FormatText = "text"
)

func isValidOutputType(t string) bool {
	_, ok := validOutputTypes[t]
	return ok
}

func isValidOutputFormat(f string) bool {
	_, ok := validOutputFormats[f]
	return ok
}

// Config 日志配置结构
type Config struct {
	Level   string         `yaml:"level" mapstructure:"level"`     // 日志级别: debug, info, warn, error
	Outputs []OutputConfig `yaml:"outputs" mapstructure:"outputs"` // 输出配置列表
}

// OutputConfig 单个日志输出配置
type OutputConfig struct {
	Type   string            `yaml:"type" mapstructure:"type"`     // console, file
	Format string            `yaml:"format" mapstructure:"format"` // json, text
	File   *FileOutputConfig `yaml:"file,omitempty" mapstructure:"file"`
}

// FileOutputConfig 文件输出配置
type FileOutputConfig struct {
	Dir        string `yaml:"dir" mapstructure:"dir"`                 // 日志目录
	MaxSize    int    `yaml:"max_size" mapstructure:"max_size"`       // 单个日志文件最大大小(MB)
	MaxBackups int    `yaml:"max_backups" mapstructure:"max_backups"` // 保留的旧日志文件数量
	MaxAge     int    `yaml:"max_age" mapstructure:"max_age"`         // 保留旧日志的最大天数
	Compress   bool   `yaml:"compress" mapstructure:"compress"`       // 是否压缩旧日志文件
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	if len(c.Outputs) == 0 {
		return ErrEmptyLogOutputs
	}

	if c.Level == "" {
		c.Level = "info"
	}
	if err := validateLogLevel(c.Level); err != nil {
		return err
	}

	for i := range c.Outputs {
		if err := c.Outputs[i].validateAt(i); err != nil {
			return err
		}
	}
	return nil
}

func validateLogLevel(level string) error {
	if level == "" {
		return nil
	}
	if _, err := zap.ParseAtomicLevel(level); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidLogLevel, level)
	}
	return nil
}

func (c *OutputConfig) validateAt(i int) error {
	if !isValidOutputType(c.Type) {
		return fmt.Errorf("%w: outputs[%d].type=%s", ErrInvalidOutputType, i, c.Type)
	}
	if c.Type == OutputTypeConsole && c.File != nil {
		return fmt.Errorf("%w: outputs[%d].file", ErrInvalidConfigValue, i)
	}

	if c.Format == "" {
		c.Format = FormatText
	}
	if !isValidOutputFormat(c.Format) {
		return fmt.Errorf("%w: outputs[%d].format=%s", ErrInvalidLogFormat, i, c.Format)
	}

	if c.Type != OutputTypeFile {
		return nil
	}
	if c.File == nil {
		return fmt.Errorf("%w: outputs[%d].file", ErrInvalidConfigValue, i)
	}
	if err := c.File.validateAt(i); err != nil {
		return err
	}
	return nil
}

func (f *FileOutputConfig) validateAt(i int) error {
	if f.Dir == "" {
		return fmt.Errorf("%w: outputs[%d].file.dir", ErrEmptyLogDir, i)
	}
	if f.MaxSize < 0 || f.MaxBackups < 0 || f.MaxAge < 0 {
		return fmt.Errorf("%w: outputs[%d].file", ErrInvalidConfigValue, i)
	}
	if f.MaxSize == 0 {
		f.MaxSize = 100
	}
	if f.MaxBackups == 0 {
		f.MaxBackups = 10
	}
	if f.MaxAge == 0 {
		f.MaxAge = 30
	}
	return nil
}
