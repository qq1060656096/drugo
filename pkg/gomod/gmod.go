// Package gomod 提供 Go 模块相关的工具函数，用于查找和定位 go.mod 文件及项目根目录。
package gomod

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

// FindGoModRoot 从指定目录开始向上查找包含 go.mod 文件的目录。
// 参数 start 为起始查找目录。
// 返回值为找到的项目根目录路径和是否找到的布尔值。
// 如果找到 go.mod 文件，返回该文件所在目录和 true；
// 如果一直找到文件系统根目录都未找到，返回空字符串和 false。
func FindGoModRoot(start string) (string, bool) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", false
	}

	for {
		if HasGoMod(dir) {
			return dir, true
		}

		parent := filepath.Dir(dir)
		// 已到达文件系统根目录，停止查找
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// HasGoMod 检查指定目录下是否存在 go.mod 文件。
// 参数 dir 为要检查的目录路径。
// 返回值为 true 表示存在 go.mod 文件，false 表示不存在。
func HasGoMod(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "go.mod"))
	return !os.IsNotExist(err)
}

// ProjectRoot 获取项目根目录。
// 从 runDir 开始向上查找 go.mod 文件，返回找到的项目根目录。
// 如果未找到 go.mod 文件，则返回 runDir 本身作为项目根目录。
func ProjectRoot(runDir string) string {
	dir, ok := FindGoModRoot(runDir)
	if !ok {
		return runDir
	}
	return dir
}

// ModuleName 获取指定目录下 go.mod 文件中的 module 名称。
// 参数 dir 为包含 go.mod 文件的目录路径。
// 返回值为 module 名称和可能的错误。
// 如果目录下不存在 go.mod 文件或文件读取失败，返回相应错误。
func ModuleName(dir string) (string, error) {
	gomodPath := filepath.Join(dir, "go.mod")
	content, err := os.ReadFile(gomodPath)
	if err != nil {
		return "", err
	}
	return ParseModuleName(content)
}

// ParseModuleName 从 go.mod 文件内容中解析 module 名称。
// 参数 content 为 go.mod 文件的原始内容。
// 返回值为解析出的 module 名称和可能的错误。
// 如果内容中不包含有效的 module 声明，返回空字符串和错误。
func ParseModuleName(content []byte) (string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		// 查找 module 声明行
		if strings.HasPrefix(line, "module") {
			// 移除 "module " 前缀并提取模块名
			moduleLine := strings.TrimPrefix(line, "module")
			moduleName := strings.TrimSpace(moduleLine)
			// 移除可能的行尾注释
			if idx := strings.Index(moduleName, "//"); idx != -1 {
				moduleName = strings.TrimSpace(moduleName[:idx])
			}
			if moduleName == "" {
				return "", &ModuleNameNotFoundError{}
			}
			return moduleName, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", &ModuleNameNotFoundError{}
}

// ModuleNameNotFoundError 表示在 go.mod 文件中未找到 module 声明的错误。
type ModuleNameNotFoundError struct{}

func (e *ModuleNameNotFoundError) Error() string {
	return "module name not found in go.mod"
}
