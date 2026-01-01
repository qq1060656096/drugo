// Package config 提供配置管理功能，支持从 YAML 文件加载多业务配置。
package config

import "errors"

// 配置操作的哨兵错误。
var (
	// ErrNotFound 表示请求的配置不存在。
	ErrNotFound = errors.New("config: not found")

	// ErrDirRead 表示读取配置目录失败。
	ErrDirRead = errors.New("config: directory read failed")

	// ErrFileRead 表示读取配置文件失败。
	ErrFileRead = errors.New("config: file read failed")

	// ErrDuplicateKey 表示检测到重复的配置键。
	ErrDuplicateKey = errors.New("config: duplicate key")
)

// IsNotFound 判断错误是否为配置不存在错误。
// 它使用 errors.Is 进行判断，因此可以正确处理包装的错误。
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsDirRead 判断错误是否为目录读取失败错误。
// 它使用 errors.Is 进行判断，因此可以正确处理包装的错误。
func IsDirRead(err error) bool {
	return errors.Is(err, ErrDirRead)
}

// IsFileRead 判断错误是否为文件读取失败错误。
// 它使用 errors.Is 进行判断，因此可以正确处理包装的错误。
func IsFileRead(err error) bool {
	return errors.Is(err, ErrFileRead)
}

// IsDuplicateKey 判断错误是否为重复键错误。
// 它使用 errors.Is 进行判断，因此可以正确处理包装的错误。
func IsDuplicateKey(err error) bool {
	return errors.Is(err, ErrDuplicateKey)
}
