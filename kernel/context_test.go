package kernel

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestWithContext 测试 WithContext 函数
func TestWithContext(t *testing.T) {
	// 创建模拟内核
	kernel := NewMockKernel()

	// 创建上下文
	ctx := context.Background()

	// 使用 WithContext 将内核添加到上下文
	ctxWithKernel := WithContext(ctx, kernel)

	// 验证上下文不为空
	assert.NotNil(t, ctxWithKernel, "WithContext 应该返回非空的上下文")

	// 验证可以从上下文中获取内核
	retrievedKernel, ok := FromContext(ctxWithKernel)
	assert.True(t, ok, "应该能够从上下文中获取内核")
	assert.Equal(t, kernel, retrievedKernel, "获取的内核应该与设置的内核相同")
}

// TestWithContext_NilKernel 测试传入 nil 内核的情况
func TestWithContext_NilKernel(t *testing.T) {
	ctx := context.Background()

	// 传入 nil 内核
	ctxWithKernel := WithContext(ctx, nil)

	// 验证从上下文中获取内核时，由于 nil 不满足 Kernel 接口类型断言，返回 false
	retrievedKernel, ok := FromContext(ctxWithKernel)
	assert.False(t, ok, "当存储 nil 时，类型断言应该失败")
	assert.Nil(t, retrievedKernel, "获取的内核应该为 nil")
}

// TestFromContext 测试 FromContext 函数
func TestFromContext(t *testing.T) {
	// 测试从包含内核的上下文中获取
	kernel := NewMockKernel()
	ctx := WithContext(context.Background(), kernel)

	retrievedKernel, ok := FromContext(ctx)
	assert.True(t, ok, "应该能够从上下文中获取内核")
	assert.Equal(t, kernel, retrievedKernel, "获取的内核应该与设置的内核相同")

	// 测试从不包含内核的上下文中获取
	ctxWithoutKernel := context.Background()
	retrievedKernel, ok = FromContext(ctxWithoutKernel)
	assert.False(t, ok, "不应该能够从普通上下文中获取内核")
	assert.Nil(t, retrievedKernel, "获取的内核应该为 nil")
}

// TestMustFromContext 测试 MustFromContext 函数
func TestMustFromContext(t *testing.T) {
	// 测试成功获取
	kernel := NewMockKernel()
	ctx := WithContext(context.Background(), kernel)

	retrievedKernel := MustFromContext(ctx)
	assert.Equal(t, kernel, retrievedKernel, "MustFromContext 应该返回正确的内核")

	// 测试上下文中没有内核时的 panic 行为
	ctxWithoutKernel := context.Background()
	assert.Panics(t, func() {
		MustFromContext(ctxWithoutKernel)
	}, "当上下文中没有内核时，MustFromContext 应该 panic")
}

// TestServiceFromContext 测试 ServiceFromContext 函数
func TestServiceFromContext(t *testing.T) {
	// 设置测试环境
	kernel := NewMockKernel()
	service := NewMockService("test-service")
	kernel.Container().Bind("test-service", service)

	ctx := WithContext(context.Background(), kernel)

	// 测试成功获取服务
	retrievedService, err := ServiceFromContext[*MockService](ctx, "test-service")
	assert.NoError(t, err, "获取服务不应该出错")
	assert.Equal(t, service, retrievedService, "获取的服务应该与绑定的服务相同")

	// 测试获取不存在的服务
	_, err = ServiceFromContext[*MockService](ctx, "non-existent-service")
	assert.Error(t, err, "获取不存在的服务应该出错")
	assert.True(t, IsServiceNotFound(err), "应该是服务未找到错误")

	// 测试类型不匹配
	_, err = ServiceFromContext[string](ctx, "test-service")
	assert.Error(t, err, "类型不匹配应该出错")
	assert.True(t, IsServiceType(err), "应该是服务类型错误")
}

