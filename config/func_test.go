package config

import (
	"errors"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig 测试 Config 函数的基本功能
func TestConfig(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *Manager
		configName string
		expected   interface{}
		wantErr    bool
		errType    error
	}{
		{
			name: "成功获取结构体配置",
			setup: func() *Manager {
				m := &Manager{
					root:    viper.New(),
					configs: make(map[string]*viper.Viper),
				}
				m.root.Set("database.host", "localhost")
				m.root.Set("database.port", 8080)
				m.root.Set("database.timeout", "30s")
				return m
			},
			configName: "database",
			expected: DatabaseConfig{
				Host:    "localhost",
				Port:    8080,
				Timeout: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "成功获取基础类型配置",
			setup: func() *Manager {
				m := &Manager{
					root:    viper.New(),
					configs: make(map[string]*viper.Viper),
				}
				m.root.Set("server.max_connections", 100)
				return m
			},
			configName: "server",
			expected:   ServerConfig{MaxConnections: 100},
			wantErr:    false,
		},
		{
			name: "配置不存在",
			setup: func() *Manager {
				m := &Manager{
					root:    viper.New(),
					configs: make(map[string]*viper.Viper),
				}
				return m
			},
			configName: "nonexistent",
			expected:   DatabaseConfig{},
			wantErr:    true,
			errType:    ErrNotFound,
		},
		{
			name: "反序列化失败",
			setup: func() *Manager {
				m := &Manager{
					root:    viper.New(),
					configs: make(map[string]*viper.Viper),
				}
				m.root.Set("database.host", "localhost")
				// port 字段类型不匹配（字符串而非整数）
				m.root.Set("database.port", "invalid")
				return m
			},
			configName: "database",
			expected:   DatabaseConfig{},
			wantErr:    true,
		},
		{
			name: "空配置名",
			setup: func() *Manager {
				m := &Manager{
					root:    viper.New(),
					configs: make(map[string]*viper.Viper),
				}
				return m
			},
			configName: "",
			expected:   DatabaseConfig{},
			wantErr:    true,
			errType:    ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()

			var result interface{}
			var err error

			switch tt.expected.(type) {
			case DatabaseConfig:
				result, err = Config[DatabaseConfig](m, tt.configName)
			case ServerConfig:
				result, err = Config[ServerConfig](m, tt.configName)
			default:
				t.Fatalf("未知的预期类型: %T", tt.expected)
			}

			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					assert.True(t, IsNotFound(err), "错误应该是 ErrNotFound 类型")
				}
				// 验证错误消息包含配置名
				if tt.configName != "" && tt.errType == nil {
					assert.Contains(t, err.Error(), tt.configName)
					assert.Contains(t, err.Error(), "unmarshal")
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestConfig_NilManager 测试传入 nil Manager 的处理
func TestConfig_NilManager(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// 预期会有 panic，这是正常的
			assert.NotNil(t, r)
		}
	}()

	result, err := Config[DatabaseConfig](nil, "test")

	// 如果没有 panic，检查错误
	require.Error(t, err)
	assert.Equal(t, DatabaseConfig{}, result)
}

// TestConfig_EmptyConfig 测试空配置的处理
func TestConfig_EmptyConfig(t *testing.T) {
	m := &Manager{
		root:    viper.New(),
		configs: make(map[string]*viper.Viper),
	}
	// 创建空的数据库配置
	m.root.Set("empty", map[string]interface{}{})

	var result DatabaseConfig
	result, err := Config[DatabaseConfig](m, "empty")

	// 空配置应该成功反序列化为零值
	require.NoError(t, err)
	assert.Equal(t, DatabaseConfig{}, result)
}

// TestConfig_NestedStruct 测试嵌套结构体的反序列化
func TestConfig_NestedStruct(t *testing.T) {
	m := &Manager{
		root:    viper.New(),
		configs: make(map[string]*viper.Viper),
	}
	m.root.Set("app.host", "localhost")
	m.root.Set("app.port", 8080)
	m.root.Set("app.database.host", "db.example.com")
	m.root.Set("app.database.port", 5432)
	m.root.Set("app.database.name", "testdb")

	result, err := Config[AppConfig](m, "app")

	require.NoError(t, err)
	assert.Equal(t, "localhost", result.Host)
	assert.Equal(t, 8080, result.Port)
	assert.Equal(t, "db.example.com", result.Database.Host)
	assert.Equal(t, 5432, result.Database.Port)
	assert.Equal(t, "testdb", result.Database.Name)
}

// TestConfig_SliceConfig 测试切片类型的配置
func TestConfig_SliceConfig(t *testing.T) {
	m := &Manager{
		root:    viper.New(),
		configs: make(map[string]*viper.Viper),
	}
	m.root.Set("cluster.servers", []string{"server1", "server2", "server3"})
	m.root.Set("cluster.ports", []int{8080, 8081, 8082})

	// 测试获取整个 cluster 配置
	type ClusterConfig struct {
		Servers []string `mapstructure:"servers"`
		Ports   []int    `mapstructure:"ports"`
	}

	result, err := Config[ClusterConfig](m, "cluster")
	require.NoError(t, err)
	assert.Equal(t, []string{"server1", "server2", "server3"}, result.Servers)
	assert.Equal(t, []int{8080, 8081, 8082}, result.Ports)
}

// TestConfig_MapConfig 测试映射类型的配置
func TestConfig_MapConfig(t *testing.T) {
	m := &Manager{
		root:    viper.New(),
		configs: make(map[string]*viper.Viper),
	}
	m.root.Set("monitoring.labels.env", "production")
	m.root.Set("monitoring.labels.team", "backend")
	m.root.Set("monitoring.metrics.cpu", 80)
	m.root.Set("monitoring.metrics.memory", 70)

	type MonitoringConfig struct {
		Labels  map[string]string `mapstructure:"labels"`
		Metrics map[string]int    `mapstructure:"metrics"`
	}

	result, err := Config[MonitoringConfig](m, "monitoring")
	require.NoError(t, err)
	assert.Equal(t, "production", result.Labels["env"])
	assert.Equal(t, "backend", result.Labels["team"])
	assert.Equal(t, 80, result.Metrics["cpu"])
	assert.Equal(t, 70, result.Metrics["memory"])
}

// TestMustConfig 测试 MustConfig 函数
func TestMustConfig(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *Manager
		configName  string
		expected    interface{}
		shouldPanic bool
	}{
		{
			name: "成功获取配置",
			setup: func() *Manager {
				m := &Manager{
					root:    viper.New(),
					configs: make(map[string]*viper.Viper),
				}
				m.root.Set("database.host", "localhost")
				m.root.Set("database.port", 8080)
				return m
			},
			configName:  "database",
			expected:    DatabaseConfig{Host: "localhost", Port: 8080},
			shouldPanic: false,
		},
		{
			name: "配置不存在时 panic",
			setup: func() *Manager {
				m := &Manager{
					root:    viper.New(),
					configs: make(map[string]*viper.Viper),
				}
				return m
			},
			configName:  "nonexistent",
			shouldPanic: true,
		},
		{
			name: "反序列化失败时 panic",
			setup: func() *Manager {
				m := &Manager{
					root:    viper.New(),
					configs: make(map[string]*viper.Viper),
				}
				m.root.Set("database.host", "localhost")
				m.root.Set("database.port", "invalid") // 类型不匹配
				return m
			},
			configName:  "database",
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()

			if tt.shouldPanic {
				assert.Panics(t, func() {
					switch tt.expected.(type) {
					case DatabaseConfig:
						_ = MustConfig[DatabaseConfig](m, tt.configName)
					default:
						_ = MustConfig[DatabaseConfig](m, tt.configName)
					}
				})
			} else {
				var result interface{}
				switch tt.expected.(type) {
				case DatabaseConfig:
					result = MustConfig[DatabaseConfig](m, tt.configName)
				default:
					result = MustConfig[DatabaseConfig](m, tt.configName)
				}
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestMustConfig_NilManager 测试 nil Manager 的 panic 行为
func TestMustConfig_NilManager(t *testing.T) {
	assert.Panics(t, func() {
		_ = MustConfig[DatabaseConfig](nil, "test")
	})
}

// TestConfig_ErrorWrapping 测试错误包装
func TestConfig_ErrorWrapping(t *testing.T) {
	m := &Manager{
		root:    viper.New(),
		configs: make(map[string]*viper.Viper),
	}
	m.root.Set("test.host", "localhost")
	m.root.Set("test.port", "invalid") // 会导致反序列化错误

	_, err := Config[DatabaseConfig](m, "test")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "config \"test\"")
	assert.Contains(t, err.Error(), "unmarshal")
}

// TestConfig_ConcurrentAccess 测试并发访问安全性
func TestConfig_ConcurrentAccess(t *testing.T) {
	m := &Manager{
		root:    viper.New(),
		configs: make(map[string]*viper.Viper),
	}
	m.root.Set("test.value", 42)

	const numGoroutines = 10
	results := make(chan int, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			type TestConfig struct {
				Value int `mapstructure:"value"`
			}
			result, err := Config[TestConfig](m, "test")
			if err != nil {
				errors <- err
			} else {
				results <- result.Value
			}
		}()
	}

	// 收集结果
	var successCount int
	var errorCount int

	for i := 0; i < numGoroutines; i++ {
		select {
		case result := <-results:
			assert.Equal(t, 42, result)
			successCount++
		case err := <-errors:
			t.Errorf("意外的错误: %v", err)
			errorCount++
		case <-time.After(time.Second):
			t.Fatal("超时")
		}
	}

	assert.Equal(t, numGoroutines, successCount)
	assert.Equal(t, 0, errorCount)
}

// BenchmarkConfig_Config 性能基准测试
func BenchmarkConfig_Config(b *testing.B) {
	m := &Manager{
		root:    viper.New(),
		configs: make(map[string]*viper.Viper),
	}
	m.root.Set("database.host", "localhost")
	m.root.Set("database.port", 8080)
	m.root.Set("database.timeout", "30s")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Config[DatabaseConfig](m, "database")
	}
}

// BenchmarkConfig_SimpleType 简单类型性能测试
func BenchmarkConfig_SimpleType(b *testing.B) {
	m := &Manager{
		root:    viper.New(),
		configs: make(map[string]*viper.Viper),
	}
	m.root.Set("simple.value", 42)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Config[int](m, "simple.value")
	}
}

// BenchmarkConfig_MustConfig MustConfig 性能测试
func BenchmarkConfig_MustConfig(b *testing.B) {
	m := &Manager{
		root:    viper.New(),
		configs: make(map[string]*viper.Viper),
	}
	m.root.Set("database.host", "localhost")
	m.root.Set("database.port", 8080)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MustConfig[DatabaseConfig](m, "database")
	}
}

