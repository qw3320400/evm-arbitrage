package action

import (
	"context"
)

type Action interface {
	Init(context.Context) error
	OnNewBlockHandler(context.Context, []uint64) error
}
