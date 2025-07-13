package devhub

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type MCPServer struct {
	httpServer     *server.StreamableHTTPServer
	mcpServer      *server.MCPServer
	serviceManager *ServiceManager
	Port           int
}

func NewMCPServer(serviceManager *ServiceManager, port int) *MCPServer {
	mcpServer := server.NewMCPServer(
		"DevHub MCP Server",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	ms := &MCPServer{
		mcpServer:      mcpServer,
		serviceManager: serviceManager,
		Port:          port,
	}

	ms.registerTools()

	return ms
}

func (ms *MCPServer) registerTools() {
	// Register list_services tool
	listServicesTool := mcp.NewTool("list_services",
		mcp.WithDescription("List all configured services and their current status"),
	)

	ms.mcpServer.AddTool(listServicesTool, ms.handleListServices)

	// Register get_logs tool
	getLogsTool := mcp.NewTool("get_logs",
		mcp.WithDescription("Get logs for a specific service"),
		mcp.WithString("service",
			mcp.Required(),
			mcp.Description("Name of the service to get logs for"),
		),
		mcp.WithNumber("lines",
			mcp.Description("Number of log lines to retrieve (default: 50)"),
		),
		mcp.WithNumber("offset",
			mcp.Description("Number of lines to skip from the end (default: 0)"),
		),
	)

	ms.mcpServer.AddTool(getLogsTool, ms.handleGetLogs)

	// Register service control tools
	serviceControlTool := mcp.NewTool("service_control",
		mcp.WithDescription("Control service lifecycle (start/stop/restart)"),
		mcp.WithString("service",
			mcp.Required(),
			mcp.Description("Name of the service to control"),
		),
		mcp.WithString("action",
			mcp.Required(),
			mcp.Description("Action to perform on the service"),
			mcp.Enum("start", "stop", "restart"),
		),
	)

	ms.mcpServer.AddTool(serviceControlTool, ms.handleServiceControl)

	// Register health check tool
	healthCheckTool := mcp.NewTool("health_check",
		mcp.WithDescription("Check health status of a service"),
		mcp.WithString("service",
			mcp.Required(),
			mcp.Description("Name of the service to check health for"),
		),
	)

	ms.mcpServer.AddTool(healthCheckTool, ms.handleHealthCheck)

	// Register search_logs tool
	searchLogsTool := mcp.NewTool("search_logs",
		mcp.WithDescription("Search for a pattern in service logs"),
		mcp.WithString("service",
			mcp.Required(),
			mcp.Description("Name of the service to search logs in"),
		),
		mcp.WithString("pattern",
			mcp.Required(),
			mcp.Description("Pattern to search for (case-insensitive)"),
		),
		mcp.WithNumber("context_lines",
			mcp.Description("Number of context lines to show before and after matches (default: 2)"),
		),
		mcp.WithNumber("max_results",
			mcp.Description("Maximum number of matches to return (default: 50)"),
		),
	)

	ms.mcpServer.AddTool(searchLogsTool, ms.handleSearchLogs)
}

func (ms *MCPServer) handleListServices(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	services := ms.serviceManager.GetServices()
	
	result := "Services Status:\n\n"
	for _, service := range services {
		result += fmt.Sprintf("- %s (%s): %s", service.Name, service.Port, service.Status)
		if service.ErrorMsg != "" {
			result += fmt.Sprintf(" - Error: %s", service.ErrorMsg)
		}
		result += "\n"
	}

	return mcp.NewToolResultText(result), nil
}

func (ms *MCPServer) handleGetLogs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	serviceName, err := request.RequireString("service")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	lines := 50
	if linesArg, ok := request.GetArguments()["lines"]; ok {
		if l, ok := linesArg.(float64); ok {
			lines = int(l)
		}
	}

	offset := 0
	if offsetArg, ok := request.GetArguments()["offset"]; ok {
		if o, ok := offsetArg.(float64); ok {
			offset = int(o)
		}
	}

	service := ms.serviceManager.GetService(serviceName)
	if service == nil {
		return mcp.NewToolResultError(fmt.Sprintf("service '%s' not found", serviceName)), nil
	}

	logs := service.Logs()
	logString := string(logs)
	logLines := strings.Split(logString, "\n")

	// Calculate the range with offset
	totalLines := len(logLines)
	end := totalLines - offset
	if end < 0 {
		end = 0
	}
	
	start := end - lines
	if start < 0 {
		start = 0
	}
	
	if start >= totalLines {
		return mcp.NewToolResultText(fmt.Sprintf("No logs found for %s with offset %d", serviceName, offset)), nil
	}

	actualLines := end - start
	result := fmt.Sprintf("Logs for %s (showing %d lines, offset %d from end):\n\n", serviceName, actualLines, offset)
	for i := start; i < end && i < totalLines; i++ {
		line := logLines[i]
		if line != "" {
			result += line + "\n"
		}
	}

	return mcp.NewToolResultText(result), nil
}

