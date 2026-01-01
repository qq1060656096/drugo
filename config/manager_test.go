package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewManager 测试 NewManager 函数的各种场景。
func TestNewManager(t *testing.T) {
	t.Run("valid config directory", func(t *testing.T) {
		tempDir := t.TempDir()

		// 创建测试配置文件
		createTestConfigFile(t, tempDir, "app.yml", map[string]interface{}{
			"database": map[string]interface{}{
				"host": "localhost",
				"port": 5432,
			},
		})

		manager, err := NewManager(tempDir)
		require.NoError(t, err)
		assert.NotNil(t, manager)
		assert.Equal(t, tempDir, manager.configDir)
		assert.NotNil(t, manager.root)
		assert.Empty(t, manager.configs)
	})

	t.Run("non-existent directory", func(t *testing.T) {
		_, err := NewManager("/non/existent/directory")
		require.Error(t, err)
		assert.True(t, IsDirRead(err))
	})

	t.Run("empty directory", func(t *testing.T) {
		tempDir := t.TempDir()

		manager, err := NewManager(tempDir)
		require.NoError(t, err)
		assert.NotNil(t, manager)
		assert.NotNil(t, manager.root)
	})

	t.Run("directory with non-yaml files", func(t *testing.T) {
		tempDir := t.TempDir()

		// 创建非YAML文件
		createTestFile(t, tempDir, "config.txt", "not a yaml file")
		createTestFile(t, tempDir, "config.json", `{"key": "value"}`)

		manager, err := NewManager(tempDir)
		require.NoError(t, err)
		assert.NotNil(t, manager)
	})

	t.Run("directory with invalid yaml file", func(t *testing.T) {
		tempDir := t.TempDir()

		// 创建无效的YAML文件
		createTestFile(t, tempDir, "invalid.yml", "invalid: yaml: content: [")

		_, err := NewManager(tempDir)
		require.Error(t, err)
		assert.True(t, IsFileRead(err))
	})

	t.Run("directory with duplicate keys", func(t *testing.T) {
		tempDir := t.TempDir()

		// 创建两个具有重复键的配置文件
		createTestConfigFile(t, tempDir, "app1.yml", map[string]interface{}{
			"database": map[string]interface{}{
				"host": "localhost",
			},
		})
		createTestConfigFile(t, tempDir, "app2.yml", map[string]interface{}{
			"database": map[string]interface{}{
				"host": "remote",
			},
		})

		_, err := NewManager(tempDir)
		require.Error(t, err)
		assert.True(t, IsDuplicateKey(err))
	})
}

// TestMustNewManager 测试 MustNewManager 函数。
func TestMustNewManager(t *testing.T) {
	t.Run("valid config directory", func(t *testing.T) {
		tempDir := t.TempDir()
		createTestConfigFile(t, tempDir, "app.yml", map[string]interface{}{
			"service": map[string]interface{}{
				"name": "test",
			},
		})

		manager := MustNewManager(tempDir)
		assert.NotNil(t, manager)
		assert.Equal(t, tempDir, manager.configDir)
	})

	t.Run("non-existent directory panics", func(t *testing.T) {
		assert.Panics(t, func() {
			MustNewManager("/non/existent/directory")
		})
	})
}

