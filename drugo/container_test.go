package drugo

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/qq1060656096/drugo/kernel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockContainerService 是一个用于测试容器的模拟服务实现
type mockContainerService struct {
	name string
}

func (m *mockContainerService) Name() string {
	return m.name
}

func (m *mockContainerService) Boot(ctx context.Context) error {
	return nil
}

func (m *mockContainerService) Close(ctx context.Context) error {
	return nil
}

// TestNewContainer 测试容器创建
func TestNewContainer(t *testing.T) {
	container := NewContainer[kernel.Service]()

	require.NotNil(t, container)
	assert.NotNil(t, container.services)
	assert.Empty(t, container.services)
	assert.Empty(t, container.servicesIds)
}

// TestContainer_Bind 测试服务绑定功能
func TestContainer_Bind(t *testing.T) {
	container := NewContainer[kernel.Service]()
	service1 := &mockContainerService{name: "service1"}
	service2 := &mockContainerService{name: "service2"}

	// 测试绑定新服务
	container.Bind("service1", service1)
	assert.Equal(t, service1, container.services["service1"])
	assert.Contains(t, container.servicesIds, "service1")
	assert.Len(t, container.servicesIds, 1)

	// 测试覆盖已存在的服务
	service1New := &mockContainerService{name: "service1-new"}
	container.Bind("service1", service1New)
	assert.Equal(t, service1New, container.services["service1"])
	assert.Len(t, container.servicesIds, 1) // 服务ID列表长度不应增加

	// 测试绑定第二个服务
	container.Bind("service2", service2)
	assert.Equal(t, service2, container.services["service2"])
	assert.Contains(t, container.servicesIds, "service2")
	assert.Len(t, container.servicesIds, 2)
}

// TestContainer_Bind_EmptyName 测试绑定空名称服务
func TestContainer_Bind_EmptyName(t *testing.T) {
	container := NewContainer[kernel.Service]()
	service := &mockContainerService{name: "empty-service"}

	container.Bind("", service)
	assert.Equal(t, service, container.services[""])
	assert.Contains(t, container.servicesIds, "")
}

// TestContainer_Get 测试服务获取功能
func TestContainer_Get(t *testing.T) {
	container := NewContainer[kernel.Service]()
	service := &mockContainerService{name: "test-service"}

	// 测试获取不存在的服务
	result, err := container.Get("nonexistent")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, kernel.ErrServiceNotFound))
	var zeroService kernel.Service
	assert.Equal(t, zeroService, result)

	// 绑定服务后测试获取
	container.Bind("test-service", service)
	result, err = container.Get("test-service")
	assert.NoError(t, err)
	assert.Equal(t, service, result)
}

// TestContainer_Get_EmptyName 测试获取空名称服务
func TestContainer_Get_EmptyName(t *testing.T) {
	container := NewContainer[kernel.Service]()
	service := &mockContainerService{name: "empty-service"}

	// 绑定空名称服务
	container.Bind("", service)

	// 获取空名称服务
	result, err := container.Get("")
	assert.NoError(t, err)
	assert.Equal(t, service, result)
}

// TestContainer_MustGet 测试必须获取服务功能
func TestContainer_MustGet(t *testing.T) {
	container := NewContainer[kernel.Service]()
	service := &mockContainerService{name: "must-service"}

	// 绑定服务
	container.Bind("must-service", service)

	// 测试获取存在的服务
	result := container.MustGet("must-service")
	assert.Equal(t, service, result)

	// 测试获取不存在的服务会panic
	assert.Panics(t, func() {
		container.MustGet("nonexistent")
	})
}

// TestContainer_MustGet_EmptyName 测试必须获取空名称服务
func TestContainer_MustGet_EmptyName(t *testing.T) {
	container := NewContainer[kernel.Service]()
	service := &mockContainerService{name: "empty-service"}

	// 绑定空名称服务
	container.Bind("", service)

	// 获取空名称服务
	result := container.MustGet("")
	assert.Equal(t, service, result)
}

