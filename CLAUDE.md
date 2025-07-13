# DevHub - AI Assistant Instructions

This document contains instructions for AI assistants working with the DevHub codebase.

## Project Overview

DevHub is a development environment orchestrator written in Go that provides:
- **TUI Application**: Terminal user interface built with Bubble Tea for managing development services
- **Service Management**: Start, stop, restart, and monitor development services (databases, servers, build tools)
- **MCP Server**: Model Context Protocol server that exposes DevHub functionality to AI assistants
- **Health Monitoring**: TCP, HTTP, and command-based health checks for services
- **Log Management**: Real-time log viewing, searching, and following
- **Dependency Resolution**: Automatic service startup ordering based on dependencies

## Architecture

### Core Components

1. **DevHub (`devhub.go`)**: Main application orchestrating the TUI and services
2. **ServiceManager (`service_manager.go`)**: Manages service lifecycle and state
3. **MCPServer (`mcp_server.go`)**: Exposes MCP tools for AI assistant integration
4. **Services Package (`services/`)**: Service abstractions and implementations
5. **StateManager (`state_manager.go`)**: Handles async state updates and monitoring
6. **DependencyResolver (`dependency_resolver.go`)**: Resolves service dependencies

### Key Files

- `cmd/devhub/main.go`: CLI entry point
- `devhub.yml`: Configuration file format (see README.md for examples)
- `services/service.go`: Core service interfaces and BaseService implementation
- `view.go`: TUI rendering logic
- `types.go`: Data structures and type definitions

## Code Style Guidelines

- **Error Handling**: Use explicit error returns, avoid panics
- **Concurrency**: Use context.Context for cancellation, sync.RWMutex for thread safety
- **Logging**: Log to service buffers, not directly to stdout/stderr
- **Interface Design**: Small, focused interfaces following Go conventions
- **Testing**: Prefer integration tests over unit tests for service management

## Common Development Tasks

### Adding New MCP Tools

1. Define the tool in `mcp_server.go` using `mcp.NewTool()`
2. Implement the handler function following the pattern `handleToolName`
3. Register the tool with `ms.mcpServer.AddTool()`
4. Update `CLAUDE.md` documentation with tool description

### Adding New Health Check Types

1. Implement the `HealthChecker` interface in `services/health_check.go`
2. Add parsing logic in `service_manager.go` `createHealthChecker()` function
3. Update configuration documentation in README.md

### Service Implementation

Services must implement the `Service` interface:
- `ServiceInfo`: Name, Description, Port
- `ServiceLifecycle`: Start, Stop methods  
- `ServiceStatusProvider`: Status, health, error information
- `ServiceLogger`: Log management

Use `BaseService` as the foundation and `CmdService` for command-based services.

## MCP Integration

DevHub exposes development tools via Model Context Protocol:

### Available Tools

1. **list_services**: Get status of all configured services
2. **get_logs**: Retrieve logs with offset support for pagination
3. **service_control**: Start, stop, restart services
4. **health_check**: Get detailed health and resource usage
5. **search_logs**: Search for patterns in service logs with context

### Usage Patterns

- **Debugging**: Use `search_logs` to find errors, then `get_logs` for context
- **Service Management**: Check status with `list_services`, control with `service_control`
- **Health Monitoring**: Use `health_check` for detailed service information

## Configuration Format

DevHub uses YAML configuration with these key sections:

```yaml
service_name:
  desc: "Human readable description"
  port: 1234
  run: "command to execute"
  needs: ["dependency1", "dependency2"]
  health:
    tcp: 1234 | http: "url" | cmd: "command"
    interval: "5s"
    timeout: "3s"
    wait: "10s"
    retries: 3
  os:
    windows: "windows-specific command"
    darwin: "macOS-specific command"
    linux: "linux-specific command"
```

## Dependencies and External Libraries

- **Bubble Tea**: TUI framework (`github.com/charmbracelet/bubbletea`)
- **Lip Gloss**: Terminal styling (`github.com/charmbracelet/lipgloss`)
- **MCP-Go**: Model Context Protocol (`github.com/mark3labs/mcp-go`)
- **gopsutil**: System monitoring (`github.com/shirou/gopsutil/v4`)
- **YAML**: Configuration parsing (`gopkg.in/yaml.v3`)

## Testing Guidelines

- **Integration Tests**: Test service lifecycle, dependency resolution
- **Mock Services**: Use in-memory services for testing complex scenarios
- **Concurrency**: Test service state changes under concurrent access
- **Error Conditions**: Test service failures, network issues, invalid configurations

## Performance Considerations

- **Log Buffering**: Use `CircularLogBuffer` to prevent memory growth
- **Resource Monitoring**: Sample CPU/memory usage efficiently
- **State Updates**: Batch updates to prevent UI flickering
- **Context Cancellation**: Ensure proper cleanup on shutdown

## Security Notes

- **Command Execution**: Services run with DevHub's privileges
- **MCP Server**: Binds to localhost only by default
- **Log Content**: May contain sensitive information, handle appropriately
- **File Access**: Configuration files should be validated

## Common Debugging Steps

1. **Service Won't Start**: Check dependencies, command path, working directory
2. **Health Check Failing**: Verify health check configuration, timeouts
3. **High Resource Usage**: Check for log buffer growth, monitoring frequency
4. **MCP Connection Issues**: Verify port availability, localhost binding
5. **UI Rendering Issues**: Check terminal size, Lip Gloss styling

## Future Improvements

Consider these areas for enhancement:
- **Configuration Validation**: Schema validation for devhub.yml
- **Plugin System**: External service implementations
- **Metrics Export**: Prometheus/OpenTelemetry integration
- **Remote Services**: SSH/Docker service management
- **Service Templates**: Predefined service configurations