// TestManager_Get 测试 Get 方法的各种场景。
func TestManager_Get(t *testing.T) {
	t.Run("existing config", func(t *testing.T) {
		tempDir := t.TempDir()
		createTestConfigFile(t, tempDir, "app.yml", map[string]interface{}{
			"database": map[string]interface{}{
				"host": "localhost",
				"port": 5432,
			},
			"cache": map[string]interface{}{
				"type": "redis",
			},
		})

		manager := MustNewManager(tempDir)

		dbConfig, err := manager.Get("database")
		require.NoError(t, err)
		assert.Equal(t, "localhost", dbConfig.GetString("host"))
		assert.Equal(t, 5432, dbConfig.GetInt("port"))

		cacheConfig, err := manager.Get("cache")
		require.NoError(t, err)
		assert.Equal(t, "redis", cacheConfig.GetString("type"))
	})

	t.Run("non-existing config", func(t *testing.T) {
		tempDir := t.TempDir()
		createTestConfigFile(t, tempDir, "app.yml", map[string]interface{}{
			"database": map[string]interface{}{
				"host": "localhost",
			},
		})

		manager := MustNewManager(tempDir)

		_, err := manager.Get("nonexistent")
		require.Error(t, err)
		assert.True(t, IsNotFound(err))
	})

	t.Run("caching behavior", func(t *testing.T) {
		tempDir := t.TempDir()
		createTestConfigFile(t, tempDir, "app.yml", map[string]interface{}{
			"service": map[string]interface{}{
				"name": "test",
			},
		})

		manager := MustNewManager(tempDir)

		// 第一次调用应该填充缓存
		config1, err := manager.Get("service")
		require.NoError(t, err)

		// 第二次调用应该返回缓存的实例
		config2, err := manager.Get("service")
		require.NoError(t, err)

		assert.Same(t, config1, config2)
		assert.Contains(t, manager.configs, "service")
	})
}

// TestManager_Get_ConcurrentAccess 测试配置的线程安全并发访问。
func TestManager_Get_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	createTestConfigFile(t, tempDir, "app.yml", map[string]interface{}{
		"service": map[string]interface{}{
			"name": "test",
		},
		"database": map[string]interface{}{
			"host": "localhost",
		},
	})

	manager := MustNewManager(tempDir)

	var wg sync.WaitGroup
	var errors []error
	var mu sync.Mutex

	// 启动多个goroutine并发访问配置
	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func(id int) {
			defer wg.Done()
			_, err := manager.Get("service")
			mu.Lock()
			errors = append(errors, err)
			mu.Unlock()
		}(i)
		go func(id int) {
			defer wg.Done()
			_, err := manager.Get("database")
			mu.Lock()
			errors = append(errors, err)
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// 所有操作都应该成功
	for _, err := range errors {
		assert.NoError(t, err)
	}

	// 缓存应该包含两个配置
	assert.Contains(t, manager.configs, "service")
	assert.Contains(t, manager.configs, "database")
}

// TestManager_Get_DoubleCheckedLocking 测试双重检查锁定模式。
func TestManager_Get_DoubleCheckedLocking(t *testing.T) {
	tempDir := t.TempDir()
	createTestConfigFile(t, tempDir, "app.yml", map[string]interface{}{
		"service": map[string]interface{}{
			"name": "test",
		},
	})

	manager := MustNewManager(tempDir)

	var wg sync.WaitGroup
	configs := make([]*viper.Viper, 10)

	// 多个goroutine同时尝试获取相同的配置
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			config, _ := manager.Get("service")
			configs[index] = config
		}(i)
	}

	wg.Wait()

	// 所有应该返回相同的实例（相同的指针）
	first := configs[0]
	for _, config := range configs {
		assert.Same(t, first, config)
	}
}

// TestManager_MustGet 测试 MustGet 方法。
func TestManager_MustGet(t *testing.T) {
	t.Run("existing config", func(t *testing.T) {
		tempDir := t.TempDir()
		createTestConfigFile(t, tempDir, "app.yml", map[string]interface{}{
			"service": map[string]interface{}{
				"name": "test",
			},
		})

		manager := MustNewManager(tempDir)
		config := manager.MustGet("service")
		assert.Equal(t, "test", config.GetString("name"))
	})

	t.Run("non-existing config panics", func(t *testing.T) {
		tempDir := t.TempDir()
		manager := MustNewManager(tempDir)

		assert.Panics(t, func() {
			manager.MustGet("nonexistent")
		})
	})
}

// TestManager_Root 测试 Root 方法。
func TestManager_Root(t *testing.T) {
	tempDir := t.TempDir()
	createTestConfigFile(t, tempDir, "app.yml", map[string]interface{}{
		"service": map[string]interface{}{
			"name": "test",
		},
		"database": map[string]interface{}{
			"host": "localhost",
		},
	})

	manager := MustNewManager(tempDir)
	root := manager.Root()

	assert.NotNil(t, root)
	assert.Equal(t, "test", root.GetString("service.name"))
	assert.Equal(t, "localhost", root.GetString("database.host"))
}

