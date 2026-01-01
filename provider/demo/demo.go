package demo

import (
	"context"

	"github.com/qq1060656096/drugo/kernel"
)

var _ kernel.Service = (*Demo)(nil)

type Demo struct {
}

func (*Demo) Name() string {
	return "demo"
}

func (*Demo) Boot(ctx context.Context) error {
	k := kernel.MustFromContext(ctx)
	k.Logger().MustGet("demo").Info("boot demo")
	return nil
}

func (*Demo) Close(ctx context.Context) error {
	k := kernel.MustFromContext(ctx)
	k.Logger().MustGet("demo").Info("close demo")
	return nil
}

func New() *Demo {
	return &Demo{}
}