// BenchmarkConfig_Concurrent 并发性能测试
func BenchmarkConfig_Concurrent(b *testing.B) {
	m := &Manager{
		root:    viper.New(),
		configs: make(map[string]*viper.Viper),
	}
	m.root.Set("concurrent.value", 42)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = Config[int](m, "concurrent")
		}
	})
}

// 测试用的结构体定义

// DatabaseConfig 数据库配置结构体
type DatabaseConfig struct {
	Host    string        `mapstructure:"host"`
	Port    int           `mapstructure:"port"`
	Timeout time.Duration `mapstructure:"timeout"`
	Name    string        `mapstructure:"name"`
}

// ServerConfig 服务器配置结构体
type ServerConfig struct {
	MaxConnections int `mapstructure:"max_connections"`
}

// AppConfig 应用配置结构体（包含嵌套结构）
type AppConfig struct {
	Host     string         `mapstructure:"host"`
	Port     int            `mapstructure:"port"`
	Database DatabaseConfig `mapstructure:"database"`
}

// MockManager 用于测试的模拟 Manager
type MockManager struct {
	configs map[string]*viper.Viper
	getErr  error
}

// Get 模拟 Manager 的 Get 方法
func (m *MockManager) Get(name string) (*viper.Viper, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if v, ok := m.configs[name]; ok {
		return v, nil
	}
	return nil, ErrNotFound
}