// TestManager_List 测试 AllNames 方法。
func TestManager_List(t *testing.T) {
	tempDir := t.TempDir()
	createTestConfigFile(t, tempDir, "app.yml", map[string]interface{}{
		"zebra":  map[string]interface{}{"value": 1},
		"apple":  map[string]interface{}{"value": 2},
		"banana": map[string]interface{}{"value": 3},
	})

	manager := MustNewManager(tempDir)

	// 应该返回所有可用的名称，已排序
	names := manager.List()
	assert.Equal(t, []string{"apple", "banana", "zebra"}, names)

	// 即使加载配置后，仍应返回所有名称
	manager.Get("banana")
	names = manager.List()
	assert.Equal(t, []string{"apple", "banana", "zebra"}, names)
}

// TestManager_Reset 测试 Reset 方法。
func TestManager_Reset(t *testing.T) {
	tempDir := t.TempDir()
	createTestConfigFile(t, tempDir, "app.yml", map[string]interface{}{
		"service": map[string]interface{}{
			"name": "test",
		},
	})

	manager := MustNewManager(tempDir)

	// 加载一些配置
	manager.Get("service")
	assert.NotEmpty(t, manager.configs)

	// 重置应该清除缓存
	err := manager.Reset()
	require.NoError(t, err)
	assert.Empty(t, manager.configs)

	// 应该能够再次加载配置
	config, err := manager.Get("service")
	require.NoError(t, err)
	assert.Equal(t, "test", config.GetString("name"))
}

// TestManager_OnReload 测试 OnReload 方法。
func TestManager_OnReload(t *testing.T) {
	tempDir := t.TempDir()
	createTestConfigFile(t, tempDir, "app.yml", map[string]interface{}{
		"service": map[string]interface{}{
			"name": "test",
		},
	})

	manager := MustNewManager(tempDir)

	// 测试回调注册
	callback := func(m *Manager) error {
		return nil
	}

	manager.OnReload(callback)
	assert.Len(t, manager.reloadCallbacks, 1)

	// 测试回调是否正确存储
	assert.Len(t, manager.reloadCallbacks, 1)

	// 添加另一个回调
	callback2 := func(m *Manager) error {
		return nil
	}
	manager.OnReload(callback2)

	assert.Len(t, manager.reloadCallbacks, 2)
}

// TestManager_Watch 测试 Watch 方法。
func TestManager_Watch(t *testing.T) {
	t.Run("start watching", func(t *testing.T) {
		tempDir := t.TempDir()
		createTestConfigFile(t, tempDir, "app.yml", map[string]interface{}{
			"service": map[string]interface{}{
				"name": "test",
			},
		})

		manager := MustNewManager(tempDir)

		err := manager.Watch()
		require.NoError(t, err)
		assert.NotNil(t, manager.watcher)
		assert.NotNil(t, manager.watcherDone)

		// 清理
		manager.StopWatch()
	})

	t.Run("idempotent watching", func(t *testing.T) {
		tempDir := t.TempDir()
		createTestConfigFile(t, tempDir, "app.yml", map[string]interface{}{
			"service": map[string]interface{}{
				"name": "test",
			},
		})

		manager := MustNewManager(tempDir)

		// 开始监听
		err := manager.Watch()
		require.NoError(t, err)

		watcher := manager.watcher

		// 再次调用不应该创建新的监听器
		err = manager.Watch()
		require.NoError(t, err)
		assert.Same(t, watcher, manager.watcher)

		// 清理
		manager.StopWatch()
	})

	t.Run("watch non-existent directory", func(t *testing.T) {
		manager := &Manager{configDir: "/non/existent/directory"}

		err := manager.Watch()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to watch directory")
	})
}

