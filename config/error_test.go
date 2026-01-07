package config

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestErrorVariables 测试错误变量的定义
func TestErrorVariables(t *testing.T) {
	// 测试错误变量不为 nil
	assert.NotNil(t, ErrNotFound, "ErrNotFound 不应为 nil")
	assert.NotNil(t, ErrDirRead, "ErrDirRead 不应为 nil")
	assert.NotNil(t, ErrFileRead, "ErrFileRead 不应为 nil")
	assert.NotNil(t, ErrDuplicateKey, "ErrDuplicateKey 不应为 nil")

	// 测试错误消息内容
	assert.Equal(t, "config: not found", ErrNotFound.Error(), "ErrNotFound 消息不正确")
	assert.Equal(t, "config: directory read failed", ErrDirRead.Error(), "ErrDirRead 消息不正确")
	assert.Equal(t, "config: file read failed", ErrFileRead.Error(), "ErrFileRead 消息不正确")
	assert.Equal(t, "config: duplicate key", ErrDuplicateKey.Error(), "ErrDuplicateKey 消息不正确")
}

// TestIsNotFound 测试 IsNotFound 函数
func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "直接匹配 ErrNotFound",
			err:  ErrNotFound,
			want: true,
		},
		{
			name: "包装的 ErrNotFound",
			err:  fmt.Errorf("wrapped: %w", ErrNotFound),
			want: true,
		},
		{
			name: "其他错误",
			err:  errors.New("other error"),
			want: false,
		},
		{
			name: "nil 错误",
			err:  nil,
			want: false,
		},
		{
			name: "包装其他错误",
			err:  fmt.Errorf("wrapped: %w", errors.New("other")),
			want: false,
		},
		{
			name: "多层包装的 ErrNotFound",
			err:  fmt.Errorf("level1: %w", fmt.Errorf("level2: %w", ErrNotFound)),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNotFound(tt.err)
			assert.Equal(t, tt.want, got, "IsNotFound(%v) = %v, want %v", tt.err, got, tt.want)
		})
	}
}

// TestIsDirRead 测试 IsDirRead 函数
func TestIsDirRead(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "直接匹配 ErrDirRead",
			err:  ErrDirRead,
			want: true,
		},
		{
			name: "包装的 ErrDirRead",
			err:  fmt.Errorf("wrapped: %w", ErrDirRead),
			want: true,
		},
		{
			name: "其他错误",
			err:  errors.New("other error"),
			want: false,
		},
		{
			name: "nil 错误",
			err:  nil,
			want: false,
		},
		{
			name: "包装其他错误",
			err:  fmt.Errorf("wrapped: %w", errors.New("other")),
			want: false,
		},
		{
			name: "多层包装的 ErrDirRead",
			err:  fmt.Errorf("level1: %w", fmt.Errorf("level2: %w", ErrDirRead)),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDirRead(tt.err)
			assert.Equal(t, tt.want, got, "IsDirRead(%v) = %v, want %v", tt.err, got, tt.want)
		})
	}
}

// TestIsFileRead 测试 IsFileRead 函数
func TestIsFileRead(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "直接匹配 ErrFileRead",
			err:  ErrFileRead,
			want: true,
		},
		{
			name: "包装的 ErrFileRead",
			err:  fmt.Errorf("wrapped: %w", ErrFileRead),
			want: true,
		},
		{
			name: "其他错误",
			err:  errors.New("other error"),
			want: false,
		},
		{
			name: "nil 错误",
			err:  nil,
			want: false,
		},
		{
			name: "包装其他错误",
			err:  fmt.Errorf("wrapped: %w", errors.New("other")),
			want: false,
		},
		{
			name: "多层包装的 ErrFileRead",
			err:  fmt.Errorf("level1: %w", fmt.Errorf("level2: %w", ErrFileRead)),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsFileRead(tt.err)
			assert.Equal(t, tt.want, got, "IsFileRead(%v) = %v, want %v", tt.err, got, tt.want)
		})
	}
}

