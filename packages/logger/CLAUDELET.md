# logger

## Purpose
Centralized logging utilities using HashiCorp hclog with optional OpenTelemetry OTLP gRPC log export. Supports dual output: human-readable format for stdout/file and JSON format for OTLP export. Provides both package-level convenience functions and interface-based loggers for dependency injection and testing. When telemetry is configured, logs are exported to OTLP collectors with automatic trace correlation.

## Key Files
- `logger.go` (188 lines) - Logger interface with dual-logger adapter (human-readable + JSON), initialization, and convenience functions

## Main Exports

### Initialization
- `InitLogger(logDirPath, logLevelController string) error`: Initializes logger with file and stdout output. Creates `service-seed.log` in specified directory. Log levels: "debug", "info", "warn", "error", "fatal".
- `InitLoggerWithOptions(logDirPath, logLevelController string, opts *LoggerOptions) error`: Extended initialization supporting OTLP export via LoggerOptions.
- `CloseLogger()`: Closes log file handle (should be deferred after InitLogger).

### Configuration Types
- `LoggerOptions`: Logger initialization options
  - `ExtraWriter io.Writer`: OTLP adapter destination for JSON-formatted logs

### Package-Level Functions
- `Debug(msg string, args ...interface{})`: Debug-level logging
- `Info(msg string, args ...interface{})`: Info-level logging
- `Warn(msg string, args ...interface{})`: Warning-level logging
- `Error(msg string, args ...interface{})`: Error-level logging
- `Fatal(msg string, args ...interface{})`: Logs error with "FATAL:" prefix and exits with code 1

### Interface-Based Logging
- `Logger` interface: Abstracts logging operations for testing and dependency injection
  - Methods: `Debug`, `Info`, `Warn`, `Error`, `Named(name string) Logger`
- `GetLogger() Logger`: Returns root logger wrapped in Logger interface
- `NewLogger(name string) Logger`: Creates named logger instance (e.g., "api", "stats")

## Implementation Details

### Core Logger Architecture
- **Dual-Logger Pattern**: Maintains two separate hclog instances:
  - `logger`: Human-readable format for stdout and file output
  - `jsonLogger`: JSON format for OTLP export (nil if OTLP not configured)
- `hclogAdapter`: Wrapper implementing Logger interface that writes to both loggers
- All package-level functions (Debug, Info, etc.) write to both loggers if JSON logger exists
- Interface methods (via Named) preserve dual-logger behavior in child loggers
- All logs prefixed with "SERVICE-SEED" name
- Default log level: Info (if invalid level specified)

## Interactions
- Used by all packages for logging (cli, api, bootstrap, config, stats)
- OTLP exporter exports to same endpoint as metrics via gRPC (if configured)

## Configuration/Dependencies
- Uses HashiCorp `go-hclog` (https://github.com/hashicorp/go-hclog)
- Log level controlled by `logLevelController` string argument
- Log file path: `{logDirPath}/service-seed.log`
- OTLP endpoint configured via telemetry block in config.hcl (when implemented)

## Example Usage

### Standard Logging
```go
import "github.com/organization/service-seed/packages/logger"

// Initialize at startup
logger.InitLogger("/var/log/service-seed", "info")
defer logger.CloseLogger()

// Package-level functions (writes to stdout + file)
logger.Info("Server started", "port", 3001)
logger.Error("Failed to connect", "error", err)

// Interface-based (for dependency injection)
log := logger.NewLogger("api")
log.Info("Request received", "method", "GET", "path", "/health")
```

### With OTLP Export (Future)
```go
// When OTLP support is added to this package:
// Initialize OTLP exporter
otlpWriter, err := logger.InitOTLPLogs(
    "localhost:4317",
    "service-seed",
    "1.0.0",
    "production",
)
if err != nil {
    return err
}
defer logger.ShutdownOTLP(context.Background())

// Initialize logger with OTLP export
// Human-readable goes to stdout/file, JSON goes to OTLP
err = logger.InitLoggerWithOptions("/var/log/service-seed", "info", &logger.LoggerOptions{
    ExtraWriter: otlpWriter,
})
if err != nil {
    return err
}
defer logger.CloseLogger()

// Logs now sent to both destinations automatically
logger.Info("Processing request", "request_id", id, "status", "success")
```

---
Handles logging logic for the application. Thread-safe for concurrent use across goroutines. Dual-logger architecture ensures human-readable logs for operators while maintaining structured JSON logs for observability platforms.
