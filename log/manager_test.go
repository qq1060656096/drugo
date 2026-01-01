package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TestNewManager 测试创建新的日志管理器
func TestNewManager(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
		errType error
	}{
		{
			name: "有效配置",
			cfg: Config{
				Dir:        "/tmp/logs",
				Level:      "info",
				Format:     "json",
				MaxSize:    100,
				MaxBackups: 10,
				MaxAge:     30,
				Compress:   true,
				Console:    false,
			},
			wantErr: false,
		},
		{
			name: "空目录",
			cfg: Config{
				Dir: "",
			},
			wantErr: true,
			errType: ErrEmptyLogDir,
		},
		{
			name: "无效的MaxSize",
			cfg: Config{
				Dir:     "/tmp/logs",
				MaxSize: -1,
			},
			wantErr: true,
			errType: ErrInvalidConfigValue,
		},
		{
			name: "无效的MaxBackups",
			cfg: Config{
				Dir:        "/tmp/logs",
				MaxBackups: -1,
			},
			wantErr: true,
			errType: ErrInvalidConfigValue,
		},
		{
			name: "无效的MaxAge",
			cfg: Config{
				Dir:    "/tmp/logs",
				MaxAge: -1,
			},
			wantErr: true,
			errType: ErrInvalidConfigValue,
		},
		{
			name: "无效的日志级别",
			cfg: Config{
				Dir:   "/tmp/logs",
				Level: "invalid",
			},
			wantErr: true,
			errType: ErrInvalidLogLevel,
		},
		{
			name: "无效的日志格式",
			cfg: Config{
				Dir:    "/tmp/logs",
				Format: "invalid",
			},
			wantErr: true,
			errType: ErrInvalidLogFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewManager(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, IsEmptyLogDir(err) || IsInvalidConfigValue(err) || IsInvalidLogLevel(err) || IsInvalidLogFormat(err))
				assert.Nil(t, m)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, m)
				assert.Equal(t, tt.cfg, m.cfg)
				assert.NotNil(t, m.loggers)
				assert.NotNil(t, m.levels)
			}
		})
	}
}

// TestMustNewManager 测试MustNewManager函数
func TestMustNewManager(t *testing.T) {
	cfg := Config{
		Dir:   "/tmp/logs",
		Level: "info",
	}

	// 测试正常情况
	assert.NotPanics(t, func() {
		m := MustNewManager(cfg)
		assert.NotNil(t, m)
	})

	// 测试panic情况
	assert.Panics(t, func() {
		MustNewManager(Config{Dir: ""})
	})
}

// TestManager_Get 测试获取日志实例
func TestManager_Get(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()
	cfg := Config{
		Dir:    tempDir,
		Level:  "info",
		Format: "json",
	}

	m, err := NewManager(cfg)
	require.NoError(t, err)

	tests := []struct {
		name    string
		bizName string
		wantErr bool
		errType error
	}{
		{
			name:    "有效的业务名称",
			bizName: "test",
			wantErr: false,
		},
		{
			name:    "空业务名称",
			bizName: "",
			wantErr: true,
			errType: ErrEmptyBizName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := m.Get(tt.bizName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, IsEmptyBizName(err))
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
			}
		})
	}

	// 测试缓存功能 - 同一个业务名称应该返回相同的logger实例
	logger1, err := m.Get("cache_test")
	require.NoError(t, err)
	logger2, err := m.Get("cache_test")
	require.NoError(t, err)
	assert.Same(t, logger1, logger2, "应该返回缓存的相同logger实例")
}

