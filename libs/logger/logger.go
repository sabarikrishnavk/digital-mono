package logger

import "log" // Standard log, can be replaced with a structured logger like zerolog, zap

// Logger defines a common interface for logging.
type Logger interface {
	Info(message string, fields ...interface{})
	Error(err error, message string, fields ...interface{})
	// Add other levels like Debug, Warn as needed
}

// stdLogger is a simple implementation of Logger using the standard log package.
type stdLogger struct{}

// NewStdLogger creates a new Logger that uses the standard log package.
// This was the intended constructor name from previous steps.
func NewStdLogger() Logger {
	return &stdLogger{}
}

// Info logs an info message.
func (l *stdLogger) Info(message string, fields ...interface{}) {
	// Basic formatting, a real structured logger would handle fields better
	log.Printf("INFO: %s %v\n", message, fields)
}

// Error logs an error message.
func (l *stdLogger) Error(err error, message string, fields ...interface{}) {
	log.Printf("ERROR: %s: %v %v\n", message, err, fields)
}