package drugo

import (
	"context"
	"testing"
	"time"

	"github.com/qq1060656096/drugo/kernel"
	"github.com/stretchr/testify/assert"
)

// mockService 是一个用于测试的模拟服务实现
type mockService struct {
	name string
}

func (m *mockService) Name() string {
	return m.name
}

func (m *mockService) Boot(ctx context.Context) error {
	return nil
}

func (m *mockService) Close(ctx context.Context) error {
	return nil
}

// TestWithRoot 测试 WithRoot 选项函数
func TestWithRoot(t *testing.T) {
	tests := []struct {
		name     string
		root     string
		expected string
	}{
		{
			name:     "设置有效根目录",
			root:     "/app/root",
			expected: "/app/root",
		},
		{
			name:     "设置空根目录",
			root:     "",
			expected: "",
		},
		{
			name:     "设置相对路径根目录",
			root:     "./relative/path",
			expected: "./relative/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &options{}
			opt := WithRoot(tt.root)
			opt(opts)
			assert.Equal(t, tt.expected, opts.root)
		})
	}
}

// TestWithContext 测试 WithContext 选项函数
func TestWithContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected context.Context
	}{
		{
			name:     "设置有效上下文",
			ctx:      context.Background(),
			expected: context.Background(),
		},
		{
			name:     "设置带取消的上下文",
			ctx:      context.WithValue(context.Background(), "key", "value"),
			expected: context.WithValue(context.Background(), "key", "value"),
		},
		{
			name:     "设置nil上下文",
			ctx:      nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &options{}
			opt := WithContext(tt.ctx)
			opt(opts)
			assert.Equal(t, tt.expected, opts.ctx)
		})
	}
}

// TestWithNameService 测试 WithNameService 选项函数
func TestWithNameService(t *testing.T) {
	service := &mockService{name: "test-service"}

	tests := []struct {
		name            string
		serviceName     string
		service         kernel.Service
		initialServices []map[string]kernel.Service
		expectedLen     int
	}{
		{
			name:            "添加单个命名服务",
			serviceName:     "test",
			service:         service,
			initialServices: nil,
			expectedLen:     1,
		},
		{
			name:            "添加到现有服务列表",
			serviceName:     "another",
			service:         service,
			initialServices: []map[string]kernel.Service{{"existing": service}},
			expectedLen:     2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &options{
				services: tt.initialServices,
			}
			opt := WithNameService(tt.serviceName, tt.service)
			opt(opts)
			assert.Len(t, opts.services, tt.expectedLen)
			assert.Equal(t, tt.service, opts.services[len(opts.services)-1][tt.serviceName])
		})
	}
}

// TestWithService 测试 WithService 选项函数
func TestWithService(t *testing.T) {
	service := &mockService{name: "auto-service"}

	tests := []struct {
		name        string
		service     kernel.Service
		expectedLen int
	}{
		{
			name:        "添加单个服务",
			service:     service,
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &options{}
			opt := WithService(tt.service)
			opt(opts)
			assert.Len(t, opts.services, tt.expectedLen)
			assert.Equal(t, tt.service, opts.services[0][tt.service.Name()])
		})
	}
}

// TestWithShutdownTimeout 测试 WithShutdownTimeout 选项函数
func TestWithShutdownTimeout(t *testing.T) {
	tests := []struct {
		name     string
		timeout  time.Duration
		expected time.Duration
	}{
		{
			name:     "设置5秒超时",
			timeout:  5 * time.Second,
			expected: 5 * time.Second,
		},
		{
			name:     "设置30秒超时",
			timeout:  30 * time.Second,
			expected: 30 * time.Second,
		},
		{
			name:     "设置0秒超时",
			timeout:  0,
			expected: 0,
		},
		{
			name:     "设置1分钟超时",
			timeout:  time.Minute,
			expected: time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &options{}
			opt := WithShutdownTimeout(tt.timeout)
			opt(opts)
			assert.Equal(t, tt.expected, opts.shutdownTimeout)
		})
	}
}

