package kernel

// Container 表示一个通用的服务容器，用于管理满足 Service 或 RunnerService 约束的实例。
// T 约束确保了存入容器的对象具备预定义的行为。
type Container[T Service] interface {
	// Bind 将给定的服务实例与名称绑定。
	// 如果容器中已存在同名服务，原有的绑定将被新的服务覆盖。
	Bind(name string, service T)

	// Get 根据名称查找并返回服务实例。
	// 如果找不到对应的服务，则返回该类型的零值并附带一个错误。
	Get(name string) (T, error)

	// MustGet 尝试根据名称获取服务，如果获取失败（如服务不存在），则直接触发 panic。
	// 该方法适用于系统初始化等必须确保依赖存在的场景。
	MustGet(name string) T

	// Services 返回所有已注册的服务实例
	Services() []T

	// Names 返回所有已注册的服务名称
	Names() []string
}
