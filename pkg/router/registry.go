package router

import (
	"sync"

	"github.com/gin-gonic/gin"
)

// Registry 是一个函数注册表，注册的函数会在 Setup 时统一执行。
type Registry[T any] struct {
	mu sync.Mutex
	fs []func(T)
}

// New 创建一个新的 Registry
func New[T any]() *Registry[T] {
	return &Registry[T]{}
}

// Register 添加一个注册函数
func (r *Registry[T]) Register(f func(T)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fs = append(r.fs, f)
}

// Setup 执行所有注册函数，将 p 透传给每个函数
func (r *Registry[T]) Setup(p T) {
	r.mu.Lock()
	fs := make([]func(T), len(r.fs))
	copy(fs, r.fs) // 拷贝一份，避免在执行时被修改
	r.mu.Unlock()

	for _, f := range fs {
		f(p)
	}
}

// defaultRegistry 是默认的注册表实例，用于存放所有注册的路由
// 使用泛型指定注册的对象类型为 *gin.Engine
var defaultRegistry = New[*gin.Engine]()

// Default 返回默认的注册表实例
// 在应用初始化或 main 函数中可以通过它统一获取注册的路由
func Default() *Registry[*gin.Engine] {
	return defaultRegistry
}
