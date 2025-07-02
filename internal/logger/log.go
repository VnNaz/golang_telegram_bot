package logger

import "log"

type Logger struct {
	log []*log.Logger
}

// implement middleware template
