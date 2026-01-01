package drugo

import (
	"context"
	"time"

	"github.com/qq1060656096/drugo/kernel"
)

// DefaultShutdownTimeout 默认优雅停机超时时间
const DefaultShutdownTimeout = 10 * time.Second

type options struct {
	root string
	// Changed to a simple map for easier registration
	services        []map[string]kernel.Service
	ctx             context.Context
	shutdownTimeout time.Duration
}

type Option func(*options)

func WithRoot(root string) Option {
	return func(o *options) {
		o.root = root
	}
}

func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

func WithNameService(name string, service kernel.Service) Option {
	return func(o *options) {
		if o.services == nil {
			o.services = make([]map[string]kernel.Service, 0)
		}
		o.services = append(o.services, map[string]kernel.Service{name: service})
	}
}

func WithService(service kernel.Service) Option {
	return WithNameService(service.Name(), service)
}

// WithShutdownTimeout 设置优雅停机的超时时间
// 如果不设置，默认使用 DefaultShutdownTimeout (10秒)
func WithShutdownTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.shutdownTimeout = timeout
	}
}
