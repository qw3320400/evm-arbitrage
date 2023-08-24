package utils

import "context"

type Keeper interface {
	Init(ctx context.Context) error
	ShutDown(ctx context.Context)
}
