package logging

import (
	"log"
	"os"
)

//Logger is a generic function logger
type Logger func(...interface{})

//Specified loggers
var (
	debugLogger Logger = nil
	errorLogger Logger = nil
)

//ConsoleDebugLogger returns a stdout debug logger
func ConsoleDebugLogger() Logger {
	return log.New(os.Stdout, "Gocompositor-DEBUG: ", log.Ldate|log.Ltime).Println
}

//ConsoleErrorLogger returns a stderr error logger
func ConsoleErrorLogger() Logger {
	return log.New(os.Stderr, "Gocompositor-ERROR: ", log.Ldate|log.Ltime).Println
}

//Debug logs an debug info
func Debug(args ...interface{}) {
	if debugLogger != nil {
		debugLogger(args...)
	}
}

//Error logs an error
func Error(args ...interface{}) {
	if errorLogger != nil {
		errorLogger(args...)
	}
}

//SetDebugLogger sets the debug logger function
func SetDebugLogger(logger Logger) {
	debugLogger = logger
}

//SetErrorLogger sets the error logger function
func SetErrorLogger(logger Logger) {
	errorLogger = logger
}