// TestServiceFromContext_NoKernel 测试上下文中没有内核的情况
func TestServiceFromContext_NoKernel(t *testing.T) {
	ctx := context.Background()

	// 测试从没有内核的上下文获取服务
	_, err := ServiceFromContext[*MockService](ctx, "test-service")
	assert.Error(t, err, "从没有内核的上下文获取服务应该出错")
	assert.True(t, IsKernelError(err), "应该是内核错误")
}

// TestServiceFromContext_NilKernel 测试内核为 nil 的情况
func TestServiceFromContext_NilKernel(t *testing.T) {
	ctx := WithContext(context.Background(), nil)

	// 测试从内核为 nil 的上下文获取服务
	_, err := ServiceFromContext[*MockService](ctx, "test-service")
	assert.Error(t, err, "从内核为 nil 的上下文获取服务应该出错")
	assert.True(t, IsKernelError(err), "应该是内核错误")
}

// TestMustServiceFromContext 测试 MustServiceFromContext 函数
func TestMustServiceFromContext(t *testing.T) {
	// 设置测试环境
	kernel := NewMockKernel()
	service := NewMockService("test-service")
	kernel.Container().Bind("test-service", service)

	ctx := WithContext(context.Background(), kernel)

	// 测试成功获取服务
	retrievedService := MustServiceFromContext[*MockService](ctx, "test-service")
	assert.Equal(t, service, retrievedService, "MustServiceFromContext 应该返回正确的服务")

	// 测试获取不存在的服务时的 panic 行为
	assert.Panics(t, func() {
		MustServiceFromContext[*MockService](ctx, "non-existent-service")
	}, "获取不存在的服务时应该 panic")

	// 测试类型不匹配时的 panic 行为
	assert.Panics(t, func() {
		MustServiceFromContext[string](ctx, "test-service")
	}, "类型不匹配时应该 panic")

	// 测试上下文中没有内核时的 panic 行为
	ctxWithoutKernel := context.Background()
	assert.Panics(t, func() {
		MustServiceFromContext[*MockService](ctxWithoutKernel, "test-service")
	}, "上下文中没有内核时应该 panic")
}

// TestContextKeyUniqueness 测试上下文键的唯一性
func TestContextKeyUniqueness(t *testing.T) {
	// 创建两个不同的内核
	kernel1 := NewMockKernel()
	kernel2 := NewMockKernel()

	// 在第一个内核中绑定一个特定的服务
	service1 := NewMockService("service1")
	kernel1.Container().Bind("service1", service1)

	// 在第二个内核中绑定一个不同的服务
	service2 := NewMockService("service2")
	kernel2.Container().Bind("service2", service2)

	// 分别创建上下文
	ctx1 := WithContext(context.Background(), kernel1)
	ctx2 := WithContext(context.Background(), kernel2)

	// 验证每个上下文返回正确的内核
	_, ok1 := FromContext(ctx1)
	_, ok2 := FromContext(ctx2)

	assert.True(t, ok1, "应该能够从第一个上下文获取内核")
	assert.True(t, ok2, "应该能够从第二个上下文获取内核")

	// 验证上下文隔离：第一个上下文只能访问第一个内核的服务
	retrievedService1, err1 := ServiceFromContext[*MockService](ctx1, "service1")
	assert.NoError(t, err1, "应该能够从第一个上下文获取 service1")
	assert.Equal(t, service1, retrievedService1, "获取的服务应该正确")

	// 第一个上下文不应该能访问第二个内核的服务
	_, err2 := ServiceFromContext[*MockService](ctx1, "service2")
	assert.Error(t, err2, "第一个上下文不应该能访问 service2")
	assert.True(t, IsServiceNotFound(err2), "应该是服务未找到错误")
}

