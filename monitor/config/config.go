package config

import "time"

type Config struct {
	Node     string
	Loop     bool
	LoopTime time.Duration
}
