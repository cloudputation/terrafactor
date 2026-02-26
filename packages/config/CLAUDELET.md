# Config - HCL Configuration Management

## Purpose

Centralize application configuration loading, parsing, and access for all modules. Provides single source of truth for all configuration with intelligent defaults, including OpenTelemetry telemetry export. Configuration is split into logical units (server, telemetry) for maintainability. Supports runtime log level configuration.

## Core Functionality

- **HCL Parsing**: Parse HashiCorp Configuration Language files
- **Default Application**: Apply sensible defaults for all optional fields via modular default functions
- **Environment Variables**: Support env var overrides (SS_CONFIG_FILE_PATH)
- **Validation**: Ensure required fields present, validate field types
- **Telemetry Export**: OpenTelemetry OTLP gRPC export configuration with TLS support and signal-specific settings
- **Modular Structure**: Configuration split into config.go and telemetry.go for logical separation

## Configuration Structure

```go
type Configuration struct {
    LogDir    string
    DataDir   string
    Server    Server          // Defined in config.go
    Telemetry *Telemetry      // Defined in telemetry.go
}
```

**Server Configuration** (config.go):
```go
type Server struct {
    ServerPort    string
    ServerAddress string
}
```

**Telemetry Configuration** (telemetry.go):
```go
type Telemetry struct {
    // Shared config (inherited by metrics, logs, traces)
    Endpoint string
    TLS      *OTLPTLSConfig
    Headers  map[string]string

    // Signal-specific config
    Metrics *OTLPMetricsConfig
    Logs    *OTLPLogsConfig
    Traces  *OTLPTracesConfig
}

type OTLPMetricsConfig struct {
    Enabled         bool
    Endpoint        string              // Overrides shared endpoint
    Protocol        string              // "grpc" or "http"
    IntervalSeconds int
}

type OTLPLogsConfig struct {
    Enabled  bool
    Endpoint string              // Inherits from shared if empty
}

type OTLPTracesConfig struct {
    Enabled      bool
    Endpoint     string              // Inherits from shared if empty
    SamplingRate float64             // 0.0-1.0 (default: 1.0)
}

type OTLPTLSConfig struct {
    Enabled  bool
    Insecure bool
    CAFile   string
    CertFile string
    KeyFile  string
}
```

## Key Files

- **config.go** (79 lines) - HCL parsing, struct definitions, configuration loading, modular defaults application
- **telemetry.go** - OpenTelemetry OTLP export configuration with signal-specific settings and inheritance

## Exports

**Main Functions**:
- `LoadConfiguration() error` - Parse HCL, apply defaults
- `GetConfigPath() string` - Return config file path from env or default
- `applyDefaults()` - Delegate to modular default functions
- `applyTelemetryDefaults()` - Apply telemetry-specific defaults (protocol, interval, signal inheritance)

**Global Variables**:
- `AppConfig Configuration` - Loaded configuration (singleton)
- `ConfigPath string` - Resolved config file path
- `RootDir string` - Application root directory

**Constants**:
- `MaxWorkers = 10` - Worker pool size limit

## Configuration Loading Flow

```go
LoadConfiguration():
  1. Read SS_CONFIG_FILE_PATH env var (or use default)
  2. Parse HCL file with hclparse
  3. Decode into Configuration struct with gohcl
  4. applyDefaults() - delegates to modular functions:
     - applyTelemetryDefaults()
  5. Validate required fields
  6. Set global AppConfig variable
```

## Configuration Blocks

### Root Block

```hcl
log_dir = "/var/log/service-seed"
data_dir = "/var/lib/service-seed"
```

### Server Block

```hcl
server {
  port = "3001"
  address = "0.0.0.0"
}
```

### Telemetry Block

```hcl
telemetry {
  # Shared configuration (inherited by all signals)
  endpoint = "localhost:4317"
  headers = {
    "X-API-Key" = "secret-key"
  }

  # Optional shared TLS configuration
  tls {
    enabled = true
    insecure = false
    ca_file = "/path/to/ca.crt"
    cert_file = "/path/to/client.crt"
    key_file = "/path/to/client.key"
  }

  # Metrics export configuration
  metrics {
    enabled = true
    # endpoint inherits from shared if not specified
    protocol = "grpc"              # "grpc" or "http"
    interval_seconds = 60
  }

  # Logs export configuration
  logs {
    enabled = true
    # endpoint, tls, headers inherit from shared if not specified
  }

  # Traces export configuration
  traces {
    enabled = true
    # endpoint, tls, headers inherit from shared if not specified
    sampling_rate = 1.0           # 0.0-1.0 (default: 1.0 = 100%)
  }
}
```