// TestContainer_Services 测试获取所有服务功能
func TestContainer_Services(t *testing.T) {
	container := NewContainer[kernel.Service]()
	service1 := &mockContainerService{name: "service1"}
	service2 := &mockContainerService{name: "service2"}
	service3 := &mockContainerService{name: "service3"}

	// 测试空容器
	services := container.Services()
	assert.Empty(t, services)

	// 绑定服务
	container.Bind("service1", service1)
	container.Bind("service2", service2)
	container.Bind("service3", service3)

	// 获取所有服务
	services = container.Services()
	assert.Len(t, services, 3)
	assert.Equal(t, service1, services[0])
	assert.Equal(t, service2, services[1])
	assert.Equal(t, service3, services[2])

	// 验证返回的是副本，修改不影响原容器
	services[0] = nil
	assert.Equal(t, service1, container.services["service1"])
}

// TestContainer_Services_WithOverride 测试服务覆盖后的服务列表
func TestContainer_Services_WithOverride(t *testing.T) {
	container := NewContainer[kernel.Service]()
	service1 := &mockContainerService{name: "service1"}
	service1New := &mockContainerService{name: "service1-new"}
	service2 := &mockContainerService{name: "service2"}

	// 绑定服务
	container.Bind("service1", service1)
	container.Bind("service2", service2)

	// 覆盖服务
	container.Bind("service1", service1New)

	// 验证服务列表
	services := container.Services()
	assert.Len(t, services, 2)
	assert.Equal(t, service1New, services[0]) // 应该是新的服务
	assert.Equal(t, service2, services[1])
}

// TestContainer_Names 测试获取所有服务名称功能
func TestContainer_Names(t *testing.T) {
	container := NewContainer[kernel.Service]()
	service := &mockContainerService{name: "test-service"}

	// 测试空容器
	names := container.Names()
	assert.Empty(t, names)

	// 绑定服务
	container.Bind("service1", service)
	container.Bind("service2", service)
	container.Bind("service3", service)

	// 获取所有名称
	names = container.Names()
	assert.Len(t, names, 3)
	assert.Equal(t, "service1", names[0])
	assert.Equal(t, "service2", names[1])
	assert.Equal(t, "service3", names[2])

	// 验证返回的是副本，修改不影响原容器
	names[0] = "modified"
	assert.Equal(t, "service1", container.servicesIds[0])
}

// TestContainer_Names_EmptyName 测试包含空名称的服务名称列表
func TestContainer_Names_EmptyName(t *testing.T) {
	container := NewContainer[kernel.Service]()
	service := &mockContainerService{name: "empty-service"}

	// 绑定空名称服务
	container.Bind("", service)
	container.Bind("normal", service)

	// 获取名称列表
	names := container.Names()
	assert.Len(t, names, 2)
	assert.Equal(t, "", names[0])
	assert.Equal(t, "normal", names[1])
}

// TestContainer_ConcurrentAccess 测试并发访问安全性
func TestContainer_ConcurrentAccess(t *testing.T) {
	container := NewContainer[kernel.Service]()
	service := &mockContainerService{name: "concurrent-service"}

	// 使用 WaitGroup 来协调并发操作
	var wg sync.WaitGroup
	numGoroutines := 100

	// 并发绑定服务
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			serviceName := "service-" + string(rune(id))
			container.Bind(serviceName, service)
		}(i)
	}

	// 并发读取服务
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			container.Services()
			container.Names()
		}()
	}

	wg.Wait()

	// 验证所有服务都被正确绑定
	assert.Len(t, container.services, numGoroutines)
	assert.Len(t, container.servicesIds, numGoroutines)
}