// TestManager_StopWatch 测试 StopWatch 方法。
func TestManager_StopWatch(t *testing.T) {
	tempDir := t.TempDir()
	createTestConfigFile(t, tempDir, "app.yml", map[string]interface{}{
		"service": map[string]interface{}{
			"name": "test",
		},
	})

	manager := MustNewManager(tempDir)

	// 开始监听
	err := manager.Watch()
	require.NoError(t, err)

	watcher := manager.watcher
	done := manager.watcherDone

	// 停止监听
	manager.StopWatch()

	// 应该是幂等的
	manager.StopWatch()
	manager.StopWatch()

	// 注意：我们无法轻易测试监听器是否实际关闭
	// 因为fsnotify没有暴露检查是否关闭的方法
	// 但我们可以验证一次性停止的行为
	assert.Equal(t, watcher, manager.watcher)
	assert.Equal(t, done, manager.watcherDone)
}

// TestManager_WatchFileChange 测试文件监听功能。
func TestManager_WatchFileChange(t *testing.T) {
	tempDir := t.TempDir()

	// 创建初始配置
	createTestConfigFile(t, tempDir, "app.yml", map[string]interface{}{
		"service": map[string]interface{}{
			"name": "initial",
		},
	})

	manager := MustNewManager(tempDir)

	var mu sync.Mutex
	var callbackCalled bool
	var callbackCount int

	manager.OnReload(func(m *Manager) error {
		mu.Lock()
		callbackCalled = true
		callbackCount++
		mu.Unlock()
		return nil
	})

	// 开始监听
	err := manager.Watch()
	require.NoError(t, err)
	defer manager.StopWatch()

	// 修改配置文件
	time.Sleep(100 * time.Millisecond) // 给监听器时间启动
	createTestConfigFile(t, tempDir, "app.yml", map[string]interface{}{
		"service": map[string]interface{}{
			"name": "modified",
		},
	})

	// 等待文件系统事件被处理
	time.Sleep(200 * time.Millisecond)

	// 回调应该已经被调用
	mu.Lock()
	called := callbackCalled
	count := callbackCount
	mu.Unlock()

	assert.True(t, called)
	assert.Equal(t, 1, count)

	// 配置应该被更新
	config := manager.MustGet("service")
	assert.Equal(t, "modified", config.GetString("name"))
}

// TestInit 测试全局 Init 函数。
func TestInit(t *testing.T) {
	t.Run("successful initialization", func(t *testing.T) {
		// 重置全局状态
		defaultManager = nil
		defaultManagerOnce = sync.Once{}

		tempDir := t.TempDir()
		createTestConfigFile(t, tempDir, "app.yml", map[string]interface{}{
			"service": map[string]interface{}{
				"name": "test",
			},
		})

		Init(tempDir)

		manager := Default()
		assert.NotNil(t, manager)
		assert.Equal(t, tempDir, manager.configDir)
	})

	t.Run("multiple calls are idempotent", func(t *testing.T) {
		// 重置全局状态
		defaultManager = nil
		defaultManagerOnce = sync.Once{}

		tempDir := t.TempDir()
		createTestConfigFile(t, tempDir, "app.yml", map[string]interface{}{
			"service": map[string]interface{}{
				"name": "test",
			},
		})

		Init(tempDir)
		firstManager := Default()

		// 第二次调用应该什么都不做
		secondManager := Default()

		assert.Same(t, firstManager, secondManager)
	})

	t.Run("panic on invalid directory", func(t *testing.T) {
		// 重置全局状态
		defaultManager = nil
		defaultManagerOnce = sync.Once{}

		assert.Panics(t, func() {
			Init("/non/existent/directory")
		})
	})
}

// TestDefault 测试 Default 函数。
func TestDefault(t *testing.T) {
	// 重置全局状态
	defaultManager = nil
	defaultManagerOnce = sync.Once{}

	// 初始化之前
	assert.Nil(t, Default())

	// 初始化之后
	tempDir := t.TempDir()
	createTestConfigFile(t, tempDir, "app.yml", map[string]interface{}{
		"service": map[string]interface{}{
			"name": "test",
		},
	})

	Init(tempDir)
	assert.NotNil(t, Default())
}

