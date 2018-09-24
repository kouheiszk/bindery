package main

import (
	"fmt"
)

const (
	LogLevelDebug = 0
	LogLevelInfo  = 1
	LogLevelError = 2
)

var minLogLevel_ int

func log(level int, s string, a ...interface{}) {
	if level < minLogLevel_ {
		return
	}
	fmt.Printf(Name+": "+s+"\n", a...)
}

func logDebug(s string, a ...interface{}) {
	log(LogLevelDebug, s, a...)
}

func logInfo(s string, a ...interface{}) {
	log(LogLevelInfo, s, a...)
}

func logError(s string, a ...interface{}) {
	log(LogLevelError, s, a...)
}
