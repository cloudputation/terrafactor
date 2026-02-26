# bootstrap

## Purpose
Initializes the application filesystem by creating required data directories at startup. Ensures the data directory structure is in place before the application begins processing requests and managing state.

## Key Files
- `bootstrap.go` (29 lines): Single-file package containing filesystem initialization logic

## Exports
- `BootstrapFileSystem() error`: Creates data directory structure based on configuration. Returns error if directory creation fails.

## Dependencies
- `config`: Reads `AppConfig.DataDir` for directory path and `RootDir` for base directory location
- `logger`: Logs initialization progress and errors

## Implementation Details

**Directory Creation**:
- Uses `AppConfig.DataDir` and `RootDir` to construct absolute path
- Creates directory with `0755` permissions (rwxr-xr-x)
- Uses `os.MkdirAll` to create parent directories if needed

**Error Handling**:
- Logs and returns wrapped error if `MkdirAll` fails
- Continues application startup only if directory creation succeeds

**Logging**:
- Logs configuration file path for traceability
- Logs data directory path upon successful creation
- Uses structured logging with key-value pairs

## Example Usage
```go
import "github.com/organization/service-seed/packages/bootstrap"

// Called during application initialization (before API starts)
if err := bootstrap.BootstrapFileSystem(); err != nil {
    log.Fatal("Failed to bootstrap filesystem", "error", err)
}
```

## Integration Points
- **Called by**: `cli` package during `agent` command startup
- **Must run before**: API server initialization (may need data directory for state)
- **Configuration dependency**: Requires `config.LoadConfiguration()` to have been called first

## Future Enhancements

Consider adding:
- **Version Management**: Load and track API version from file (see sentinel/bootstrap/bootstrap.go)
- **Database Initialization**: Initialize embedded databases (SQLite, BoltDB)
- **State Recovery**: Load previous state from disk on restart
- **Directory Validation**: Verify directory permissions and writability
- **Migration Support**: Handle data directory migrations between versions
- **Health Checks**: Verify filesystem health and available disk space

---
Single-responsibility package focused on filesystem initialization. No runtime logic beyond startup initialization.
