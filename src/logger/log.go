package logger

import (
	"fmt"
	"log"
	"os"
)

// Now is the time when the process executed.
var Now string

// GetLogger returns a logger in log package.
// the logger writes logs on a log file named executed time
func GetLogger(s string) *log.Logger {
	os.Chdir("log")
	logFile, err := os.OpenFile(Now, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
	}
	logger := log.New(logFile, s, log.Ldate|log.Ltime|log.Lshortfile)
	return logger
}
