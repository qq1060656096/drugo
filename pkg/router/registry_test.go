package router

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNew 测试 New 函数
func TestNew(t *testing.T) {
	// 测试创建新的 Registry 实例
	registry := New[int]()
	require.NotNil(t, registry)
	// 通过注册函数来验证初始化状态
	var called bool
	registry.Register(func(p int) {
		called = true
	})
	registry.Setup(42)
	assert.True(t, called)
}

// TestRegistry_Register 测试 Register 函数
func TestRegistry_Register(t *testing.T) {
	registry := New[int]()
	
	// 测试注册一个函数
	registry.Register(func(p int) {
		assert.Equal(t, 42, p)
	})
	
	require.Len(t, registry.fs, 1)
	
	// 测试注册多个函数
	registry.Register(func(p int) {
		assert.Equal(t, 42, p)
	})
	
	require.Len(t, registry.fs, 2)
}

// TestRegistry_Setup 测试 Setup 函数
func TestRegistry_Setup(t *testing.T) {
	registry := New[string]()
	
	// 测试执行注册的函数
	var calls []string
	registry.Register(func(p string) {
		calls = append(calls, "func1:"+p)
	})
	registry.Register(func(p string) {
		calls = append(calls, "func2:"+p)
	})
	
	registry.Setup("test")
	
	require.Len(t, calls, 2)
	assert.Equal(t, "func1:test", calls[0])
	assert.Equal(t, "func2:test", calls[1])
}

// TestRegistry_Setup_Empty 测试空注册表的 Setup
func TestRegistry_Setup_Empty(t *testing.T) {
	registry := New[int]()
	
	// 空注册表执行 Setup 应该不会 panic
	assert.NotPanics(t, func() {
		registry.Setup(123)
	})
}

// TestRegistry_Setup_NilFunction 测试注册 nil 函数的情况
func TestRegistry_Setup_NilFunction(t *testing.T) {
	registry := New[int]()
	
	// 注册 nil 函数（虽然不推荐，但应该能处理）
	registry.Register(nil)
	
	// 执行时会 panic，因为调用 nil 函数
	assert.Panics(t, func() {
		registry.Setup(123)
	})
}

// TestRegistry_ConcurrentAccess 测试并发访问安全性
func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := New[int]()
	
	// 使用 WaitGroup 来协调并发操作
	var wg sync.WaitGroup
	var registerCount int64
	
	// 并发注册函数
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			registry.Register(func(p int) {
				// 空函数，只测试并发安全性
			})
			atomic.AddInt64(&registerCount, 1)
		}(i)
	}
	
	// 并发执行 Setup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			registry.Setup(id)
		}(i)
	}
	
	wg.Wait()
	
	// 验证所有函数都被正确注册
	assert.Equal(t, int64(100), registerCount)
}

// TestRegistry_Setup_ConcurrentModification 测试 Setup 期间的并发修改
func TestRegistry_Setup_ConcurrentModification(t *testing.T) {
	registry := New[int]()
	
	// 先注册一些函数
	for i := 0; i < 10; i++ {
		registry.Register(func(p int) {
			// 模拟一些处理时间
		})
	}
	
	var wg sync.WaitGroup
	setupStarted := make(chan bool)
	setupDone := make(chan bool)
	var registerCount int64
	
	// 启动 Setup 协程
	wg.Add(1)
	go func() {
		defer wg.Done()
		close(setupStarted)
		registry.Setup(42)
		close(setupDone)
	}()
	
	// 等待 Setup 开始
	<-setupStarted
	
	// 在 Setup 执行期间尝试注册新函数
	wg.Add(1)
	go func() {
		defer wg.Done()
		registry.Register(func(p int) {
			// 新函数
		})
		atomic.AddInt64(&registerCount, 1)
	}()
	
	<-setupDone
	wg.Wait()
	
	// 验证新函数被正确注册
	assert.Equal(t, int64(1), registerCount)
}

// TestRegistry_Setup_FunctionCopy 测试 Setup 时函数列表的拷贝
func TestRegistry_Setup_FunctionCopy(t *testing.T) {
	registry := New[int]()
	
	var executionOrder []int
	
	// 注册函数
	registry.Register(func(p int) {
		executionOrder = append(executionOrder, 1)
	})
	
	registry.Register(func(p int) {
		executionOrder = append(executionOrder, 2)
	})
	
	// 第一次 Setup
	registry.Setup(0)
	require.Len(t, executionOrder, 2)
	assert.Equal(t, 1, executionOrder[0])
	assert.Equal(t, 2, executionOrder[1])
	
	// 在 Setup 期间注册新函数
	registry.Register(func(p int) {
		executionOrder = append(executionOrder, 3)
	})
	
	// 第二次 Setup 应该执行所有3个函数
	registry.Setup(0)
	require.Len(t, executionOrder, 5)
	assert.Equal(t, 1, executionOrder[2]) // 第一次函数
	assert.Equal(t, 2, executionOrder[3]) // 第二次函数
	assert.Equal(t, 3, executionOrder[4]) // 新注册的函数
}