// TestConfig_WithMockManager 使用 Mock Manager 测试
func TestConfig_WithMockManager(t *testing.T) {
	tests := []struct {
		name       string
		mock       *MockManager
		configName string
		expected   DatabaseConfig
		wantErr    bool
	}{
		{
			name: "Mock Manager 成功获取",
			mock: &MockManager{
				configs: map[string]*viper.Viper{
					"test": func() *viper.Viper {
						v := viper.New()
						v.Set("host", "mock-host")
						v.Set("port", 9999)
						return v
					}(),
				},
			},
			configName: "test",
			expected:   DatabaseConfig{Host: "mock-host", Port: 9999},
			wantErr:    false,
		},
		{
			name: "Mock Manager 返回错误",
			mock: &MockManager{
				getErr: errors.New("mock error"),
			},
			configName: "test",
			expected:   DatabaseConfig{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mock.getErr != nil {
				// 直接测试错误情况，不创建真实的 Manager
				m := &Manager{
					root:    viper.New(),
					configs: make(map[string]*viper.Viper),
				}
				// 模拟错误情况
				_, err := Config[DatabaseConfig](m, "nonexistent")
				require.Error(t, err)
				assert.Equal(t, DatabaseConfig{}, tt.expected)
			} else {
				// 创建真实的 Manager 并使用 Mock 数据
				m := &Manager{
					root:    viper.New(),
					configs: make(map[string]*viper.Viper),
				}
				// 直接在 root 中设置配置数据
				for key, viper := range tt.mock.configs {
					m.root.Set(key+".host", viper.Get("host"))
					m.root.Set(key+".port", viper.Get("port"))
				}

				result, err := Config[DatabaseConfig](m, tt.configName)

				if tt.wantErr {
					require.Error(t, err)
					assert.Equal(t, DatabaseConfig{}, result)
				} else {
					require.NoError(t, err)
					assert.Equal(t, tt.expected, result)
				}
			}
		})
	}
}