// TestManager_Get_ConcurrentAccess 测试并发访问Get方法
func TestManager_Get_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{
		Dir:    tempDir,
		Level:  "info",
		Format: "json",
	}

	m, err := NewManager(cfg)
	require.NoError(t, err)

	const numGoroutines = 100
	const numRequests = 10

	var wg sync.WaitGroup
	var mu sync.Mutex
	loggers := make(map[string]*zap.Logger)

	// 并发获取logger
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numRequests; j++ {
				bizName := fmt.Sprintf("service_%d", id%10) // 10个不同的业务名称
				logger, err := m.Get(bizName)
				assert.NoError(t, err)
				assert.NotNil(t, logger)

				mu.Lock()
				loggers[bizName] = logger
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// 验证每个业务名称只有一个logger实例
	assert.Len(t, loggers, 10, "应该有10个不同的logger实例")
	for bizName, logger := range loggers {
		// 再次获取应该返回相同的实例
		sameLogger, err := m.Get(bizName)
		assert.NoError(t, err)
		assert.Same(t, logger, sameLogger, "业务名称 %s 应该返回相同的logger实例", bizName)
	}
}

// TestManager_Get_DoubleCheckedLocking 测试双重检查锁定模式
func TestManager_Get_DoubleCheckedLocking(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{
		Dir:    tempDir,
		Level:  "info",
		Format: "json",
	}

	m, err := NewManager(cfg)
	require.NoError(t, err)

	bizName := "double_check_test"

	// 第一次获取，创建新的logger
	logger1, err := m.Get(bizName)
	require.NoError(t, err)

	// 使用race detector测试双重检查锁定
	done := make(chan bool, 2)

	// 两个goroutine同时尝试获取同一个logger
	go func() {
		logger, err := m.Get(bizName)
		assert.NoError(t, err)
		assert.Same(t, logger1, logger)
		done <- true
	}()

	go func() {
		logger, err := m.Get(bizName)
		assert.NoError(t, err)
		assert.Same(t, logger1, logger)
		done <- true
	}()

	// 等待两个goroutine完成
	<-done
	<-done
}

// TestManager_MustGet 测试MustGet方法
func TestManager_MustGet(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{
		Dir:    tempDir,
		Level:  "info",
		Format: "json",
	}

	m, err := NewManager(cfg)
	require.NoError(t, err)

	// 测试正常情况
	assert.NotPanics(t, func() {
		logger := m.MustGet("test")
		assert.NotNil(t, logger)
	})

	// 测试panic情况
	assert.Panics(t, func() {
		m.MustGet("")
	})
}

// TestManager_Sync 测试同步方法
func TestManager_Sync(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{
		Dir:    tempDir,
		Level:  "info",
		Format: "json",
	}

	m, err := NewManager(cfg)
	require.NoError(t, err)

	// 创建几个logger
	_, err = m.Get("test1")
	require.NoError(t, err)
	_, err = m.Get("test2")
	require.NoError(t, err)

	// 测试同步
	err = m.Sync()
	assert.NoError(t, err)

	// 测试空manager的同步
	emptyManager, _ := NewManager(cfg)
	err = emptyManager.Sync()
	assert.NoError(t, err)
}

// TestManager_Close 测试关闭方法
func TestManager_Close(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{
		Dir:    tempDir,
		Level:  "info",
		Format: "json",
	}

	m, err := NewManager(cfg)
	require.NoError(t, err)

	// 创建几个logger
	_, err = m.Get("test1")
	require.NoError(t, err)
	_, err = m.Get("test2")
	require.NoError(t, err)

	// 验证logger存在
	assert.Len(t, m.List(), 2)

	// 关闭manager
	err = m.Close()
	assert.NoError(t, err)

	// 验证logger缓存被清空
	assert.Len(t, m.List(), 0)
	assert.Len(t, m.loggers, 0)
	assert.Len(t, m.levels, 0)

	// 关闭后应该能重新创建logger
	_, err = m.Get("test3")
	assert.NoError(t, err)
	assert.Len(t, m.List(), 1)
}

// TestManager_List 测试列表方法
func TestManager_List(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{
		Dir:    tempDir,
		Level:  "info",
		Format: "json",
	}

	m, err := NewManager(cfg)
	require.NoError(t, err)

	// 空列表
	assert.Empty(t, m.List())

	// 添加logger
	bizNames := []string{"service1", "service2", "service3"}
	for _, name := range bizNames {
		_, err = m.Get(name)
		require.NoError(t, err)
	}

	// 获取列表
	list := m.List()
	assert.Len(t, list, 3)

	// 验证列表包含所有业务名称
	for _, name := range bizNames {
		assert.Contains(t, list, name)
	}
}

