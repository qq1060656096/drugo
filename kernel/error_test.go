package kernel

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestErrorVariables 测试预定义的错误变量
func TestErrorVariables(t *testing.T) {
	// 验证所有预定义错误变量都不为空
	assert.NotNil(t, ErrServiceNotFound, "ErrServiceNotFound 不应为空")
	assert.NotNil(t, ErrKernelNotInContext, "ErrKernelNotInContext 不应为空")
	assert.NotNil(t, ErrServiceInitFailed, "ErrServiceInitFailed 不应为空")
	assert.NotNil(t, ErrServiceRunFailed, "ErrServiceRunFailed 不应为空")
	assert.NotNil(t, ErrServiceCloseFailed, "ErrServiceCloseFailed 不应为空")
	assert.NotNil(t, ErrServiceType, "ErrServiceType 不应为空")

	// 验证错误消息格式
	assert.Equal(t, "kernel: service not found", ErrServiceNotFound.Error())
	assert.Equal(t, "kernel: kernel not found in context", ErrKernelNotInContext.Error())
	assert.Equal(t, "kernel: service initialization failed", ErrServiceInitFailed.Error())
	assert.Equal(t, "kernel: service run failed", ErrServiceRunFailed.Error())
	assert.Equal(t, "kernel: service close failed", ErrServiceCloseFailed.Error())
	assert.Equal(t, "kernel: service type mismatch", ErrServiceType.Error())
}

// TestIsKernelError 测试 IsKernelError 函数
func TestIsKernelError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil 错误",
			err:      nil,
			expected: false,
		},
		{
			name:     "服务未找到错误",
			err:      ErrServiceNotFound,
			expected: true,
		},
		{
			name:     "内核不在上下文错误",
			err:      ErrKernelNotInContext,
			expected: true,
		},
		{
			name:     "服务初始化失败错误",
			err:      ErrServiceInitFailed,
			expected: true,
		},
		{
			name:     "服务运行失败错误",
			err:      ErrServiceRunFailed,
			expected: true,
		},
		{
			name:     "服务关闭失败错误",
			err:      ErrServiceCloseFailed,
			expected: true,
		},
		{
			name:     "服务类型错误",
			err:      ErrServiceType,
			expected: true,
		},
		{
			name:     "包装的内核错误",
			err:      NewError("test.op", ErrServiceNotFound),
			expected: true,
		},
		{
			name:     "非内核错误",
			err:      errors.New("其他错误"),
			expected: false,
		},
		{
			name:     "包装的非内核错误",
			err:      NewError("test.op", errors.New("其他错误")),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsKernelError(tt.err)
			assert.Equal(t, tt.expected, result, "IsKernelError(%v) 应该返回 %v", tt.err, tt.expected)
		})
	}
}

// TestIsServiceNotFound 测试 IsServiceNotFound 函数
func TestIsServiceNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil 错误",
			err:      nil,
			expected: false,
		},
		{
			name:     "服务未找到错误",
			err:      ErrServiceNotFound,
			expected: true,
		},
		{
			name:     "包装的服务未找到错误",
			err:      NewError("test.service", ErrServiceNotFound),
			expected: true,
		},
		{
			name:     "其他内核错误",
			err:      ErrServiceType,
			expected: false,
		},
		{
			name:     "非内核错误",
			err:      errors.New("其他错误"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsServiceNotFound(tt.err)
			assert.Equal(t, tt.expected, result, "IsServiceNotFound(%v) 应该返回 %v", tt.err, tt.expected)
		})
	}
}

// TestIsServiceInitFailed 测试 IsServiceInitFailed 函数
func TestIsServiceInitFailed(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil 错误",
			err:      nil,
			expected: false,
		},
		{
			name:     "服务初始化失败错误",
			err:      ErrServiceInitFailed,
			expected: true,
		},
		{
			name:     "包装的服务初始化失败错误",
			err:      NewError("test.service", ErrServiceInitFailed),
			expected: true,
		},
		{
			name:     "其他内核错误",
			err:      ErrServiceNotFound,
			expected: false,
		},
		{
			name:     "非内核错误",
			err:      errors.New("其他错误"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsServiceInitFailed(tt.err)
			assert.Equal(t, tt.expected, result, "IsServiceInitFailed(%v) 应该返回 %v", tt.err, tt.expected)
		})
	}
}

