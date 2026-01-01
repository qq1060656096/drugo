package drugo

import (
	"sync"

	"github.com/qq1060656096/drugo/kernel"
)

// 确保 Container 结构体完整实现了 kernel.Container 接口。
// 这是一种常见的静态编译检查技巧。
var _ kernel.Container[kernel.Service] = (*Container[kernel.Service])(nil)

// Container 是一个通用的服务容器，负责管理具有特定约束的服务实例。
// 它通过 map 提供快速查询，并通过 servicesIds 维护服务的注册顺序。
type Container[T kernel.Service] struct {
	services    map[string]T // 存储服务名称到实例的映射
	servicesIds []string     // 记录服务注册的先后顺序
	mu          sync.RWMutex // 保护并发读写的读写锁
}

// Bind 将一个服务实例绑定到指定的名称。
// 如果名称已存在，则覆盖旧实例；如果是新服务，则记录其注册顺序。
func (c *Container[T]) Bind(name string, service T) {
	c.mu.Lock()
	defer c.mu.Unlock() // 修正：必须与 Lock() 配对使用 Unlock()

	if _, ok := c.services[name]; !ok {
		c.servicesIds = append(c.servicesIds, name)
	}
	c.services[name] = service
}

// Get 根据名称获取对应的服务实例。
// 如果服务不存在，则返回 os.ErrNotExist 错误。
func (c *Container[T]) Get(name string) (T, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	svc, ok := c.services[name]
	if !ok {
		return svc, kernel.NewServiceNotFound(name)
	}
	return svc, nil
}

// MustGet 尝试获取服务实例，如果服务不存在则直接触发 panic。
// 建议仅在程序初始化等确定服务必须存在的场景下使用。
func (c *Container[T]) MustGet(name string) T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	svc, ok := c.services[name]
	if !ok {
		panic(kernel.NewServiceNotFound(name)) // 修正：panic 有意义的错误信息
	}
	return svc
}

// Services 返回当前容器中所有已注册的服务实例。
// 返回的切片顺序与服务注册（Bind）的先后顺序一致。
func (c *Container[T]) Services() []T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	services := make([]T, 0, len(c.servicesIds))
	for _, name := range c.servicesIds {
		services = append(services, c.services[name])
	}
	return services
}

// Names 返回当前容器中所有已注册的服务名称。
// 返回的切片顺序与服务注册（Bind）的先后顺序一致。
func (c *Container[T]) Names() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 返回服务名称的副本，避免外部修改影响内部状态
	names := make([]string, len(c.servicesIds))
	copy(names, c.servicesIds)
	return names
}

func NewContainer[T kernel.Service]() *Container[T] {
	return &Container[T]{
		services: make(map[string]T),
	}
}