// TestManager_Remove 测试移除方法
func TestManager_Remove(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{
		Dir:    tempDir,
		Level:  "info",
		Format: "json",
	}

	m, err := NewManager(cfg)
	require.NoError(t, err)

	// 添加logger
	_, err = m.Get("test")
	require.NoError(t, err)
	assert.Len(t, m.List(), 1)

	// 移除logger
	err = m.Remove("test")
	assert.NoError(t, err)
	assert.Empty(t, m.List())

	// 移除不存在的logger
	err = m.Remove("nonexistent")
	assert.NoError(t, err)
}

// TestManager_SetLevel 测试设置日志级别
func TestManager_SetLevel(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{
		Dir:    tempDir,
		Level:  "info",
		Format: "json",
	}

	m, err := NewManager(cfg)
	require.NoError(t, err)

	// 创建logger
	logger, err := m.Get("test")
	require.NoError(t, err)

	tests := []struct {
		name    string
		bizName string
		level   string
		wantErr bool
		errType error
	}{
		{
			name:    "设置有效级别",
			bizName: "test",
			level:   "debug",
			wantErr: false,
		},
		{
			name:    "空业务名称",
			bizName: "",
			level:   "debug",
			wantErr: true,
			errType: ErrEmptyBizName,
		},
		{
			name:    "无效级别",
			bizName: "test",
			level:   "invalid",
			wantErr: true,
			errType: ErrInvalidLogLevel,
		},
		{
			name:    "logger不存在",
			bizName: "nonexistent",
			level:   "debug",
			wantErr: true,
			errType: ErrLoggerNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := m.SetLevel(tt.bizName, tt.level)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, IsEmptyBizName(err) || IsInvalidLogLevel(err) || IsLoggerNotFound(err))
			} else {
				assert.NoError(t, err)
				// 验证级别确实被修改了
				currentLevel, err := m.GetLevel(tt.bizName)
				assert.NoError(t, err)
				assert.Equal(t, tt.level, currentLevel)
			}
		})
	}

	// 测试日志级别确实生效
	initialLevel := logger.Core().Enabled(zapcore.InfoLevel)
	m.SetLevel("test", "error")
	newLevel := logger.Core().Enabled(zapcore.InfoLevel)
	assert.True(t, initialLevel, "info级别最初应该启用")
	assert.False(t, newLevel, "设置为error级别后，info级别应该被禁用")
}

// TestManager_GetLevel 测试获取日志级别
func TestManager_GetLevel(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{
		Dir:    tempDir,
		Level:  "info",
		Format: "json",
	}

	m, err := NewManager(cfg)
	require.NoError(t, err)

	// 创建logger
	_, err = m.Get("test")
	require.NoError(t, err)

	tests := []struct {
		name    string
		bizName string
		wantErr bool
		errType error
	}{
		{
			name:    "获取存在的logger级别",
			bizName: "test",
			wantErr: false,
		},
		{
			name:    "空业务名称",
			bizName: "",
			wantErr: true,
			errType: ErrEmptyBizName,
		},
		{
			name:    "logger不存在",
			bizName: "nonexistent",
			wantErr: true,
			errType: ErrLoggerNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, err := m.GetLevel(tt.bizName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, IsEmptyBizName(err) || IsLoggerNotFound(err))
				assert.Empty(t, level)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, level)
				assert.Equal(t, "info", level) // 初始级别应该是info
			}
		})
	}
}