// TestIsServiceRunFailed 测试 IsServiceRunFailed 函数
func TestIsServiceRunFailed(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil 错误",
			err:      nil,
			expected: false,
		},
		{
			name:     "服务运行失败错误",
			err:      ErrServiceRunFailed,
			expected: true,
		},
		{
			name:     "包装的服务运行失败错误",
			err:      NewError("test.service", ErrServiceRunFailed),
			expected: true,
		},
		{
			name:     "其他内核错误",
			err:      ErrServiceNotFound,
			expected: false,
		},
		{
			name:     "非内核错误",
			err:      errors.New("其他错误"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsServiceRunFailed(tt.err)
			assert.Equal(t, tt.expected, result, "IsServiceRunFailed(%v) 应该返回 %v", tt.err, tt.expected)
		})
	}
}

// TestIsServiceCloseFailed 测试 IsServiceCloseFailed 函数
func TestIsServiceCloseFailed(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil 错误",
			err:      nil,
			expected: false,
		},
		{
			name:     "服务关闭失败错误",
			err:      ErrServiceCloseFailed,
			expected: true,
		},
		{
			name:     "包装的服务关闭失败错误",
			err:      NewError("test.service", ErrServiceCloseFailed),
			expected: true,
		},
		{
			name:     "其他内核错误",
			err:      ErrServiceNotFound,
			expected: false,
		},
		{
			name:     "非内核错误",
			err:      errors.New("其他错误"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsServiceCloseFailed(tt.err)
			assert.Equal(t, tt.expected, result, "IsServiceCloseFailed(%v) 应该返回 %v", tt.err, tt.expected)
		})
	}
}

// TestIsServiceType 测试 IsServiceType 函数
func TestIsServiceType(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil 错误",
			err:      nil,
			expected: false,
		},
		{
			name:     "服务类型错误",
			err:      ErrServiceType,
			expected: true,
		},
		{
			name:     "包装的服务类型错误",
			err:      NewError("test.service", ErrServiceType),
			expected: true,
		},
		{
			name:     "其他内核错误",
			err:      ErrServiceNotFound,
			expected: false,
		},
		{
			name:     "非内核错误",
			err:      errors.New("其他错误"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsServiceType(tt.err)
			assert.Equal(t, tt.expected, result, "IsServiceType(%v) 应该返回 %v", tt.err, tt.expected)
		})
	}
}

// TestError_Struct 测试 Error 结构体
func TestError_Struct(t *testing.T) {
	originalErr := errors.New("原始错误")

	// 创建新的内核错误
	kernelErr := &Error{
		op:  "service.init",
		msg: "测试消息",
		err: originalErr,
	}

	// 测试 Error() 方法
	expectedMsg := "kernel service.init: 原始错误"
	assert.Equal(t, expectedMsg, kernelErr.Error(), "Error() 方法应该返回正确的错误消息")

	// 测试 Unwrap() 方法
	assert.Equal(t, originalErr, kernelErr.Unwrap(), "Unwrap() 应该返回原始错误")

	// 测试 nil 原始错误的情况
	nilErr := &Error{
		op:  "test.op",
		msg: "测试消息",
		err: nil,
	}
	assert.Equal(t, "kernel: <nil>", nilErr.Error(), "当原始错误为 nil 时应该返回特殊消息")
}

