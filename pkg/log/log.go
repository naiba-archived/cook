package log

import "fmt"

// Printf ...
func Printf(format string, args ...interface{}) {
	fmt.Printf("🐮[LOG]"+format+"\n", args...)
}