// TestOptions_Combo 测试多个选项的组合使用
func TestOptions_Combo(t *testing.T) {
	service1 := &mockService{name: "service1"}
	service2 := &mockService{name: "service2"}
	ctx := context.WithValue(context.Background(), "test", "value")

	opts := &options{}

	// 应用多个选项
	options := []Option{
		WithRoot("/test/root"),
		WithContext(ctx),
		WithNameService("custom1", service1),
		WithService(service2),
		WithShutdownTimeout(15 * time.Second),
	}

	for _, opt := range options {
		opt(opts)
	}

	// 验证所有选项都正确应用
	assert.Equal(t, "/test/root", opts.root)
	assert.Equal(t, ctx, opts.ctx)
	assert.Len(t, opts.services, 2)
	assert.Equal(t, service1, opts.services[0]["custom1"])
	assert.Equal(t, service2, opts.services[1]["service2"])
	assert.Equal(t, 15*time.Second, opts.shutdownTimeout)
}

// TestOptions_DefaultValues 测试选项结构的默认值
func TestOptions_DefaultValues(t *testing.T) {
	opts := &options{}

	// 验证默认值
	assert.Empty(t, opts.root)
	assert.Nil(t, opts.ctx)
	assert.Nil(t, opts.services)
	assert.Zero(t, opts.shutdownTimeout)
}

// TestOptions_ServicesNilHandling 测试服务切片为nil时的处理
func TestOptions_ServicesNilHandling(t *testing.T) {
	service := &mockService{name: "test-service"}

	opts := &options{}
	assert.Nil(t, opts.services)

	// 添加服务应该初始化切片
	opt := WithNameService("test", service)
	opt(opts)

	assert.NotNil(t, opts.services)
	assert.Len(t, opts.services, 1)
}

// TestOptions_MultipleSameNameServices 测试添加多个同名服务
func TestOptions_MultipleSameNameServices(t *testing.T) {
	service1 := &mockService{name: "service"}
	service2 := &mockService{name: "service"}

	opts := &options{}

	// 添加两个同名服务
	opt1 := WithNameService("same", service1)
	opt2 := WithNameService("same", service2)

	opt1(opts)
	opt2(opts)

	// 应该有两个条目，即使名称相同
	assert.Len(t, opts.services, 2)
	assert.Equal(t, service1, opts.services[0]["same"])
	assert.Equal(t, service2, opts.services[1]["same"])
}

// BenchmarkWithRoot 测试 WithRoot 函数的性能
func BenchmarkWithRoot(b *testing.B) {
	opt := WithRoot("/test/root")
	opts := &options{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opt(opts)
	}
}

// BenchmarkWithContext 测试 WithContext 函数的性能
func BenchmarkWithContext(b *testing.B) {
	ctx := context.Background()
	opt := WithContext(ctx)
	opts := &options{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opt(opts)
	}
}

// BenchmarkWithNameService 测试 WithNameService 函数的性能
func BenchmarkWithNameService(b *testing.B) {
	service := &mockService{name: "benchmark-service"}
	opt := WithNameService("benchmark", service)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opts := &options{}
		opt(opts)
	}
}

// BenchmarkWithService 测试 WithService 函数的性能
func BenchmarkWithService(b *testing.B) {
	service := &mockService{name: "benchmark-service"}
	opt := WithService(service)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opts := &options{}
		opt(opts)
	}
}

// BenchmarkWithShutdownTimeout 测试 WithShutdownTimeout 函数的性能
func BenchmarkWithShutdownTimeout(b *testing.B) {
	opt := WithShutdownTimeout(10 * time.Second)
	opts := &options{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opt(opts)
	}
}

// BenchmarkOptions_Combo 测试组合选项的性能
func BenchmarkOptions_Combo(b *testing.B) {
	service := &mockService{name: "benchmark-service"}
	ctx := context.Background()

	optionList := []Option{
		WithRoot("/benchmark/root"),
		WithContext(ctx),
		WithService(service),
		WithShutdownTimeout(5 * time.Second),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opts := &options{}
		for _, opt := range optionList {
			opt(opts)
		}
	}
}