// TestInit 测试全局初始化
func TestInit(t *testing.T) {
	// 重置全局状态
	defaultManager = nil
	defaultManagerOnce = sync.Once{}

	tempDir := t.TempDir()
	cfg := Config{
		Dir:    tempDir,
		Level:  "info",
		Format: "json",
	}

	// 测试正常初始化
	assert.NotPanics(t, func() {
		Init(cfg)
	})

	assert.NotNil(t, Default())

	// 测试多次初始化（应该只初始化一次）
	assert.NotPanics(t, func() {
		Init(Config{Dir: "/tmp/other"})
	})

	// defaultManager应该还是第一次的值
	assert.Equal(t, tempDir, Default().cfg.Dir)
}

// TestDefault 测试默认manager
func TestDefault(t *testing.T) {
	// 重置全局状态
	defaultManager = nil
	defaultManagerOnce = sync.Once{}

	// 未初始化时应该返回nil
	assert.Nil(t, Default())

	// 初始化后应该返回manager
	tempDir := t.TempDir()
	Init(Config{Dir: tempDir})
	assert.NotNil(t, Default())
}

// TestManager_Integration 集成测试
func TestManager_Integration(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{
		Dir:        tempDir,
		Level:      "info",
		Format:     "json",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
		Console:    true,
	}

	m, err := NewManager(cfg)
	require.NoError(t, err)

	// 创建多个logger
	services := []string{"auth", "api", "db", "cache"}
	for _, service := range services {
		logger, err := m.Get(service)
		require.NoError(t, err)
		assert.NotNil(t, logger)

		// 写入一些日志
		logger.Info("测试日志", zap.String("service", service))
	}

	// 验证列表
	list := m.List()
	assert.Len(t, list, len(services))

	// 测试级别动态调整
	for _, service := range services {
		err := m.SetLevel(service, "error")
		assert.NoError(t, err)

		level, err := m.GetLevel(service)
		assert.NoError(t, err)
		assert.Equal(t, "error", level)
	}

	// 测试同步
	err = m.Sync()
	// 在测试环境中，stdout的sync可能会失败，这是正常的
	if err != nil {
		t.Logf("同步时出现预期中的错误: %v", err)
	}

	// 验证日志文件是否创建
	for _, service := range services {
		logFile := filepath.Join(tempDir, service+".log")
		_, err := os.Stat(logFile)
		assert.NoError(t, err, "日志文件 %s 应该被创建", logFile)
	}

	// 测试关闭
	err = m.Close()
	// 在测试环境中，stdout的sync可能会失败，这是正常的
	if err != nil {
		t.Logf("关闭时出现预期中的错误: %v", err)
	}
	assert.Empty(t, m.List())
}

// BenchmarkManager_Get 基准测试：获取logger性能
func BenchmarkManager_Get(b *testing.B) {
	tempDir := b.TempDir()
	cfg := Config{
		Dir:    tempDir,
		Level:  "info",
		Format: "json",
	}

	m, err := NewManager(cfg)
	require.NoError(b, err)

	// 预创建一些logger
	bizNames := make([]string, 100)
	for i := 0; i < 100; i++ {
		bizNames[i] = fmt.Sprintf("service_%d", i)
		m.Get(bizNames[i])
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			bizName := bizNames[i%100]
			m.Get(bizName)
			i++
		}
	})
}

// BenchmarkManager_Get_Cached 基准测试：缓存命中的性能
func BenchmarkManager_Get_Cached(b *testing.B) {
	tempDir := b.TempDir()
	cfg := Config{
		Dir:    tempDir,
		Level:  "info",
		Format: "json",
	}

	m, err := NewManager(cfg)
	require.NoError(b, err)

	// 预创建logger
	m.Get("cached_service")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get("cached_service")
	}
}

// BenchmarkManager_SetLevel 基准测试：设置日志级别性能
func BenchmarkManager_SetLevel(b *testing.B) {
	tempDir := b.TempDir()
	cfg := Config{
		Dir:    tempDir,
		Level:  "info",
		Format: "json",
	}

	m, err := NewManager(cfg)
	require.NoError(b, err)

	// 预创建logger
	m.Get("benchmark_service")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		level := "info"
		if i%2 == 0 {
			level = "debug"
		}
		m.SetLevel("benchmark_service", level)
	}
}
