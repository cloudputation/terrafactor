# cli

## Purpose
Defines the application's command-line interface (CLI) using Cobra. Provides the `agent` command to start the service with filesystem initialization.

## Key Files
- `cli.go` (37 lines) - Root command setup and agent command definition

## Main Exports
- `SetupRootCommand() *cobra.Command`: Returns the root Cobra command with registered subcommands.

## Available Commands
- `agent` - Bootstraps the filesystem and starts the HTTP server with all registered endpoints (health checks, metrics)

## Interactions
- **bootstrap**: Initializes application filesystem (data directories) before server start
- **api**: Starts the HTTP server with all registered endpoints
- **logger**: Provides Fatal-level logging for bootstrap failures
- **stats**: Initializes metrics tracking (when implemented)

## Configuration/Dependencies
- Uses Cobra for CLI parsing and command management
- Command execution order: bootstrap filesystem → start API server
- Easily extensible for additional commands

## Example Usage
```go
import "github.com/organization/service-seed/packages/cli"

rootCmd := cli.SetupRootCommand()
rootCmd.Execute()
```

## Adding New Commands
Create a new `cobra.Command` and register it with `rootCmd.AddCommand()`. Follow the `agent` command pattern for consistency.

Example:
```go
var cmdValidate = &cobra.Command{
    Use:   "validate",
    Short: "Validate configuration file",
    Run: func(cmd *cobra.Command, args []string) {
        err := config.LoadConfiguration()
        if err != nil {
            l.Fatal("Configuration validation failed: %v", err)
        }
        l.Info("Configuration is valid")
    },
}

rootCmd.AddCommand(cmdValidate)
```

## Future Enhancements

Consider adding:
- **Config Validation Command**: Validate config.hcl without starting service
- **Version Command**: Display application version and build info
- **Status Command**: Check service status and health
- **Migrate Command**: Database or data migrations
- **Test Command**: Run integration or smoke tests

---
Focuses on CLI orchestration and command definitions. Provides clean separation between command parsing and business logic.
