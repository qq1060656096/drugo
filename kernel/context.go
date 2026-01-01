package kernel

import (
	"context"
)

type kernelCtxKey struct{}

func WithContext(ctx context.Context, kernel Kernel) context.Context {
	return context.WithValue(ctx, kernelCtxKey{}, kernel)
}

func FromContext(ctx context.Context) (Kernel, bool) {
	k, ok := ctx.Value(kernelCtxKey{}).(Kernel)
	return k, ok
}

func MustFromContext(ctx context.Context) Kernel {
	k, ok := FromContext(ctx)
	if !ok {
		panic(NewKernelNotInContext())
	}
	return k
}

func ServiceFromContext[T any](ctx context.Context, name string) (T, error) {
	var zero T
	k, ok := FromContext(ctx)
	if !ok {
		return zero, NewKernelNotInContext()
	}

	if k == nil {
		return zero, NewKernelNotInContext()
	}

	return GetService[T](k, name)
}

func MustServiceFromContext[T any](ctx context.Context, name string) T {
	svc, err := ServiceFromContext[T](ctx, name)
	if err != nil {
		panic(err)
	}
	return svc
}
