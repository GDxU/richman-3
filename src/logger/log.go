package logger

import (
	"fmt"
	"log"
	"os"
)

var Now string

func GetLogger(s string) *log.Logger {
	os.Chdir("log")
	logFile, err := os.OpenFile(Now, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
	}
	logger := log.New(logFile, s, log.Ldate|log.Ltime|log.Lshortfile)
	return logger
}
