package logger

import "fmt"

var ENABLE_LOGGING = true

func Println(args ...interface{}) {
	if !ENABLE_LOGGING {
		return
	}
	fmt.Println(args...)
}

func Printf(format string, args ...interface{}) {
	if !ENABLE_LOGGING {
		return
	}
	fmt.Printf(format, args...)
}
