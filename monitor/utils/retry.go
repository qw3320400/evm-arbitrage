package utils

import (
	"context"
	"fmt"
	"time"
)

func Retry(ctx context.Context, function func(context.Context) error, waitTime time.Duration, retyCount int64) error {
	var retryLeftCount = retyCount
	for {
		err := function(ctx)
		if err == nil {
			return nil
		}
		if retryLeftCount != -1 {
			retryLeftCount--
			if retryLeftCount <= 0 {
				return fmt.Errorf("retry %d times and fail %s", retyCount, err)
			}
		}
		Infof("try function fail, waiting for retry %d %s", retryLeftCount, err)
		<-time.After(waitTime)
	}
}
