// Package logger provides centralized logging functionality with support
// for verbose/debug modes and different log levels.
package logger

import (
	"fmt"
	"log"
	"os"
)

var (
	verboseEnabled bool
	logger         *log.Logger
)

func init() {
	logger = log.New(os.Stderr, "[mcp-executor] ", log.LstdFlags)
}

// SetVerbose enables or disables verbose logging
func SetVerbose(enabled bool) {
	verboseEnabled = enabled
}

// IsVerbose returns whether verbose logging is enabled
func IsVerbose() bool {
	return verboseEnabled
}

// Verbose prints a message only if verbose mode is enabled
func Verbose(format string, args ...any) {
	if verboseEnabled {
		logger.Printf(format, args...)
	}
}

// Info prints an info message (always shown)
func Info(format string, args ...any) {
	logger.Printf("INFO: "+format, args...)
}

// Error prints an error message (always shown)
func Error(format string, args ...any) {
	logger.Printf("ERROR: "+format, args...)
}

// Debug prints a debug message only if verbose mode is enabled
func Debug(format string, args ...any) {
	if verboseEnabled {
		logger.Printf("DEBUG: "+format, args...)
	}
}

// VerbosePrint prints to stdout if verbose mode is enabled (for startup messages)
func VerbosePrint(format string, args ...any) {
	if verboseEnabled {
		fmt.Printf(format+"\n", args...)
	}
}
