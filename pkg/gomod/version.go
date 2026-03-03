package gomod

import (
	"runtime/debug"
	"sync"
)

var (
	once sync.Once        // 确保只读取一次构建信息
	info *debug.BuildInfo // 缓存构建信息
)

// load 读取并缓存构建信息。
// 只会执行一次，线程安全。
func load() {
	once.Do(func() {
		if bi, ok := debug.ReadBuildInfo(); ok {
			info = bi
		}
	})
}

// MainVersion 返回主模块（main module）的版本号。
// 第二个返回值表示是否成功获取版本。
//
// 当以下情况时可能返回 false：
//   - 构建信息不可用
//   - 主模块版本为空（例如 go run 本地开发）
//
// 示例：
//
//	v, ok := gomod.Main()
func MainVersion() (string, bool) {
	load()
	if info == nil {
		return "", false
	}
	if info.Main.Version == "" {
		return "", false
	}
	return info.Main.Version, true
}

// Version 返回指定 module 路径的版本号。
// 第二个返回值表示是否找到该模块。
//
// 参数 path 必须是完整的 module 路径，例如：
//
//	"github.com/gin-gonic/gin"
//
// 规则：
//   - 如果是主模块，返回 Main.Version
//   - 如果存在 replace，则优先返回 replace 的版本
//   - 未找到或版本为空则返回 false
//
// 示例：
//
//	v, ok := gomod.Version("github.com/gin-gonic/gin")
func Version(path string) (string, bool) {
	load()
	if info == nil {
		return "", false
	}

	// 主模块
	if info.Main.Path == path {
		if info.Main.Version == "" {
			return "", false
		}
		return info.Main.Version, true
	}

	// 依赖模块
	for _, dep := range info.Deps {
		if dep.Path == path {
			// 如果存在 replace，优先返回 replace 的版本
			if dep.Replace != nil && dep.Replace.Version != "" {
				return dep.Replace.Version, true
			}
			if dep.Version != "" {
				return dep.Version, true
			}
			return "", false
		}
	}

	return "", false
}

// AllVersion 返回当前程序中所有 module 及其版本。
// 返回值为 map[modulePath]version。
//
// 如果构建信息不可用，返回 nil。
//
// 注意：
//   - 不会过滤空版本
//   - 不会对版本号做语义处理
func AllVersion() map[string]string {
	load()
	if info == nil {
		return nil
	}

	m := make(map[string]string, len(info.Deps)+1)

	// 主模块
	if info.Main.Version != "" {
		m[info.Main.Path] = info.Main.Version
	}

	// 依赖模块
	for _, dep := range info.Deps {
		if dep.Replace != nil && dep.Replace.Version != "" {
			m[dep.Path] = dep.Replace.Version
			continue
		}
		if dep.Version != "" {
			m[dep.Path] = dep.Version
		}
	}

	return m
}