// TestContainer_ConcurrentBindAndGet 测试并发绑定和获取
func TestContainer_ConcurrentBindAndGet(t *testing.T) {
	container := NewContainer[kernel.Service]()
	service := &mockContainerService{name: "concurrent-service"}

	var wg sync.WaitGroup
	numGoroutines := 50

	// 先绑定一些服务
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			serviceName := "service-" + string(rune(id))
			container.Bind(serviceName, service)
		}(i)
	}

	wg.Wait()

	// 然后并发获取这些服务
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			serviceName := "service-" + string(rune(id))
			_, err := container.Get(serviceName)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()
}

// TestContainer_OrderPreservation 测试服务注册顺序保持
func TestContainer_OrderPreservation(t *testing.T) {
	container := NewContainer[kernel.Service]()
	service := &mockContainerService{name: "order-service"}

	// 按特定顺序绑定服务
	bindOrder := []string{"zebra", "apple", "banana", "cherry"}
	for _, name := range bindOrder {
		container.Bind(name, service)
	}

	// 验证服务列表保持注册顺序
	services := container.Services()
	assert.Len(t, services, 4)
	for i, name := range bindOrder {
		result, _ := container.Get(name)
		assert.Equal(t, result, services[i])
	}

	// 验证名称列表保持注册顺序
	names := container.Names()
	assert.Equal(t, bindOrder, names)
}

// TestContainer_InterfaceCompliance 测试接口合规性
func TestContainer_InterfaceCompliance(t *testing.T) {
	// 这个测试验证编译时接口检查是否有效
	// 如果 Container 没有正确实现 kernel.Container 接口，编译会失败
	var _ kernel.Container[kernel.Service] = (*Container[kernel.Service])(nil)

	container := NewContainer[kernel.Service]()
	service := &mockContainerService{name: "interface-service"}

	// 测试所有接口方法
	container.Bind("test", service)

	svc, err := container.Get("test")
	assert.NoError(t, err)
	assert.Equal(t, service, svc)

	svc = container.MustGet("test")
	assert.Equal(t, service, svc)

	services := container.Services()
	assert.Len(t, services, 1)
	assert.Equal(t, service, services[0])

	names := container.Names()
	assert.Equal(t, []string{"test"}, names)
}

// BenchmarkContainer_Bind 测试绑定服务的性能
func BenchmarkContainer_Bind(b *testing.B) {
	container := NewContainer[kernel.Service]()
	service := &mockContainerService{name: "benchmark-service"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		container.Bind("service", service)
	}
}

// BenchmarkContainer_Get 测试获取服务的性能
func BenchmarkContainer_Get(b *testing.B) {
	container := NewContainer[kernel.Service]()
	service := &mockContainerService{name: "benchmark-service"}
	container.Bind("service", service)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = container.Get("service")
	}
}

// BenchmarkContainer_Services 测试获取所有服务的性能
func BenchmarkContainer_Services(b *testing.B) {
	container := NewContainer[kernel.Service]()
	service := &mockContainerService{name: "benchmark-service"}

	// 预先绑定一些服务
	for i := 0; i < 100; i++ {
		container.Bind("service-"+string(rune(i)), service)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = container.Services()
	}
}

// BenchmarkContainer_Names 测试获取所有服务名称的性能
func BenchmarkContainer_Names(b *testing.B) {
	container := NewContainer[kernel.Service]()
	service := &mockContainerService{name: "benchmark-service"}

	// 预先绑定一些服务
	for i := 0; i < 100; i++ {
		container.Bind("service-"+string(rune(i)), service)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = container.Names()
	}
}

// BenchmarkContainer_ConcurrentAccess 测试并发访问性能
func BenchmarkContainer_ConcurrentAccess(b *testing.B) {
	container := NewContainer[kernel.Service]()
	service := &mockContainerService{name: "benchmark-service"}

	// 预先绑定一些服务
	for i := 0; i < 10; i++ {
		container.Bind("service-"+string(rune(i)), service)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = container.Get("service-0")
			_ = container.Services()
			_ = container.Names()
		}
	})
}
