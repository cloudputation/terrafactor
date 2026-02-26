package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/go-hclog"
)

// Logger interface abstracts logging operations
// This allows us to decouple from hclog and makes testing easier
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Named(name string) Logger
}

// hclogAdapter wraps hclog.Logger to implement our Logger interface
// It maintains two loggers: one for human-readable output (console/file)
// and optionally one for JSON output (OTLP)
type hclogAdapter struct {
	logger     hclog.Logger
	jsonLogger hclog.Logger // nil if OTLP not configured
}

func (h *hclogAdapter) Debug(msg string, args ...interface{}) {
	h.logger.Debug(msg, args...)
	if h.jsonLogger != nil {
		h.jsonLogger.Debug(msg, args...)
	}
}

func (h *hclogAdapter) Info(msg string, args ...interface{}) {
	h.logger.Info(msg, args...)
	if h.jsonLogger != nil {
		h.jsonLogger.Info(msg, args...)
	}
}

func (h *hclogAdapter) Warn(msg string, args ...interface{}) {
	h.logger.Warn(msg, args...)
	if h.jsonLogger != nil {
		h.jsonLogger.Warn(msg, args...)
	}
}

func (h *hclogAdapter) Error(msg string, args ...interface{}) {
	h.logger.Error(msg, args...)
	if h.jsonLogger != nil {
		h.jsonLogger.Error(msg, args...)
	}
}

func (h *hclogAdapter) Named(name string) Logger {
	var namedJson hclog.Logger
	if h.jsonLogger != nil {
		namedJson = h.jsonLogger.Named(name)
	}
	return &hclogAdapter{
		logger:     h.logger.Named(name),
		jsonLogger: namedJson,
	}
}

var logger hclog.Logger     // Human-readable for stdout/file
var jsonLogger hclog.Logger // JSON for OTLP (nil if OTLP disabled)
var logFile *os.File
var logLevel hclog.Level

// LoggerOptions allows customizing logger initialization
type LoggerOptions struct {
	// ExtraWriter is an additional io.Writer to send logs to (e.g., OTLP adapter)
	// If set, a separate JSON-formatted logger will write to this destination
	ExtraWriter io.Writer
}

func InitLogger(logDirPath, logLevelController string) error {
	return InitLoggerWithOptions(logDirPath, logLevelController, nil)
}

func InitLoggerWithOptions(logDirPath, logLevelController string, opts *LoggerOptions) error {
	logFileName := "terrafactor.log"
	logFilePath := logDirPath + "/" + logFileName

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("Failed to open log file at path %s: %v", logFilePath, err)
	}

	switch logLevelController {
	case "debug":
		logLevel = hclog.Debug
	case "info":
		logLevel = hclog.Info
	case "warn":
		logLevel = hclog.Warn
	case "error":
		logLevel = hclog.Error
	case "fatal":
		logLevel = hclog.Error
	default:
		logLevel = hclog.Info
	}

	// Human-readable logger for stdout + file
	consoleWriter := io.MultiWriter(os.Stdout, logFile)
	logger = hclog.New(&hclog.LoggerOptions{
		Name:   "TERRAFACTOR",
		Level:  logLevel,
		Output: consoleWriter,
		// JSONFormat: false (default, human-readable)
	})

	// JSON logger for OTLP only (if configured)
	if opts != nil && opts.ExtraWriter != nil {
		jsonLogger = hclog.New(&hclog.LoggerOptions{
			Name:       "SERVICE-SEED",
			Level:      logLevel,
			Output:     opts.ExtraWriter,
			JSONFormat: true,
		})
	}

	return nil
}

func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

func Debug(msg string, args ...interface{}) {
	logger.Debug(msg, args...)
	if jsonLogger != nil {
		jsonLogger.Debug(msg, args...)
	}
}

func Info(msg string, args ...interface{}) {
	logger.Info(msg, args...)
	if jsonLogger != nil {
		jsonLogger.Info(msg, args...)
	}
}

func Warn(msg string, args ...interface{}) {
	logger.Warn(msg, args...)
	if jsonLogger != nil {
		jsonLogger.Warn(msg, args...)
	}
}

func Error(msg string, args ...interface{}) {
	logger.Error(msg, args...)
	if jsonLogger != nil {
		jsonLogger.Error(msg, args...)
	}
}

func Fatal(msg string, args ...interface{}) {
	logger.Error("FATAL: "+msg, args...)
	if jsonLogger != nil {
		jsonLogger.Error("FATAL: "+msg, args...)
	}
	os.Exit(1)
}

// GetLogger returns the root logger wrapped in our Logger interface
func GetLogger() Logger {
	return &hclogAdapter{logger: logger, jsonLogger: jsonLogger}
}

// NewLogger creates a named logger instance
func NewLogger(name string) Logger {
	var namedJson hclog.Logger
	if jsonLogger != nil {
		namedJson = jsonLogger.Named(name)
	}
	return &hclogAdapter{
		logger:     logger.Named(name),
		jsonLogger: namedJson,
	}
}
