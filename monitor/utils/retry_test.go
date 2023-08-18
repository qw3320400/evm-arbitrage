package utils

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestRetry(t *testing.T) {
	ctx := context.Background()
	c := 0
	f := func(context.Context) error {
		c++
		if c != 5 {
			return fmt.Errorf("function fail")
		}
		return nil
	}
	err := Retry(ctx, f, time.Millisecond, -1)
	if err != nil {
		t.Fatal(err)
	}
	if c != 5 {
		t.Fatal(c)
	}

	c = 0
	err = Retry(ctx, f, time.Millisecond, 3)
	if err == nil {
		t.Fatal("should fail")
	}
}
