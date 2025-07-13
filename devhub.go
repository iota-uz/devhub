package devhub

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NewDevHub creates a new DevHub instance with all configured services
func NewDevHub(configPath string) (*Model, error) {
	return NewDevHubWithMCPPort(configPath, 8765)
}

// NewDevHubWithMCPPort creates a new DevHub instance with all configured services and custom MCP port
func NewDevHubWithMCPPort(configPath string, mcpPort int) (*Model, error) {
	loader := NewFileConfigLoader(configPath)
	serviceManager, err := NewServiceManager(loader)
	if err != nil {
		return nil, err
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// Create state manager for async updates
	stateManager := NewStateManager(serviceManager)
	stateManager.Start()

	// Create MCP server
	mcpServer := NewMCPServer(serviceManager, mcpPort)

	return &Model{
		Services:       stateManager.GetServices(),
		ServiceManager: serviceManager,
		StateManager:   stateManager,
		MCPServer:      mcpServer,
		UI: UIState{
			SelectedIndex: 0,
			Spinner:       s,
			ViewMode:      ServiceView,
		},
		LogView: LogViewState{
			AutoScroll: true,
		},
		Search: SearchState{
			Active: false,
		},
		LogCache: NewLogCache(),
	}, nil
}

// Run starts the DevHub TUI interface
func (m *Model) Run(ctx context.Context) error {
	// Create cancellable context for the program
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start MCP server in background
	go func() {
		// Note: MCP server is available at http://localhost:<port>/mcp
		if err := m.MCPServer.Start(); err != nil {
			// Log error but don't fail the whole program
			// Since we're in a TUI context, we can't use regular logging
			// The error will be silently ignored to avoid disrupting the UI
		}
	}()

	// Handle context cancellation
	go func() {
		<-runCtx.Done()
		m.StateManager.Stop()
		m.ServiceManager.Shutdown()
		if err := m.MCPServer.Stop(); err != nil {
			// Log error but don't fail shutdown
			fmt.Fprintf(os.Stderr, "Error stopping MCP server: %v\n", err)
		}
	}()

	p := tea.NewProgram(*m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return err
	}

	m.StateManager.Stop()
	m.ServiceManager.Shutdown()
	if err := m.MCPServer.Stop(); err != nil {
		// Log error but don't fail shutdown
		fmt.Fprintf(os.Stderr, "Error stopping MCP server: %v\n", err)
	}
	return nil
}