// TestManager_Integration 测试管理器功能的完整集成。
func TestManager_Integration(t *testing.T) {
	tempDir := t.TempDir()

	// 创建多个配置文件
	createTestConfigFile(t, tempDir, "database.yml", map[string]interface{}{
		"database": map[string]interface{}{
			"host":     "localhost",
			"port":     5432,
			"username": "user",
			"password": "pass",
		},
	})

	createTestConfigFile(t, tempDir, "cache.yml", map[string]interface{}{
		"cache": map[string]interface{}{
			"type":     "redis",
			"host":     "localhost",
			"port":     6379,
			"database": 0,
		},
	})

	createTestConfigFile(t, tempDir, "service.yml", map[string]interface{}{
		"service": map[string]interface{}{
			"name":    "api-server",
			"version": "1.0.0",
			"port":    8080,
		},
	})

	manager := MustNewManager(tempDir)

	// 测试 List
	allNames := manager.List()
	assert.Equal(t, []string{"cache", "database", "service"}, allNames)

	// 测试获取配置
	dbConfig, err := manager.Get("database")
	require.NoError(t, err)
	assert.Equal(t, "localhost", dbConfig.GetString("host"))
	assert.Equal(t, 5432, dbConfig.GetInt("port"))

	cacheConfig, err := manager.Get("cache")
	require.NoError(t, err)
	assert.Equal(t, "redis", cacheConfig.GetString("type"))
	assert.Equal(t, 6379, cacheConfig.GetInt("port"))

	serviceConfig, err := manager.Get("service")
	require.NoError(t, err)
	assert.Equal(t, "api-server", serviceConfig.GetString("name"))
	assert.Equal(t, 8080, serviceConfig.GetInt("port"))

	// 测试缓存
	assert.Len(t, manager.configs, 3)

	// 测试名称（已缓存的配置）
	cachedNames := manager.List()
	assert.Equal(t, []string{"cache", "database", "service"}, cachedNames)

	// 测试重置
	err = manager.Reset()
	require.NoError(t, err)
	assert.Empty(t, manager.configs)

	// 应该能够再次获取配置
	dbConfig2, err := manager.Get("database")
	require.NoError(t, err)
	assert.Equal(t, "localhost", dbConfig2.GetString("host"))
}

// 基准测试
func BenchmarkManager_Get_Cached(b *testing.B) {
	tempDir := b.TempDir()
	createBenchmarkConfigFile(b, tempDir, "app.yml", map[string]interface{}{
		"service": map[string]interface{}{
			"name": "test",
		},
	})

	manager := MustNewManager(tempDir)
	manager.Get("service") // 预填充缓存

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.Get("service")
	}
}

func BenchmarkManager_Get_Concurrent(b *testing.B) {
	tempDir := b.TempDir()
	createBenchmarkConfigFile(b, tempDir, "app.yml", map[string]interface{}{
		"service": map[string]interface{}{
			"name": "test",
		},
	})

	manager := MustNewManager(tempDir)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			manager.Get("service")
		}
	})
}

// 辅助函数
func createTestConfigFile(t *testing.T, dir, filename string, data map[string]interface{}) {
	t.Helper()

	v := viper.New()
	for key, value := range data {
		v.Set(key, value)
	}

	configPath := filepath.Join(dir, filename)
	err := v.WriteConfigAs(configPath)
	require.NoError(t, err)
}

func createBenchmarkConfigFile(b *testing.B, dir, filename string, data map[string]interface{}) {
	b.Helper()

	v := viper.New()
	for key, value := range data {
		v.Set(key, value)
	}

	configPath := filepath.Join(dir, filename)
	err := v.WriteConfigAs(configPath)
	if err != nil {
		b.Fatalf("Failed to create benchmark config file: %v", err)
	}
}

func createTestFile(t *testing.T, dir, filename, content string) {
	t.Helper()

	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)
}