// TestIsDuplicateKey 测试 IsDuplicateKey 函数
func TestIsDuplicateKey(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "直接匹配 ErrDuplicateKey",
			err:  ErrDuplicateKey,
			want: true,
		},
		{
			name: "包装的 ErrDuplicateKey",
			err:  fmt.Errorf("wrapped: %w", ErrDuplicateKey),
			want: true,
		},
		{
			name: "其他错误",
			err:  errors.New("other error"),
			want: false,
		},
		{
			name: "nil 错误",
			err:  nil,
			want: false,
		},
		{
			name: "包装其他错误",
			err:  fmt.Errorf("wrapped: %w", errors.New("other")),
			want: false,
		},
		{
			name: "多层包装的 ErrDuplicateKey",
			err:  fmt.Errorf("level1: %w", fmt.Errorf("level2: %w", ErrDuplicateKey)),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDuplicateKey(tt.err)
			assert.Equal(t, tt.want, got, "IsDuplicateKey(%v) = %v, want %v", tt.err, got, tt.want)
		})
	}
}

// TestErrorFunctions_CrossCompatibility 测试错误函数之间的交叉兼容性
func TestErrorFunctions_CrossCompatibility(t *testing.T) {
	// 确保每个判断函数只识别对应的错误类型
	assert.True(t, IsNotFound(ErrNotFound), "IsNotFound 应该识别 ErrNotFound")
	assert.False(t, IsDirRead(ErrNotFound), "IsDirRead 不应该识别 ErrNotFound")
	assert.False(t, IsFileRead(ErrNotFound), "IsFileRead 不应该识别 ErrNotFound")
	assert.False(t, IsDuplicateKey(ErrNotFound), "IsDuplicateKey 不应该识别 ErrNotFound")

	assert.True(t, IsDirRead(ErrDirRead), "IsDirRead 应该识别 ErrDirRead")
	assert.False(t, IsNotFound(ErrDirRead), "IsNotFound 不应该识别 ErrDirRead")
	assert.False(t, IsFileRead(ErrDirRead), "IsFileRead 不应该识别 ErrDirRead")
	assert.False(t, IsDuplicateKey(ErrDirRead), "IsDuplicateKey 不应该识别 ErrDirRead")

	assert.True(t, IsFileRead(ErrFileRead), "IsFileRead 应该识别 ErrFileRead")
	assert.False(t, IsNotFound(ErrFileRead), "IsNotFound 不应该识别 ErrFileRead")
	assert.False(t, IsDirRead(ErrFileRead), "IsDirRead 不应该识别 ErrFileRead")
	assert.False(t, IsDuplicateKey(ErrFileRead), "IsDuplicateKey 不应该识别 ErrFileRead")

	assert.True(t, IsDuplicateKey(ErrDuplicateKey), "IsDuplicateKey 应该识别 ErrDuplicateKey")
	assert.False(t, IsNotFound(ErrDuplicateKey), "IsNotFound 不应该识别 ErrDuplicateKey")
	assert.False(t, IsDirRead(ErrDuplicateKey), "IsDirRead 不应该识别 ErrDuplicateKey")
	assert.False(t, IsFileRead(ErrDuplicateKey), "IsFileRead 不应该识别 ErrDuplicateKey")
}

