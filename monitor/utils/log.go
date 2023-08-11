package utils

import (
	"fmt"
	"time"
)

func Infof(format string, a ...interface{}) {
	a = append([]interface{}{time.Now()}, a...)
	s := fmt.Sprintf("%s [info] "+format, a...)
	LogWithColor(s, 32)
}

func Warnf(format string, a ...interface{}) {
	a = append([]interface{}{time.Now()}, a...)
	s := fmt.Sprintf("%s [warn] "+format, a...)
	LogWithColor(s, 33)
}

func Errorf(format string, a ...interface{}) {
	a = append([]interface{}{time.Now()}, a...)
	s := fmt.Sprintf("%s [error] "+format, a...)
	LogWithColor(s, 31)
}

func LogWithColor(s string, c int64) {
	fmt.Printf("\033[1;%d;40m%s\033[0m\n", c, s)
}
