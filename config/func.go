package config

import "fmt"

// Config 获取指定 name 对应的配置，并反序列化为泛型类型 T。
//
// 该方法会从 Manager 中读取配置项，然后将其反序列化到指定类型。
// 适用于将配置直接映射为结构体或基础类型。
//
// 参数：
//   - m: 配置管理器实例，用于读取配置数据。
//   - name: 配置名称（key）。
//
// 返回值：
//   - cfg: 反序列化后的配置对象（类型 T）。
//   - err: 错误信息：
//   - 获取配置失败时返回底层错误。
//   - 反序列化失败时返回带有配置名的包装错误。
//
// 示例：
//
//	type DBConfig struct {
//	    Host string
//	    Port int
//	}
//
//	cfg, err := Config[DBConfig](manager, "database")
//	if err != nil {
//	    panic(err)
//	}
func Config[T any](m *Manager, name string) (cfg T, err error) {
	v, err := m.Get(name)
	if err != nil {
		return
	}

	if err = v.Unmarshal(&cfg); err != nil {
		err = fmt.Errorf("config %q: unmarshal: %w", name, err)
	}

	return
}

// MustConfig 获取指定 name 对应的配置并反序列化为类型 T，失败时直接 panic。
//
// 该方法是 Config 的快捷版本，适用于配置必须存在且加载失败不可恢复的场景
// （如应用启动初始化、全局配置加载等）。
//
// 内部调用 Config：
//   - 获取配置失败时 panic。
//   - 反序列化失败时 panic。
//
// 参数：
//   - m: 配置管理器实例。
//   - name: 配置名称（key）。
//
// 返回值：
//   - cfg: 反序列化后的配置对象（类型 T）。
//
// 使用场景：
//   - 程序启动阶段加载配置。
//   - 必须存在的系统配置。
//   - 测试代码或内部工具。
//
// 示例：
//
//	cfg := MustConfig[DBConfig](manager, "database")
func MustConfig[T any](m *Manager, name string) T {
	cfg, err := Config[T](m, name)
	if err != nil {
		panic(err)
	}
	return cfg
}