// TestNewError 测试 NewError 函数
func TestNewError(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		err      error
		expected error
	}{
		{
			name:     "nil 错误",
			op:       "test.op",
			err:      nil,
			expected: nil,
		},
		{
			name:     "非空错误",
			op:       "service.init",
			err:      errors.New("测试错误"),
			expected: &Error{op: "service.init", err: errors.New("测试错误")},
		},
		{
			name:     "空操作名称",
			op:       "",
			err:      errors.New("测试错误"),
			expected: &Error{op: "", err: errors.New("测试错误")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewError(tt.op, tt.err)
			if tt.expected == nil {
				assert.Nil(t, result, "当输入错误为 nil 时应该返回 nil")
			} else {
				require.NotNil(t, result, "应该返回非 nil 的错误")
				kernelErr, ok := result.(*Error)
				require.True(t, ok, "应该返回 *Error 类型")
				assert.Equal(t, tt.op, kernelErr.op, "操作名称应该匹配")
				assert.Equal(t, tt.err, kernelErr.err, "原始错误应该匹配")
			}
		})
	}
}

// TestNewServiceNotFound 测试 NewServiceNotFound 函数
func TestNewServiceNotFound(t *testing.T) {
	serviceName := "test.service"
	err := NewServiceNotFound(serviceName)

	require.NotNil(t, err, "应该返回非 nil 的错误")
	assert.True(t, IsServiceNotFound(err), "应该是服务未找到错误")
	assert.True(t, IsKernelError(err), "应该是内核错误")

	// 验证错误消息格式
	expectedMsg := "kernel " + serviceName + ": " + ErrServiceNotFound.Error()
	assert.Equal(t, expectedMsg, err.Error(), "错误消息格式应该正确")

	// 验证错误链
	assert.Equal(t, ErrServiceNotFound, errors.Unwrap(err), "错误链应该包含 ErrServiceNotFound")
}

// TestNewServiceInitFailed 测试 NewServiceInitFailed 函数
func TestNewServiceInitFailed(t *testing.T) {
	serviceName := "test.service"
	err := NewServiceInitFailed(serviceName)

	require.NotNil(t, err, "应该返回非 nil 的错误")
	assert.True(t, IsServiceInitFailed(err), "应该是服务初始化失败错误")
	assert.True(t, IsKernelError(err), "应该是内核错误")

	// 验证错误消息格式
	expectedMsg := "kernel " + serviceName + ": " + ErrServiceInitFailed.Error()
	assert.Equal(t, expectedMsg, err.Error(), "错误消息格式应该正确")

	// 验证错误链
	assert.Equal(t, ErrServiceInitFailed, errors.Unwrap(err), "错误链应该包含 ErrServiceInitFailed")
}

// TestNewServiceRunFailed 测试 NewServiceRunFailed 函数
func TestNewServiceRunFailed(t *testing.T) {
	serviceName := "test.service"
	err := NewServiceRunFailed(serviceName)

	require.NotNil(t, err, "应该返回非 nil 的错误")
	assert.True(t, IsServiceRunFailed(err), "应该是服务运行失败错误")
	assert.True(t, IsKernelError(err), "应该是内核错误")

	// 验证错误消息格式
	expectedMsg := "kernel " + serviceName + ": " + ErrServiceRunFailed.Error()
	assert.Equal(t, expectedMsg, err.Error(), "错误消息格式应该正确")

	// 验证错误链
	assert.Equal(t, ErrServiceRunFailed, errors.Unwrap(err), "错误链应该包含 ErrServiceRunFailed")
}

// TestNewServiceCloseFailed 测试 NewServiceCloseFailed 函数
func TestNewServiceCloseFailed(t *testing.T) {
	serviceName := "test.service"
	err := NewServiceCloseFailed(serviceName)

	require.NotNil(t, err, "应该返回非 nil 的错误")
	assert.True(t, IsServiceCloseFailed(err), "应该是服务关闭失败错误")
	assert.True(t, IsKernelError(err), "应该是内核错误")

	// 验证错误消息格式
	expectedMsg := "kernel " + serviceName + ": " + ErrServiceCloseFailed.Error()
	assert.Equal(t, expectedMsg, err.Error(), "错误消息格式应该正确")

	// 验证错误链
	assert.Equal(t, ErrServiceCloseFailed, errors.Unwrap(err), "错误链应该包含 ErrServiceCloseFailed")
}

