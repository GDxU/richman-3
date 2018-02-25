package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Now is the time when the process executed.
// Coin is the Coin name
var (
	Now  string
	Coin string
)

// Loggers is a set of loggers.
type Loggers struct {
	Info    *log.Logger
	Warning *log.Logger
	Severe  *log.Logger
}

// GetLogger returns a logger in log package.
// the logger writes logs on a log file named executed time
func GetLogger(s string) *Loggers {
	os.Chdir("log/" + strings.ToLower(Coin))
	infoLogFile, err := os.OpenFile("[Info]"+Now, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	warningLogFile, err := os.OpenFile("[Warning]"+Now, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	severeLogFile, err := os.OpenFile("[Severe]"+Now, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
	}
	infoLogger := log.New(infoLogFile, s, log.Ldate|log.Ltime|log.Lshortfile)
	warningLogger := log.New(warningLogFile, s, log.Ldate|log.Ltime|log.Lshortfile)
	severeLogger := log.New(severeLogFile, s, log.Ldate|log.Ltime|log.Lshortfile)
	loggers := Loggers{
		Info:    infoLogger,
		Warning: warningLogger,
		Severe:  severeLogger,
	}
	return &loggers
}
