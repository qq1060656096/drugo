package log

import (
	"errors"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// 日志管理器，用于管理多个业务模块的日志实例
type Manager struct {
	mu      sync.RWMutex               // 读写锁，用于并发安全
	cfg     Config                     // 日志配置
	loggers map[string]*zap.Logger     // 日志实例缓存，按业务名称分组
	levels  map[string]zap.AtomicLevel // 日志级别控制器，用于动态调整级别
}

var (
	defaultManager     *Manager
	defaultManagerOnce sync.Once
)

// Init 使用给定的配置初始化全局默认 Manager。
// 如果初始化失败，它会 panic。此函数是并发安全的，只会初始化一次。
func Init(cfg Config) {
	defaultManagerOnce.Do(func() {
		defaultManager = MustNewManager(cfg)
	})
}

// Default 返回全局默认 Manager 实例。
// 如果 Init 未被调用，返回 nil。
func Default() *Manager {
	return defaultManager
}

// NewManager 创建新的日志管理器实例
// cfg: 日志配置
// 返回: 日志管理器指针和可能的错误
func NewManager(cfg Config) (*Manager, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Manager{
		cfg:     cfg,
		loggers: make(map[string]*zap.Logger),     // 初始化日志实例缓存
		levels:  make(map[string]zap.AtomicLevel), // 初始化日志级别控制器
	}, nil
}

// MustNewManager 类似于 NewManager，但如果发生错误会 panic。
func MustNewManager(cfg Config) *Manager {
	m, err := NewManager(cfg)
	if err != nil {
		panic(err)
	}
	return m
}

// Get 获取指定业务名称的日志实例
// bizName: 业务名称，用于标识不同的日志实例
// 返回: zap日志实例和可能的错误
func (m *Manager) Get(bizName string) (*zap.Logger, error) {
	// 验证业务名称不为空
	if bizName == "" {
		return nil, ErrEmptyBizName
	}

	// 先使用读锁检查缓存
	m.mu.RLock()
	l, ok := m.loggers[bizName]
	m.mu.RUnlock()
	if ok {
		return l, nil // 缓存命中，直接返回
	}

	// 缓存未命中，使用写锁创建新实例
	m.mu.Lock()
	defer m.mu.Unlock()

	// 双重检查，防止在获取写锁期间其他goroutine已经创建了实例
	if logger, ok := m.loggers[bizName]; ok {
		return logger, nil
	}

	// 创建新的zap日志实例
	l, level, err := NewZapLogger(m.cfg, bizName)
	if err != nil {
		return nil, err
	}

	// 将新创建的日志实例和级别控制器存入缓存
	m.loggers[bizName] = l
	m.levels[bizName] = level
	return l, nil
}

// MustGet 获取指定业务名称的日志实例，如果出错会panic
// bizName: 业务名称
// 返回: zap日志实例
func (m *Manager) MustGet(bizName string) *zap.Logger {
	l, err := m.Get(bizName)
	if err != nil {
		panic(err) // 出错时panic，用于必须成功的场景
	}
	return l
}

// Sync 同步所有日志实例，将缓冲区的日志刷新到磁盘
// 建议在程序退出前调用此方法，确保所有日志都被写入
// 返回: 同步过程中的所有错误（合并后）
func (m *Manager) Sync() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errs []error
	for bizName, logger := range m.loggers {
		if err := logger.Sync(); err != nil {
			// 忽略 stdout/stderr 的 sync 错误（在某些系统上是正常的）
			// 但记录其他文件的同步错误
			if bizName != "stdout" && bizName != "stderr" {
				errs = append(errs, fmt.Errorf("sync logger '%s': %w", bizName, err))
			}
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// Close 关闭所有日志实例，同步缓冲区并释放资源
// 调用后将清空日志实例缓存，后续调用 Get() 会创建新的实例
// 建议在程序退出时调用此方法
// 返回: 关闭过程中的所有错误（合并后）
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for bizName, logger := range m.loggers {
		if err := logger.Sync(); err != nil {
			// 忽略 stdout/stderr 的 sync 错误（在某些系统上是正常的）
			if bizName != "stdout" && bizName != "stderr" {
				errs = append(errs, fmt.Errorf("close logger '%s': %w", bizName, err))
			}
		}
	}

	// 清空日志实例缓存和级别控制器
	m.loggers = make(map[string]*zap.Logger)
	m.levels = make(map[string]zap.AtomicLevel)

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// List 列出所有已创建的日志实例名称
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.loggers))
	for name := range m.loggers {
		names = append(names, name)
	}
	return names
}

// Remove 移除指定的日志实例
func (m *Manager) Remove(bizName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if logger, ok := m.loggers[bizName]; ok {
		if err := logger.Sync(); err != nil {
			return err
		}
		delete(m.loggers, bizName)
		delete(m.levels, bizName)
	}
	return nil
}

// SetLevel 动态更新指定业务的日志级别
// bizName: 业务名称
// level: 新的日志级别字符串，如 "debug", "info", "warn", "error"
// 返回: 可能的错误
func (m *Manager) SetLevel(bizName, level string) error {
	// 验证业务名称不为空
	if bizName == "" {
		return ErrEmptyBizName
	}

	// 解析新的日志级别
	newLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log level '%s': %w", level, ErrInvalidLogLevel)
	}

	m.mu.RLock()
	atomicLevel, ok := m.levels[bizName]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("logger '%s': %w", bizName, ErrLoggerNotFound)
	}

	// 动态更新日志级别
	atomicLevel.SetLevel(newLevel.Level())
	return nil
}

// GetLevel 获取指定业务的当前日志级别
// bizName: 业务名称
// 返回: 日志级别字符串和可能的错误
func (m *Manager) GetLevel(bizName string) (string, error) {
	// 验证业务名称不为空
	if bizName == "" {
		return "", ErrEmptyBizName
	}

	m.mu.RLock()
	atomicLevel, ok := m.levels[bizName]
	m.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("logger '%s': %w", bizName, ErrLoggerNotFound)
	}

	return atomicLevel.Level().String(), nil
}
