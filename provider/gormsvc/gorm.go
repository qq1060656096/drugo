package gormsvc

import (
	"context"

	"github.com/qq1060656096/drugo/kernel"
)

var _ kernel.Service = (*GormService)(nil)

type GormService struct {
}

func (b *GormService) Name() string {
	return "gorm"
}

func (b *GormService) Boot(ctx context.Context) error {
	return nil
}

func (b *GormService) Close(ctx context.Context) error {
	return nil
}