// TestErrorFunctions_RealWorldScenarios 测试真实世界场景
func TestErrorFunctions_RealWorldScenarios(t *testing.T) {
	// 模拟真实场景中的错误处理
	scenarios := []struct {
		name        string
		operation   func() error
		expectFunc  func(error) bool
		expectTrue  bool
		description string
	}{
		{
			name: "配置文件不存在",
			operation: func() error {
				return fmt.Errorf("failed to load config: %w", ErrNotFound)
			},
			expectFunc:  IsNotFound,
			expectTrue:  true,
			description: "配置文件不存在应该被 IsNotFound 识别",
		},
		{
			name: "目录权限错误",
			operation: func() error {
				return fmt.Errorf("permission denied: %w", ErrDirRead)
			},
			expectFunc:  IsDirRead,
			expectTrue:  true,
			description: "目录读取错误应该被 IsDirRead 识别",
		},
		{
			name: "文件格式错误",
			operation: func() error {
				return fmt.Errorf("invalid yaml: %w", ErrFileRead)
			},
			expectFunc:  IsFileRead,
			expectTrue:  true,
			description: "文件读取错误应该被 IsFileRead 识别",
		},
		{
			name: "配置冲突",
			operation: func() error {
				return fmt.Errorf("config conflict: %w", ErrDuplicateKey)
			},
			expectFunc:  IsDuplicateKey,
			expectTrue:  true,
			description: "重复键错误应该被 IsDuplicateKey 识别",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			err := scenario.operation()
			require.NotNil(t, err, "操作应该返回错误")

			result := scenario.expectFunc(err)
			assert.Equal(t, scenario.expectTrue, result, scenario.description)
		})
	}
}

// TestErrorFunctions_ConcurrentAccess 测试并发访问安全性
func TestErrorFunctions_ConcurrentAccess(t *testing.T) {
	const numGoroutines = 100
	const numIterations = 1000

	// 测试 IsNotFound 的并发安全性
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numIterations; j++ {
				// 交替测试不同的错误
				if j%2 == 0 {
					assert.True(t, IsNotFound(ErrNotFound), "并发访问时 IsNotFound 应该返回 true")
				} else {
					assert.False(t, IsNotFound(errors.New("other")), "并发访问时 IsNotFound 应该返回 false")
				}
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// BenchmarkIsNotFound 基准测试 IsNotFound 函数
func BenchmarkIsNotFound(b *testing.B) {
	err := ErrNotFound
	wrappedErr := fmt.Errorf("wrapped: %w", ErrNotFound)
	otherErr := errors.New("other")

	b.Run("DirectMatch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsNotFound(err)
		}
	})

	b.Run("WrappedMatch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsNotFound(wrappedErr)
		}
	})

	b.Run("NoMatch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsNotFound(otherErr)
		}
	})

	b.Run("NilError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsNotFound(nil)
		}
	})
}

// BenchmarkIsDirRead 基准测试 IsDirRead 函数
func BenchmarkIsDirRead(b *testing.B) {
	err := ErrDirRead
	wrappedErr := fmt.Errorf("wrapped: %w", ErrDirRead)
	otherErr := errors.New("other")

	b.Run("DirectMatch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsDirRead(err)
		}
	})

	b.Run("WrappedMatch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsDirRead(wrappedErr)
		}
	})

	b.Run("NoMatch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsDirRead(otherErr)
		}
	})
}

// BenchmarkIsFileRead 基准测试 IsFileRead 函数
func BenchmarkIsFileRead(b *testing.B) {
	err := ErrFileRead
	wrappedErr := fmt.Errorf("wrapped: %w", ErrFileRead)
	otherErr := errors.New("other")

	b.Run("DirectMatch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsFileRead(err)
		}
	})

	b.Run("WrappedMatch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsFileRead(wrappedErr)
		}
	})

	b.Run("NoMatch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsFileRead(otherErr)
		}
	})
}

// BenchmarkIsDuplicateKey 基准测试 IsDuplicateKey 函数
func BenchmarkIsDuplicateKey(b *testing.B) {
	err := ErrDuplicateKey
	wrappedErr := fmt.Errorf("wrapped: %w", ErrDuplicateKey)
	otherErr := errors.New("other")

	b.Run("DirectMatch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsDuplicateKey(err)
		}
	})

	b.Run("WrappedMatch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsDuplicateKey(wrappedErr)
		}
	})

	b.Run("NoMatch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsDuplicateKey(otherErr)
		}
	})
}
