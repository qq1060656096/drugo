package log

import "errors"

// 哨兵错误（Sentinel Errors）- 导出的预定义错误，可被外部包使用
var (
	// ErrInvalidLogLevel 无效的日志级别错误
	ErrInvalidLogLevel = errors.New("invalid log level")
	// ErrInvalidLogFormat 无效的日志格式错误
	ErrInvalidLogFormat = errors.New("invalid log format")
	// ErrEmptyBizName 业务名称为空错误
	ErrEmptyBizName = errors.New("bizName cannot be empty")
	// ErrEmptyLogDir 日志目录为空错误
	ErrEmptyLogDir = errors.New("log directory cannot be empty")
	// ErrInvalidConfigValue 无效的配置值错误
	ErrInvalidConfigValue = errors.New("invalid config value: max_size, max_backups, max_age must be non-negative")
	// ErrLoggerNotFound logger 不存在错误
	ErrLoggerNotFound = errors.New("logger not found")
)

// IsInvalidLogLevel 检查是否为无效日志级别错误
func IsInvalidLogLevel(err error) bool {
	return errors.Is(err, ErrInvalidLogLevel)
}

// IsInvalidLogFormat 检查是否为无效日志格式错误
func IsInvalidLogFormat(err error) bool {
	return errors.Is(err, ErrInvalidLogFormat)
}

// IsEmptyBizName 检查是否为空业务名称错误
func IsEmptyBizName(err error) bool {
	return errors.Is(err, ErrEmptyBizName)
}

// IsEmptyLogDir 检查是否为空日志目录错误
func IsEmptyLogDir(err error) bool {
	return errors.Is(err, ErrEmptyLogDir)
}

// IsInvalidConfigValue 检查是否为无效配置值错误
func IsInvalidConfigValue(err error) bool {
	return errors.Is(err, ErrInvalidConfigValue)
}

// IsLoggerNotFound 检查是否为 logger 不存在错误
func IsLoggerNotFound(err error) bool {
	return errors.Is(err, ErrLoggerNotFound)
}
