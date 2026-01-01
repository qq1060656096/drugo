package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// ReloadCallback 是配置重载时调用的回调函数类型。
// 如果回调返回 error，错误会被记录但不会停止热加载。
type ReloadCallback func(m *Manager) error

// Manager 管理配置加载和缓存，支持多业务配置。
type Manager struct {
	mu        sync.RWMutex
	root      *viper.Viper
	configs   map[string]*viper.Viper
	configDir string

	// 热加载相关字段
	watcher         *fsnotify.Watcher
	watcherDone     chan struct{}
	watcherStopOnce sync.Once
	reloadCallbacks []ReloadCallback
}

var (
	defaultManager     *Manager
	defaultManagerOnce sync.Once
)

// Init 使用给定的配置目录初始化全局默认 Manager。
// 如果初始化失败，它会 panic。此函数是并发安全的，只会初始化一次。
func Init(configDir string) {
	defaultManagerOnce.Do(func() {
		defaultManager = MustNewManager(configDir)
	})
}

// Default 返回全局默认 Manager 实例。
// 如果 Init 未被调用，返回 nil。
func Default() *Manager {
	return defaultManager
}

// NewManager 创建一个新的 Manager，从 configDir 读取配置文件。
// 它读取目录中所有 .yml 和 .yaml 文件并合并它们。
func NewManager(configDir string) (*Manager, error) {
	root, err := loadConfigs(configDir)
	if err != nil {
		return nil, err
	}
	return &Manager{
		root:      root,
		configs:   make(map[string]*viper.Viper),
		configDir: configDir,
	}, nil
}

// MustNewManager 类似于 NewManager，但如果发生错误会 panic。
func MustNewManager(configDir string) *Manager {
	m, err := NewManager(configDir)
	if err != nil {
		panic(err)
	}
	return m
}

// Get 返回指定业务名称的配置。
// 它使用双重检查锁定来保证线程安全和缓存。
func (m *Manager) Get(name string) (*viper.Viper, error) {
	// 快速路径：使用读锁检查缓存。
	m.mu.RLock()
	v, ok := m.configs[name]
	m.mu.RUnlock()
	if ok {
		return v, nil
	}

	// 慢速路径：获取写锁并再次检查。
	m.mu.Lock()
	defer m.mu.Unlock()

	if v, ok = m.configs[name]; ok {
		return v, nil
	}

	sub := m.root.Sub(name)
	if sub == nil {
		return nil, fmt.Errorf("%w: %q", ErrNotFound, name)
	}

	m.configs[name] = sub
	return sub, nil
}

// MustGet 类似于 Get，但如果发生错误会 panic。
func (m *Manager) MustGet(name string) *viper.Viper {
	v, err := m.Get(name)
	if err != nil {
		panic(err)
	}
	return v
}

// Root 返回包含所有业务配置的根配置。
func (m *Manager) Root() *viper.Viper {
	return m.root
}

// List 返回根配置中所有可用业务配置名称的有序列表，
// 无论它们是否已被加载。
func (m *Manager) List() []string {
	settings := m.root.AllSettings()
	names := make([]string, 0, len(settings))
	for name := range settings {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Reset 重新加载配置并清空所有缓存的业务配置。
// 此方法是线程安全的。
func (m *Manager) Reset() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	root, err := loadConfigs(m.configDir)
	if err != nil {
		return err
	}

	m.root = root
	m.configs = make(map[string]*viper.Viper)
	return nil
}

// OnReload 注册配置重载时的回调函数。
// 回调函数会在配置文件变化并成功重载后被调用。
// 此方法是线程安全的。
func (m *Manager) OnReload(callback ReloadCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reloadCallbacks = append(m.reloadCallbacks, callback)
}

// Watch 启动配置文件的热加载监听。
// 当配置文件发生变化时，会自动重新加载配置并调用注册的回调函数。
// 此方法是幂等的，多次调用只会启动一次监听。
func (m *Manager) Watch() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果已经在监听，直接返回
	if m.watcher != nil {
		return nil
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("config: failed to create watcher: %w", err)
	}

	// 添加配置目录到监听列表
	if err := watcher.Add(m.configDir); err != nil {
		watcher.Close()
		return fmt.Errorf("config: failed to watch directory %s: %w", m.configDir, err)
	}

	m.watcher = watcher
	m.watcherDone = make(chan struct{})

	// 启动监听协程
	go m.watchLoop()

	return nil
}

// StopWatch 停止配置文件的热加载监听。
// 此方法是幂等的，多次调用是安全的。
func (m *Manager) StopWatch() {
	m.watcherStopOnce.Do(func() {
		m.mu.Lock()
		if m.watcher != nil {
			m.watcher.Close()
		}
		if m.watcherDone != nil {
			close(m.watcherDone)
		}
		m.mu.Unlock()
	})
}

// watchLoop 是监听配置文件变化的主循环。
func (m *Manager) watchLoop() {
	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}

			// 只处理写入和创建事件
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				// 检查是否是 YAML 文件
				ext := filepath.Ext(event.Name)
				if ext == ".yml" || ext == ".yaml" {
					m.handleReload()
				}
			}

		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			// 这里可以选择记录错误或者调用错误回调
			fmt.Fprintf(os.Stderr, "config watcher error: %v\n", err)

		case <-m.watcherDone:
			return
		}
	}
}

// handleReload 处理配置重载逻辑。
func (m *Manager) handleReload() {
	// 重新加载配置
	if err := m.Reset(); err != nil {
		fmt.Fprintf(os.Stderr, "config reload failed: %v\n", err)
		return
	}

	// 调用所有注册的回调函数
	m.mu.RLock()
	callbacks := make([]ReloadCallback, len(m.reloadCallbacks))
	copy(callbacks, m.reloadCallbacks)
	m.mu.RUnlock()

	for _, callback := range callbacks {
		if err := callback(m); err != nil {
			fmt.Fprintf(os.Stderr, "config reload callback error: %v\n", err)
		}
	}
}

// loadConfigs 从给定目录读取所有 YAML 配置文件，
// 并将它们合并到单个 viper 实例中。
func loadConfigs(dir string) (*viper.Viper, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrDirRead, dir, err)
	}

	root := viper.New()

	for _, fileInfo := range entries {
		if fileInfo.IsDir() {
			continue
		}

		fileExt := filepath.Ext(fileInfo.Name())
		if fileExt != ".yml" && fileExt != ".yaml" {
			continue
		}

		filePath := filepath.Join(dir, fileInfo.Name())
		if err := mergeFile(root, filePath); err != nil {
			return nil, err
		}
	}

	return root, nil
}

// mergeFile 读取单个配置文件并将其内容合并到 root 中。
// 文件中的每个顶级键代表一个业务配置。
func mergeFile(root *viper.Viper, path string) error {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("%w: %s: %v", ErrFileRead, path, err)
	}

	for name := range v.AllSettings() {
		if root.IsSet(name) {
			return fmt.Errorf("%w: %q in %s", ErrDuplicateKey, name, path)
		}

		sub := v.Sub(name)
		if sub == nil {
			return fmt.Errorf("%w: %s: cannot read sub config %q", ErrFileRead, path, name)
		}

		root.Set(name, sub.AllSettings())
	}

	return nil
}