func (ms *MCPServer) handleServiceControl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	serviceName, err := request.RequireString("service")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	action, err := request.RequireString("action")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var result string
	switch action {
	case "start":
		err = ms.serviceManager.StartService(serviceName)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to start service: %v", err)), nil
		}
		result = fmt.Sprintf("Service '%s' started successfully", serviceName)

	case "stop":
		err = ms.serviceManager.StopService(serviceName)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to stop service: %v", err)), nil
		}
		result = fmt.Sprintf("Service '%s' stopped successfully", serviceName)

	case "restart":
		// Find the service index
		services := ms.serviceManager.GetServices()
		serviceIndex := -1
		for i, s := range services {
			if s.Name == serviceName {
				serviceIndex = i
				break
			}
		}
		if serviceIndex == -1 {
			return mcp.NewToolResultError(fmt.Sprintf("service '%s' not found", serviceName)), nil
		}
		err = ms.serviceManager.RestartService(serviceIndex)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to restart service: %v", err)), nil
		}
		result = fmt.Sprintf("Service '%s' restarted successfully", serviceName)

	default:
		return mcp.NewToolResultError(fmt.Sprintf("unknown action: %s", action)), nil
	}

	return mcp.NewToolResultText(result), nil
}

func (ms *MCPServer) handleHealthCheck(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	serviceName, err := request.RequireString("service")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	service := ms.serviceManager.GetService(serviceName)
	if service == nil {
		return mcp.NewToolResultError(fmt.Sprintf("service '%s' not found", serviceName)), nil
	}

	// Get service info
	services := ms.serviceManager.GetServices()
	var serviceInfo *ServiceInfo
	for _, s := range services {
		if s.Name == serviceName {
			serviceInfo = &s
			break
		}
	}
	
	if serviceInfo == nil {
		return mcp.NewToolResultError(fmt.Sprintf("service info not found for '%s'", serviceName)), nil
	}

	result := fmt.Sprintf("Health Status for %s:\n", serviceName)
	result += fmt.Sprintf("- Status: %s\n", serviceInfo.Status)
	result += fmt.Sprintf("- Health: %s\n", serviceInfo.HealthStatus)
	
	if serviceInfo.StartTime != nil {
		uptime := time.Since(*serviceInfo.StartTime)
		result += fmt.Sprintf("- Uptime: %s\n", uptime.Round(time.Second))
	}
	
	result += fmt.Sprintf("- CPU: %.2f%%\n", serviceInfo.CPUPercent)
	result += fmt.Sprintf("- Memory: %.2f MB\n", serviceInfo.MemoryMB)
	
	if serviceInfo.ErrorMsg != "" {
		result += fmt.Sprintf("- Error: %s\n", serviceInfo.ErrorMsg)
	}

	return mcp.NewToolResultText(result), nil
}

func (ms *MCPServer) handleSearchLogs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	serviceName, err := request.RequireString("service")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	pattern, err := request.RequireString("pattern")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	contextLines := 2
	if contextArg, ok := request.GetArguments()["context_lines"]; ok {
		if c, ok := contextArg.(float64); ok {
			contextLines = int(c)
		}
	}

	maxResults := 50
	if maxArg, ok := request.GetArguments()["max_results"]; ok {
		if m, ok := maxArg.(float64); ok {
			maxResults = int(m)
		}
	}

	service := ms.serviceManager.GetService(serviceName)
	if service == nil {
		return mcp.NewToolResultError(fmt.Sprintf("service '%s' not found", serviceName)), nil
	}

	logs := service.Logs()
	logString := string(logs)
	logLines := strings.Split(logString, "\n")

	// Perform case-insensitive search
	lowerPattern := strings.ToLower(pattern)
	var matches []int
	for i, line := range logLines {
		if strings.Contains(strings.ToLower(line), lowerPattern) {
			matches = append(matches, i)
			if len(matches) >= maxResults {
				break
			}
		}
	}

	if len(matches) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No matches found for pattern '%s' in %s logs", pattern, serviceName)), nil
	}

	result := fmt.Sprintf("Found %d matches for '%s' in %s logs:\n\n", len(matches), pattern, serviceName)
	
	for _, matchIdx := range matches {
		result += fmt.Sprintf("--- Match at line %d ---\n", matchIdx+1)
		
		// Show context before
		start := matchIdx - contextLines
		if start < 0 {
			start = 0
		}
		
		// Show context after
		end := matchIdx + contextLines + 1
		if end > len(logLines) {
			end = len(logLines)
		}
		
		// Add lines with indicators
		for i := start; i < end; i++ {
			if i < len(logLines) {
				if i == matchIdx {
					result += fmt.Sprintf(">>> %s\n", logLines[i])
				} else {
					result += fmt.Sprintf("    %s\n", logLines[i])
				}
			}
		}
		result += "\n"
	}

	return mcp.NewToolResultText(result), nil
}

func (ms *MCPServer) Start() error {
	ms.httpServer = server.NewStreamableHTTPServer(ms.mcpServer)
	addr := fmt.Sprintf("localhost:%d", ms.Port)
	return ms.httpServer.Start(addr)
}

func (ms *MCPServer) Stop() error {
	if ms.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return ms.httpServer.Shutdown(ctx)
	}
	return nil
}