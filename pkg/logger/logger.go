package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// Handle the given log if its level is equals or higher than minimum acceptable level.
//
// Log levels in order are:
// Debug:
// Used for detailed information during development and debugging.
// Typically not enabled in production.
//
// Info:
// Used for general informational messages about the application's state.
// Indicates normal operation.
//
// Warn (or Warning):
// Used to indicate potential issues or unexpected situations that don't necessarily cause errors.
// May require attention.
//
// Error:
// Used to log errors that occurred during the application's execution.
// Indicates a problem that needs to be addressed.
//
// Fatal:
// Used for critical errors that prevent the application from continuing.
// Typically followed by program termination and then calls os.Exit(1).
//
// Panic:
// Used for errors that cause the application to panic.
// Indicates a severe, unrecoverable error.
type Logger interface {
	// Debug logs a message at the debug level.
	Debug(args ...any)

	// Debugf logs a formatted message at the debug level.
	Debugf(format string, args ...any)

	// Info logs a message at the info level.
	Info(args ...any)

	// Infof logs a formatted message at the info level.
	Infof(format string, args ...any)

	// Warn logs a message at the warning level.
	Warn(args ...any)

	// Warnf logs a formatted message at the warning level.
	Warnf(format string, args ...any)

	// Error logs a message at the error level.
	Error(args ...any)

	// Errorf logs a formatted message at the error level.
	Errorf(format string, args ...any)

	// Fatal logs a message at the fatal level and then calls os.Exit(1).
	// Whether the log processes or not, always call os.Exit(1) to terminate program.
	Fatal(args ...any)

	// Fatalf logs a formatted message at the fatal level and then calls os.Exit(1).
	// Whether the log processes or not, always call os.Exit(1) to terminate program.
	Fatalf(format string, args ...any)

	// Panic logs a message at the panic level and then calls panic().
	// Whether the log processes or not, always call panic() to terminate program.
	Panic(args ...any)

	// Panicf logs a formatted message at the panic level and then calls panic().
	// Whether the log processes or not, always call panic() to terminate program.
	Panicf(format string, args ...any)

	// WithFields returns a new logger with the given fields added to the context.
	// These fields append to end of log. (After the message)
	WithFields(fields map[string]any) Logger
	// Change minimum acceptable log level.
	ChangeLogLevel(minLogLevel LogLevel)
}

type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Warn
	Error
	Fatal
	Panic
	None
)

// A simple logger
type SLogger struct {
	// Minimum log level that processes and lower levels will be ignored.
	minLevel LogLevel
	// Key-value pairs to add to logs
	fields map[string]any
	mu     sync.Mutex
	writer io.Writer
}

// Create new simple logger instance.
// writer is the place the logs will write there.
func NewSLogger(minLogLevel LogLevel, fields map[string]any, writer io.Writer) Logger {
	return &SLogger{
		minLevel: minLogLevel,
		fields:   make(map[string]any),
		writer:   writer,
	}
}

func (l *SLogger) Debug(args ...any) {
	if l.isAcceptableLogLevel(Debug) {
		l.log("DEBUG", args...)
	}
}

func (l *SLogger) Debugf(format string, args ...any) {
	if l.isAcceptableLogLevel(Debug) {
		l.logf("DEBUG", format, args...)
	}
}

func (l *SLogger) Info(args ...any) {
	if l.isAcceptableLogLevel(Info) {
		l.log("INFO", args...)
	}
}

func (l *SLogger) Infof(format string, args ...any) {
	if l.isAcceptableLogLevel(Info) {
		l.logf("INFO", format, args...)
	}
}

func (l *SLogger) Warn(args ...any) {
	if l.isAcceptableLogLevel(Warn) {
		l.log("WARN", args...)
	}
}

func (l *SLogger) Warnf(format string, args ...any) {
	if l.isAcceptableLogLevel(Warn) {
		l.logf("WARN", format, args...)
	}
}

func (l *SLogger) Error(args ...any) {
	if l.isAcceptableLogLevel(Error) {
		l.log("ERROR", args...)
	}
}

func (l *SLogger) Errorf(format string, args ...any) {
	if l.isAcceptableLogLevel(Error) {
		l.logf("ERROR", format, args...)
	}
}

func (l *SLogger) Fatal(args ...any) {
	if l.isAcceptableLogLevel(Fatal) {
		l.log("FATAL", args...)
	}
	os.Exit(1)
}

func (l *SLogger) Fatalf(format string, args ...any) {
	if l.isAcceptableLogLevel(Fatal) {
		l.logf("FATAL", format, args...)
	}
	os.Exit(1)
}

func (l *SLogger) Panic(args ...any) {
	if l.isAcceptableLogLevel(Panic) {
		l.log("PANIC", args...)
	}
	panic(fmt.Sprint(args...))
}

func (l *SLogger) Panicf(format string, args ...any) {
	if l.isAcceptableLogLevel(Panic) {
		l.logf("PANIC", format, args...)
	}
	panic(fmt.Sprintf(format, args...))
}

func (l *SLogger) WithFields(fields map[string]any) Logger {
	newFields := make(map[string]any)
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}
	return &SLogger{minLevel: l.minLevel, fields: newFields}
}

func (l *SLogger) log(level string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	message := fmt.Sprint(args...)
	l.printLog(level, message)
}

func (l *SLogger) logf(level, format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	message := fmt.Sprintf(format, args...)
	l.printLog(level, message)
}

func (l *SLogger) printLog(level, message string) {
	if level == "ERROR" || level == "FATAL" || level == "PANIC" {
		fmt.Fprintf(l.writer, "[%s] %s", level, message)
	} else {
		fmt.Fprintf(l.writer, "[%s] %s", level, message)
	}

	for k, v := range l.fields {
		fmt.Fprintf(l.writer, " %s=%v", k, v)
	}
	fmt.Fprintln(l.writer)
}

func (l *SLogger) isAcceptableLogLevel(logLevel LogLevel) bool {
	return l.minLevel <= logLevel
}

func (l *SLogger) ChangeLogLevel(minLogLevel LogLevel) {
	l.minLevel = minLogLevel
}