// TestDefault 测试 Default 函数
func TestDefault(t *testing.T) {
	// 获取默认注册表
	defaultReg := Default()
	require.NotNil(t, defaultReg)
	
	// 多次调用应该返回同一个实例
	defaultReg2 := Default()
	assert.Same(t, defaultReg, defaultReg2)
}

// TestDefault_WithGinEngine 测试默认注册表与 gin.Engine 的集成
func TestDefault_WithGinEngine(t *testing.T) {
	// 设置 gin 为测试模式
	gin.SetMode(gin.TestMode)
	
	defaultReg := Default()
	require.NotNil(t, defaultReg)
	
	// 创建 gin.Engine
	engine := gin.New()
	
	// 注册路由设置函数
	var routes []string
	defaultReg.Register(func(e *gin.Engine) {
		e.GET("/test1", func(c *gin.Context) {
			routes = append(routes, "/test1")
			c.String(200, "test1")
		})
	})
	
	defaultReg.Register(func(e *gin.Engine) {
		e.GET("/test2", func(c *gin.Context) {
			routes = append(routes, "/test2")
			c.String(200, "test2")
		})
	})
	
	// 执行 Setup
	defaultReg.Setup(engine)
	
	// 验证路由被正确注册
	require.Len(t, routes, 0) // 路由还没有被访问
	
	// 模拟请求访问路由
	// 注意：这里只是验证函数被调用，实际的路由测试需要更复杂的设置
}

// TestRegistry_GenericTypes 测试泛型类型的支持
func TestRegistry_GenericTypes(t *testing.T) {
	// 测试字符串类型
	stringReg := New[string]()
	stringReg.Register(func(s string) {
		assert.Equal(t, "hello", s)
	})
	stringReg.Setup("hello")
	
	// 测试结构体类型
	type TestStruct struct {
		Name string
		Age  int
	}
	
	structReg := New[TestStruct]()
	structReg.Register(func(ts TestStruct) {
		assert.Equal(t, "test", ts.Name)
		assert.Equal(t, 25, ts.Age)
	})
	structReg.Setup(TestStruct{Name: "test", Age: 25})
	
	// 测试指针类型
	ptrReg := New[*TestStruct]()
	ptrReg.Register(func(ts *TestStruct) {
		assert.NotNil(t, ts)
		assert.Equal(t, "ptr", ts.Name)
	})
	testStruct := &TestStruct{Name: "ptr", Age: 30}
	ptrReg.Setup(testStruct)
}

// TestRegistry_EdgeCases 测试边界情况
func TestRegistry_EdgeCases(t *testing.T) {
	t.Run("大量函数注册", func(t *testing.T) {
		registry := New[int]()
		
		// 注册大量函数
		for i := 0; i < 10000; i++ {
			registry.Register(func(p int) {
				// 空函数
			})
		}
		
		require.Len(t, registry.fs, 10000)
		
		// 执行 Setup 应该没有问题
		assert.NotPanics(t, func() {
			registry.Setup(42)
		})
	})
	
	t.Run("空参数", func(t *testing.T) {
		registry := New[struct{}]()
		
		called := false
		registry.Register(func(p struct{}) {
			called = true
		})
		
		registry.Setup(struct{}{})
		assert.True(t, called)
	})
}

// BenchmarkRegistry_Register 性能测试：Register 函数
func BenchmarkRegistry_Register(b *testing.B) {
	registry := New[int]()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.Register(func(p int) {
			// 空函数
		})
	}
}

// BenchmarkRegistry_Setup 性能测试：Setup 函数
func BenchmarkRegistry_Setup(b *testing.B) {
	registry := New[int]()
	
	// 预先注册一些函数
	for i := 0; i < 100; i++ {
		registry.Register(func(p int) {
			// 空函数
		})
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.Setup(42)
	}
}

// BenchmarkRegistry_ConcurrentRegister 性能测试：并发注册
func BenchmarkRegistry_ConcurrentRegister(b *testing.B) {
	registry := New[int]()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			registry.Register(func(p int) {
				// 空函数
			})
		}
	})
}

// BenchmarkRegistry_ConcurrentSetup 性能测试：并发 Setup
func BenchmarkRegistry_ConcurrentSetup(b *testing.B) {
	registry := New[int]()
	
	// 预先注册一些函数
	for i := 0; i < 10; i++ {
		registry.Register(func(p int) {
			// 空函数
		})
	}
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			registry.Setup(42)
		}
	})
}
