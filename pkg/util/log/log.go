package log

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
)

type Level int32

const (
	TraceLevel = Level(-5)
	DebugLevel = Level(log.DebugLevel)
	InfoLevel  = Level(log.InfoLevel)
	WarnLevel  = Level(log.WarnLevel)
	ErrorLevel = Level(log.ErrorLevel)
	FatalLevel = Level(log.FatalLevel)
)

var logger = log.Default()

func SetLevel(l Level) {
	logger.SetLevel(log.Level(l))
}

// Trace logs a trace message.
func Trace(msg interface{}, keyvals ...any) {
	logger.Log(log.Level(TraceLevel), msg, keyvals...)
}

// Debug logs a debug message.
func Debug(msg interface{}, keyvals ...any) {
	logger.Log(log.DebugLevel, msg, keyvals...)
}

// Info logs an info message.
func Info(msg interface{}, keyvals ...any) {
	logger.Log(log.InfoLevel, msg, keyvals...)
}

// Warn logs a warning message.
func Warn(msg interface{}, keyvals ...any) {
	logger.Log(log.WarnLevel, msg, keyvals...)
}

// Error logs an error message.
func Error(msg interface{}, keyvals ...any) {
	logger.Log(log.ErrorLevel, msg, keyvals...)
}

// Fatal logs a fatal message and exit.
func Fatal(msg interface{}, keyvals ...any) {
	logger.Log(log.FatalLevel, msg, keyvals...)
	os.Exit(1)
}

// Tracef logs a trace message with formatting.
func Tracef(format string, args ...any) {
	logger.Log(log.Level(TraceLevel), fmt.Sprintf(format, args...))
}

// Debugf logs a debug message with formatting.
func Debugf(format string, args ...any) {
	logger.Log(log.DebugLevel, fmt.Sprintf(format, args...))
}

// Infof logs an info message with formatting.
func Infof(format string, args ...any) {
	logger.Log(log.InfoLevel, fmt.Sprintf(format, args...))
}

// Warnf logs a warning message with formatting.
func Warnf(format string, args ...any) {
	logger.Log(log.WarnLevel, fmt.Sprintf(format, args...))
}

// Errorf logs an error message with formatting.
func Errorf(format string, args ...any) {
	logger.Log(log.ErrorLevel, fmt.Sprintf(format, args...))
}

// Fatalf logs a fatal message with formatting and exit.
func Fatalf(format string, args ...any) {
	logger.Log(log.FatalLevel, fmt.Sprintf(format, args...))
	os.Exit(1)
}
