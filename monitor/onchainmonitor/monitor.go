package onchainmonitor

import (
	"context"
)

type Monitor interface {
	Init(context.Context) error
}
