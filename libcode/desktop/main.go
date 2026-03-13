// Package main provides the desktop application entry point
package main

import (
	"context"
	"embed"

	"github.com/gemone/libcode/internal/ai"
	"github.com/gemone/libcode/internal/ai/provider"
	"github.com/gemone/libcode/internal/config"
	"github.com/gemone/libcode/internal/logger"
	"github.com/gemone/libcode/internal/storage"
	"github.com/gemone/libcode/internal/tool/framework"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"go.uber.org/zap"
)

//go:embed all:frontend/dist
var assets embed.FS

// App represents the desktop application
type App struct {
	ctx           context.Context
	cfg           *config.Config
	db            *storage.DB
	toolRegistry  *framework.Registry
	aiProviders   *provider.Registry
	currentAI     ai.Provider
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize config
	cfg, err := config.Load()
	if err != nil {
		logger.Get().Error("Failed to load config", zap.Error(err))
		cfg = &config.Config{}
	}
	a.cfg = cfg

	// Initialize storage
	sessionDir := cfg.SessionDir
	if sessionDir == "" {
		sessionDir = "~/.local/share/libcode/sessions"
	}
	dbPath := sessionDir + "/sessions.db"
	db, err := storage.Open(dbPath)
	if err != nil {
		logger.Get().Error("Failed to initialize storage", zap.Error(err))
		return
	}
	a.db = db

	// Initialize tool registry
	toolRegistry := framework.NewRegistry()
	a.toolRegistry = toolRegistry

	// Initialize AI providers
	aiProviders := provider.DefaultRegistry()
	a.aiProviders = aiProviders

	// Get the first available provider
	providers := aiProviders.List()
	for _, p := range providers {
		a.currentAI = p
		break
	}

	logger.Get().Info("Desktop application started")
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	if a.db != nil {
		_ = a.db.Close()
	}
	logger.Get().Info("Desktop application shutdown")
}

// SendMessage sends a message to the AI
func (a *App) SendMessage(message string) (string, error) {
	if a.currentAI == nil {
		return "", ErrAIProviderNotInitialized
	}

	req := &ai.Request{
		Messages: []ai.Message{
			{Role: "user", Content: message},
		},
	}

	response, err := a.currentAI.Complete(a.ctx, req)
	if err != nil {
		return "", err
	}

	return response.Message.Content, nil
}

// GetConfig returns the current configuration
func (a *App) GetConfig() map[string]interface{} {
	if a.cfg == nil {
		return make(map[string]interface{})
	}

	// Convert providers to a simpler map for JSON
	providers := make(map[string]interface{})
	for name, p := range a.cfg.Providers {
		providers[name] = map[string]interface{}{
			"apiKey":  p.APIKey,
			"baseURL": p.BaseURL,
			"enabled": p.Enabled,
			"model":   p.Model,
		}
	}

	return map[string]interface{}{
		"providers":      providers,
		"defaultModel":   a.cfg.DefaultModel,
		"modelVariant":   a.cfg.ModelVariant,
		"sessionDir":     a.cfg.SessionDir,
		"maxHistory":     a.cfg.MaxHistory,
		"logLevel":       a.cfg.LogLevel,
		"logFile":        a.cfg.LogFile,
		"tui":            a.cfg.Features.TUI,
		"desktopMode":    a.cfg.Features.DesktopMode,
		"streamResponse": a.cfg.Features.StreamResponse,
	}
}

// SaveConfig saves the configuration
func (a *App) SaveConfig(configData map[string]interface{}) error {
	if a.cfg == nil {
		return ErrConfigNotInitialized
	}

	// Update config from map
	if defaultModel, ok := configData["defaultModel"].(string); ok {
		a.cfg.DefaultModel = defaultModel
	}
	if modelVariant, ok := configData["modelVariant"].(string); ok {
		a.cfg.ModelVariant = modelVariant
	}
	if sessionDir, ok := configData["sessionDir"].(string); ok {
		a.cfg.SessionDir = sessionDir
	}
	if maxHistory, ok := configData["maxHistory"].(float64); ok {
		a.cfg.MaxHistory = int(maxHistory)
	}
	if logLevel, ok := configData["logLevel"].(string); ok {
		a.cfg.LogLevel = logLevel
	}
	if logFile, ok := configData["logFile"].(string); ok {
		a.cfg.LogFile = logFile
	}

	return a.cfg.Validate()
}

// GetSessions returns all sessions
func (a *App) GetSessions() ([]map[string]interface{}, error) {
	if a.db == nil {
		return nil, ErrDBNotInitialized
	}

	sessions, err := a.db.ListSessions(100)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(sessions))
	for i, s := range sessions {
		result[i] = map[string]interface{}{
			"id":         s.ID,
			"title":      s.Title,
			"created_at": s.CreatedAt,
			"updated_at": s.UpdatedAt,
		}
	}

	return result, nil
}

// LoadSession loads a session by ID
func (a *App) LoadSession(id string) (map[string]interface{}, error) {
	if a.db == nil {
		return nil, ErrDBNotInitialized
	}

	sess, err := a.db.GetSession(id)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":         sess.ID,
		"title":      sess.Title,
		"messages":   sess.Messages,
		"created_at": sess.CreatedAt,
		"updated_at": sess.UpdatedAt,
	}, nil
}

// ExecuteTool executes a tool with the given parameters
func (a *App) ExecuteTool(toolName string, params map[string]interface{}) (map[string]interface{}, error) {
	if a.toolRegistry == nil {
		return nil, ErrToolRegistryNotInitialized
	}

	tool, err := a.toolRegistry.Get(toolName)
	if err != nil {
		return nil, err
	}

	// Convert params from map[string]interface{} to map[string]any
	paramsAny := make(map[string]any)
	for k, v := range params {
		paramsAny[k] = v
	}

	// Execute the tool with nil ExecutionContext for desktop
	result, err := tool.Execute(a.ctx, paramsAny, nil)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"output": result,
	}, nil
}

// ListTools returns all available tools
func (a *App) ListTools() ([]map[string]interface{}, error) {
	if a.toolRegistry == nil {
		return nil, ErrToolRegistryNotInitialized
	}

	toolIDs := a.toolRegistry.List()
	result := make([]map[string]interface{}, len(toolIDs))

	for i, id := range toolIDs {
		tool, err := a.toolRegistry.Get(id)
		if err != nil {
			continue
		}
		result[i] = map[string]interface{}{
			"id":          tool.ID(),
			"description": tool.Description(),
			"schema":      tool.Parameters(),
		}
	}

	return result, nil
}

// Errors
var (
	ErrAIProviderNotInitialized   = &AppError{Message: "AI provider not initialized"}
	ErrConfigNotInitialized       = &AppError{Message: "Config not initialized"}
	ErrDBNotInitialized           = &AppError{Message: "Database not initialized"}
	ErrToolRegistryNotInitialized = &AppError{Message: "Tool registry not initialized"}
)

// AppError represents an application error
type AppError struct {
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Libcode",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		// macOS options
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            true,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			WebviewIsTransparent: true,
			WindowIsTranslucent:  false,
			About: &mac.AboutInfo{
				Title:   "Libcode",
				Message: "AI-powered development tool",
			},
		},
		// Windows options
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		// Linux options
		Linux: &linux.Options{
			Icon: []byte{}, // Add icon here
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
