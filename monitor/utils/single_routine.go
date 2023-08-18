package utils

import "context"

type SingleRoutine struct {
	cancel context.CancelFunc
}

func NewSingleRoutine(ctx context.Context) *SingleRoutine {
	return &SingleRoutine{}
}

func (r *SingleRoutine) Run(ctx context.Context, f func(context.Context)) {
	if r.cancel != nil {
		r.cancel()
	}
	ctx, r.cancel = context.WithCancel(ctx)
	go f(ctx)
}