// TestNewServiceType 测试 NewServiceType 函数
func TestNewServiceType(t *testing.T) {
	serviceName := "test.service"
	err := NewServiceType(serviceName)

	require.NotNil(t, err, "应该返回非 nil 的错误")
	assert.True(t, IsServiceType(err), "应该是服务类型错误")
	assert.True(t, IsKernelError(err), "应该是内核错误")

	// 验证错误消息格式
	expectedMsg := "kernel " + serviceName + ": " + ErrServiceType.Error()
	assert.Equal(t, expectedMsg, err.Error(), "错误消息格式应该正确")

	// 验证错误链
	assert.Equal(t, ErrServiceType, errors.Unwrap(err), "错误链应该包含 ErrServiceType")
}

// TestNewKernelNotInContext 测试 NewKernelNotInContext 函数
func TestNewKernelNotInContext(t *testing.T) {
	err := NewKernelNotInContext()

	require.NotNil(t, err, "应该返回非 nil 的错误")
	assert.True(t, IsKernelError(err), "应该是内核错误")

	// 验证错误消息格式
	expectedMsg := "kernel kernel: " + ErrKernelNotInContext.Error()
	assert.Equal(t, expectedMsg, err.Error(), "错误消息格式应该正确")

	// 验证错误链
	assert.Equal(t, ErrKernelNotInContext, errors.Unwrap(err), "错误链应该包含 ErrKernelNotInContext")
}

// TestError_Chain 测试错误链的行为
func TestError_Chain(t *testing.T) {
	// 创建多层错误包装
	originalErr := errors.New("底层错误")
	wrappedErr1 := NewError("layer1", originalErr)
	wrappedErr2 := NewError("layer2", wrappedErr1)

	// 测试错误链
	assert.True(t, errors.Is(wrappedErr2, originalErr), "应该能通过 errors.Is 找到底层错误")
	assert.True(t, errors.Is(wrappedErr2, wrappedErr1), "应该能通过 errors.Is 找到中间错误")

	// 测试错误解包
	unwrapped := errors.Unwrap(wrappedErr2)
	assert.Equal(t, wrappedErr1, unwrapped, "第一层解包应该得到 wrappedErr1")

	unwrapped = errors.Unwrap(unwrapped)
	assert.Equal(t, originalErr, unwrapped, "第二层解包应该得到 originalErr")
}

// TestError_EdgeCases 测试边界情况
func TestError_EdgeCases(t *testing.T) {
	t.Run("空字符串操作名称", func(t *testing.T) {
		err := NewError("", errors.New("测试"))
		require.NotNil(t, err)
		assert.Equal(t, "kernel : 测试", err.Error())
	})

	t.Run("特殊字符操作名称", func(t *testing.T) {
		err := NewError("service.init!@#$%", errors.New("测试"))
		require.NotNil(t, err)
		assert.Equal(t, "kernel service.init!@#$%: 测试", err.Error())
	})

	t.Run("长操作名称", func(t *testing.T) {
		longOp := string(make([]byte, 1000))
		for i := range longOp {
			longOp = longOp[:i] + "a" + longOp[i+1:]
		}
		err := NewError(longOp, errors.New("测试"))
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "kernel ")
		assert.Contains(t, err.Error(), ": 测试")
	})
}

// BenchmarkIsKernelError IsKernelError 性能测试
func BenchmarkIsKernelError(b *testing.B) {
	err := ErrServiceNotFound
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsKernelError(err)
	}
}

// BenchmarkIsServiceNotFound IsServiceNotFound 性能测试
func BenchmarkIsServiceNotFound(b *testing.B) {
	err := ErrServiceNotFound
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsServiceNotFound(err)
	}
}

// BenchmarkNewError NewError 性能测试
func BenchmarkNewError(b *testing.B) {
	err := errors.New("测试错误")
	op := "service.init"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewError(op, err)
	}
}

// BenchmarkNewServiceNotFound NewServiceNotFound 性能测试
func BenchmarkNewServiceNotFound(b *testing.B) {
	serviceName := "test.service"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewServiceNotFound(serviceName)
	}
}
