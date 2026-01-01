package drugo

import "github.com/qq1060656096/drugo/kernel"

// GetService 从 Kernel 中获取指定名称和类型的服务。
// 它是 kernel.GetService 的门面封装，保证用户只依赖 drugo 包。
func GetService[T any](k kernel.Kernel, name string) (T, error) {
	return kernel.GetService[T](k, name)
}

// MustGetService 从 Kernel 中获取指定名称和类型的服务。
// 如果获取失败，会直接 panic。
// 它是 kernel.MustGetService 的门面封装，用于保持用户侧 API 统一。
func MustGetService[T any](k kernel.Kernel, name string) T {
	return kernel.MustGetService[T](k, name)
}