**Defaults**:
- `metrics.protocol`: "grpc"
- `metrics.interval_seconds`: 60
- `traces.sampling_rate`: 1.0
- `logs.endpoint`: Inherits from shared `endpoint` if empty
- `traces.endpoint`: Inherits from shared `endpoint` if empty

**Behavior**:
- Each signal (metrics, logs, traces) inherits shared configuration
- Signal-specific `endpoint` overrides shared `endpoint` if provided
- Logs and traces inherit TLS and headers from shared config
- Supports both gRPC and HTTP protocols for OTLP export
- Optional TLS with CA, client cert/key for mutual TLS
- Custom headers for authentication (API keys, tokens)

## applyDefaults() Behavior

Called automatically during `LoadConfiguration()`. Delegates to modular default-setting functions for each configuration area:

```go
func applyDefaults() {
    applyTelemetryDefaults() // Telemetry protocol, interval, signal inheritance
}
```

**Philosophy**: Application code should never check "if field == 0, use default". Config package handles all defaults centrally in modular functions.

## Environment Variable Support

**Config File Path**:
```bash
export SS_CONFIG_FILE_PATH=./config.hcl
```

No need to set in config file if env vars present.

## Dependencies

- **hashicorp/hcl/v2** - HCL parsing and decoding
- **spf13/viper** - Environment variable binding
- **logger** - Initialization logging

## Validation

**Required Fields**:
- `server.port`
- `server.address`

**Optional Fields**: Everything else (has defaults)

**Telemetry Validation**:
- Signal configs inherit from shared telemetry config if not specified
- Protocol must be "grpc" or "http" (defaults to "grpc")
- Sampling rate must be 0.0-1.0 (defaults to 1.0)

## Error Handling

- **File not found**: Returns error, prevents startup
- **Parse errors**: Returns error with file/line number from HCL diagnostics
- **Invalid types**: Returns error during gohcl decode
- **Missing required fields**: Returns error after parsing
- **Invalid telemetry protocol**: Caught during validation

## Usage Pattern

```go
import "github.com/organization/service-seed/packages/config"

func main() {
    // Load configuration
    err := config.LoadConfiguration()
    if err != nil {
        log.Fatal(err)
    }

    // Access configuration
    port := config.AppConfig.Server.ServerPort

    // Access telemetry config
    if config.AppConfig.Telemetry != nil {
        if config.AppConfig.Telemetry.Metrics != nil && config.AppConfig.Telemetry.Metrics.Enabled {
            // Initialize OTLP metrics exporter
        }
        if config.AppConfig.Telemetry.Logs != nil && config.AppConfig.Telemetry.Logs.Enabled {
            // Initialize OTLP logs exporter
        }
        if config.AppConfig.Telemetry.Traces != nil && config.AppConfig.Telemetry.Traces.Enabled {
            // Initialize OTLP traces exporter with sampling
        }
    }
}
```

## Configuration Precedence

1. **HCL file values**: Explicit values in config.hcl
2. **Environment variables**: For config file path (SS_CONFIG_FILE_PATH)
3. **Defaults**: Applied by applyDefaults() for missing optional fields
4. **Inheritance**: Signal configs inherit from shared telemetry config

**No fallback logic in application code** - single source of truth in config package.

## Thread Safety

- **Global AppConfig**: Loaded once at startup, read-only after initialization

## Design Decisions

1. **Modular Structure**: Configuration split into config.go and telemetry.go for maintainability and separation of concerns
2. **Modular Defaults**: Default application split into multiple functions (applyTelemetryDefaults, etc.) for clarity
3. **Single Load Point**: All configuration loaded via LoadConfiguration()
4. **Default Values**: Sensible defaults for all optional fields eliminate null checks throughout codebase
5. **Signal Inheritance**: Metrics, logs, and traces inherit shared telemetry config (endpoint, TLS, headers) for DRY principle
6. **Protocol Flexibility**: Support both gRPC and HTTP protocols for OTLP export
7. **TLS Support**: Optional mutual TLS for secure telemetry export to collectors
8. **Signal-Specific Settings**: Each signal (metrics, logs, traces) can override shared config with signal-specific settings