// TestContextChain 测试上下文链式传递
func TestContextChain(t *testing.T) {
	// 创建原始上下文
	originalCtx := context.Background()

	// 添加值到原始上下文
	valueCtx := context.WithValue(originalCtx, "test-key", "test-value")

	// 添加内核到上下文
	kernel := NewMockKernel()
	kernelCtx := WithContext(valueCtx, kernel)

	// 验证原始值仍然存在
	assert.Equal(t, "test-value", kernelCtx.Value("test-key"), "原始上下文的值应该保留")

	// 验证内核可以获取
	retrievedKernel, ok := FromContext(kernelCtx)
	assert.True(t, ok, "应该能够获取内核")
	assert.Equal(t, kernel, retrievedKernel, "获取的内核应该正确")
}

// TestContextCancellation 测试上下文取消行为
func TestContextCancellation(t *testing.T) {
	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 添加内核
	kernel := NewMockKernel()
	ctxWithKernel := WithContext(ctx, kernel)

	// 验证内核可以获取
	retrievedKernel, ok := FromContext(ctxWithKernel)
	assert.True(t, ok, "取消前应该能够获取内核")
	assert.Equal(t, kernel, retrievedKernel, "获取的内核应该正确")

	// 取消上下文
	cancel()

	// 验证上下文已取消
	select {
	case <-ctxWithKernel.Done():
		// 上下文已取消，这是预期的
	default:
		t.Error("上下文应该已被取消")
	}

	// 验证即使上下文取消，内核仍然可以获取
	retrievedKernel, ok = FromContext(ctxWithKernel)
	assert.True(t, ok, "取消后仍然应该能够获取内核")
	assert.Equal(t, kernel, retrievedKernel, "获取的内核应该仍然正确")
}

// TestConcurrentAccess 测试并发访问安全性
func TestConcurrentAccess(t *testing.T) {
	kernel := NewMockKernel()
	ctx := WithContext(context.Background(), kernel)

	// 并发读取内核
	const numGoroutines = 100
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()

			// 多次读取内核
			for j := 0; j < 10; j++ {
				retrievedKernel, ok := FromContext(ctx)
				assert.True(t, ok, "应该能够获取内核")
				assert.Equal(t, kernel, retrievedKernel, "获取的内核应该正确")

				// 测试 MustFromContext
				mustKernel := MustFromContext(ctx)
				assert.Equal(t, kernel, mustKernel, "MustFromContext 应该返回正确的内核")
			}
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			// ok
		case <-time.After(5 * time.Second):
			t.Fatal("并发测试超时")
		}
	}
}

// BenchmarkWithContext 性能测试：WithContext 函数
func BenchmarkWithContext(b *testing.B) {
	kernel := NewMockKernel()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = WithContext(ctx, kernel)
	}
}

// BenchmarkFromContext 性能测试：FromContext 函数
func BenchmarkFromContext(b *testing.B) {
	kernel := NewMockKernel()
	ctx := WithContext(context.Background(), kernel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FromContext(ctx)
	}
}

// BenchmarkMustFromContext 性能测试：MustFromContext 函数
func BenchmarkMustFromContext(b *testing.B) {
	kernel := NewMockKernel()
	ctx := WithContext(context.Background(), kernel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MustFromContext(ctx)
	}
}

// BenchmarkServiceFromContext 性能测试：ServiceFromContext 函数
func BenchmarkServiceFromContext(b *testing.B) {
	kernel := NewMockKernel()
	service := NewMockService("benchmark-service")
	kernel.Container().Bind("benchmark-service", service)

	ctx := WithContext(context.Background(), kernel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ServiceFromContext[*MockService](ctx, "benchmark-service")
	}
}

// BenchmarkMustServiceFromContext 性能测试：MustServiceFromContext 函数
func BenchmarkMustServiceFromContext(b *testing.B) {
	kernel := NewMockKernel()
	service := NewMockService("benchmark-service")
	kernel.Container().Bind("benchmark-service", service)

	ctx := WithContext(context.Background(), kernel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MustServiceFromContext[*MockService](ctx, "benchmark-service")
	}
}
