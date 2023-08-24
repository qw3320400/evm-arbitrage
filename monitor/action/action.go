package action

import (
	"context"
)

type Action interface {
	Init(context.Context) error
	OnNewBlockHandler(context.Context, ...interface{}) error
	OnNewLogHandler(context.Context, ...interface{}) error
}
