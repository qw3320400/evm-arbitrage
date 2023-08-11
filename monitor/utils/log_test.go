package utils

import "testing"

func TestLog(t *testing.T) {
	Infof("abcd")
	Warnf("test %d", 123)
	Errorf("fail %s", "111")
}
