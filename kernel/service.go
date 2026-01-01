package kernel

import (
	"context"
	"fmt"
)

// Booter 定义了具有初始化能力的组件。
// Boot 方法应在使用前调用一次。
type Booter interface {
	Boot(ctx context.Context) error
}

// Closer 定义了资源释放接口，与 io.Closer 类似，但支持上下文。
// Close 方法应当是幂等的。
type Closer interface {
	Close(ctx context.Context) error
}

// Service 组合了基础的生命周期行为。
// 它代表一个拥有名称、可初始化且可关闭的管理单元。
type Service interface {
	Name() string
	Boot(ctx context.Context) error
	Close(ctx context.Context) error
}

// Runner 描述了一个长期运行的服务。
// Run 方法应当阻塞，直到上下文取消或发生不可恢复的错误。
type Runner interface {
	Service
	Run(ctx context.Context) error
}

func GetService[T any](k Kernel, name string) (T, error) {
	var zero T
	svc, err := k.Container().Get(name)
	if err != nil {
		return zero, err
	}
	concreteSvc, ok := svc.(T)
	if !ok {
		return zero, fmt.Errorf("service %s is not of type %T %w", name, zero, NewServiceType(name))
	}
	return concreteSvc, nil
}

func MustGetService[T any](k Kernel, name string) T {
	svc, err := GetService[T](k, name)
	if err != nil {
		panic(err)
	}
	return svc
}
