package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var logLevelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

var logLevelColors = map[LogLevel]string{
	DEBUG: "\033[36m", // Cyan
	INFO:  "\033[32m", // Green
	WARN:  "\033[33m", // Yellow
	ERROR: "\033[31m", // Red
	FATAL: "\033[35m", // Magenta
}

const colorReset = "\033[0m"

// Logger provides structured logging functionality
type Logger struct {
	level      LogLevel
	colored    bool
	component  string
	output     *log.Logger
	errorCount int64
	warnCount  int64
}

// Config contains logger configuration
type Config struct {
	Level     LogLevel
	Colored   bool
	Component string
	Output    *os.File
}

// NewLogger creates a new logger instance
func NewLogger(config Config) *Logger {
	if config.Output == nil {
		config.Output = os.Stderr
	}

	return &Logger{
		level:     config.Level,
		colored:   config.Colored,
		component: config.Component,
		output:    log.New(config.Output, "", 0), // We'll format timestamps ourselves
	}
}

// NewDefaultLogger creates a logger with default settings
func NewDefaultLogger(component string) *Logger {
	return NewLogger(Config{
		Level:     INFO,
		Colored:   true,
		Component: component,
		Output:    os.Stderr,
	})
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.log(DEBUG, msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.log(INFO, msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.warnCount++
	l.log(WARN, msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	l.errorCount++
	l.log(ERROR, msg, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.log(FATAL, msg, args...)
	os.Exit(1)
}

// ErrorWithStack logs an error with stack trace
func (l *Logger) ErrorWithStack(err error, msg string, args ...interface{}) {
	if err == nil {
		return
	}

	l.errorCount++
	
	// Format the message
	formattedMsg := fmt.Sprintf(msg, args...)
	if formattedMsg != "" {
		formattedMsg += ": "
	}
	formattedMsg += err.Error()

	// Get stack trace
	stack := l.getStackTrace(2) // Skip this function and the caller
	
	l.logWithStack(ERROR, formattedMsg, stack)
}

// WarnError logs an error as a warning (non-fatal)
func (l *Logger) WarnError(err error, msg string, args ...interface{}) {
	if err == nil {
		return
	}

	l.warnCount++
	
	formattedMsg := fmt.Sprintf(msg, args...)
	if formattedMsg != "" {
		formattedMsg += ": "
	}
	formattedMsg += err.Error()

	l.log(WARN, formattedMsg)
}

// LogError is a helper that logs an error if it's not nil
func (l *Logger) LogError(err error, msg string, args ...interface{}) {
	if err != nil {
		l.Error(msg+": %v", append(args, err)...)
	}
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// SetColored enables or disables colored output
func (l *Logger) SetColored(colored bool) {
	l.colored = colored
}

// GetErrorCount returns the number of errors logged
func (l *Logger) GetErrorCount() int64 {
	return l.errorCount
}

// GetWarnCount returns the number of warnings logged
func (l *Logger) GetWarnCount() int64 {
	return l.warnCount
}

// HasErrors returns true if any errors have been logged
func (l *Logger) HasErrors() bool {
	return l.errorCount > 0
}

// HasWarnings returns true if any warnings have been logged
func (l *Logger) HasWarnings() bool {
	return l.warnCount > 0
}

// Child creates a child logger with additional component context
func (l *Logger) Child(component string) *Logger {
	childComponent := l.component
	if childComponent != "" {
		childComponent += "." + component
	} else {
		childComponent = component
	}

	return &Logger{
		level:     l.level,
		colored:   l.colored,
		component: childComponent,
		output:    l.output,
	}
}

// Private methods

func (l *Logger) log(level LogLevel, msg string, args ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := logLevelNames[level]
	
	// Add color if enabled
	if l.colored {
		levelStr = logLevelColors[level] + levelStr + colorReset
	}

	// Format the message
	formattedMsg := fmt.Sprintf(msg, args...)

	// Build the log line
	var logLine string
	if l.component != "" {
		logLine = fmt.Sprintf("%s [%s] [%s] %s", timestamp, levelStr, l.component, formattedMsg)
	} else {
		logLine = fmt.Sprintf("%s [%s] %s", timestamp, levelStr, formattedMsg)
	}

	l.output.Println(logLine)
}

func (l *Logger) logWithStack(level LogLevel, msg string, stack []string) {
	if level < l.level {
		return
	}

	// Log the main message
	l.log(level, msg)

	// Log stack trace if debug level
	if l.level <= DEBUG {
		for _, frame := range stack {
			l.log(level, "  %s", frame)
		}
	}
}

func (l *Logger) getStackTrace(skip int) []string {
	var stack []string
	
	for i := skip; i < skip+10; i++ { // Limit to 10 frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		var name string
		if fn != nil {
			name = fn.Name()
			// Simplify function name
			if idx := strings.LastIndex(name, "/"); idx != -1 {
				name = name[idx+1:]
			}
		} else {
			name = "unknown"
		}

		// Simplify file path
		if idx := strings.LastIndex(file, "/"); idx != -1 {
			file = file[idx+1:]
		}

		stack = append(stack, fmt.Sprintf("%s:%d in %s", file, line, name))
	}

	return stack
}

// Global logger instance
var defaultLogger = NewDefaultLogger("corynth")

// Package-level functions that use the default logger

// Debug logs a debug message using the default logger
func Debug(msg string, args ...interface{}) {
	defaultLogger.Debug(msg, args...)
}

// Info logs an info message using the default logger
func Info(msg string, args ...interface{}) {
	defaultLogger.Info(msg, args...)
}

// Warn logs a warning message using the default logger
func Warn(msg string, args ...interface{}) {
	defaultLogger.Warn(msg, args...)
}

// Error logs an error message using the default logger
func Error(msg string, args ...interface{}) {
	defaultLogger.Error(msg, args...)
}

// Fatal logs a fatal message and exits using the default logger
func Fatal(msg string, args ...interface{}) {
	defaultLogger.Fatal(msg, args...)
}

// ErrorWithStack logs an error with stack trace using the default logger
func ErrorWithStack(err error, msg string, args ...interface{}) {
	defaultLogger.ErrorWithStack(err, msg, args...)
}

// WarnError logs an error as a warning using the default logger
func WarnError(err error, msg string, args ...interface{}) {
	defaultLogger.WarnError(err, msg, args...)
}

// LogError logs an error if it's not nil using the default logger
func LogError(err error, msg string, args ...interface{}) {
	defaultLogger.LogError(err, msg, args...)
}

// SetLevel sets the global logging level
func SetLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

// SetColored enables or disables colored output globally
func SetColored(colored bool) {
	defaultLogger.SetColored(colored)
}

// GetDefaultLogger returns the default logger
func GetDefaultLogger() *Logger {
	return defaultLogger
}

// NewFileLogger creates a logger that writes to a file
func NewFileLogger(filename, component string, level LogLevel) (*Logger, error) {
	// Ensure log directory exists
	logDir := filepath.Dir(filename)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return NewLogger(Config{
		Level:     level,
		Colored:   false, // Don't use colors in file output
		Component: component,
		Output:    file,
	}), nil
}

// Helper functions for converting from string

// ParseLogLevel parses a log level from string
func ParseLogLevel(level string) (LogLevel, error) {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG, nil
	case "INFO":
		return INFO, nil
	case "WARN", "WARNING":
		return WARN, nil
	case "ERROR":
		return ERROR, nil
	case "FATAL":
		return FATAL, nil
	default:
		return INFO, fmt.Errorf("invalid log level: %s", level)
	}
}