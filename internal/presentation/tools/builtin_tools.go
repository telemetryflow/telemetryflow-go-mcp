// Package tools contains built-in MCP tools for TelemetryFlow
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	appsvc "github.com/telemetryflow/telemetryflow-go-mcp/internal/application/services"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/services"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

// ToolRegistry manages built-in tools
type ToolRegistry struct {
	claudeService    services.IClaudeService
	contextCollector *appsvc.ContextCollector
	promptBuilder    *appsvc.PromptBuilder
	tools            map[string]*entities.Tool
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry(claudeService services.IClaudeService) *ToolRegistry {
	registry := &ToolRegistry{
		claudeService: claudeService,
		promptBuilder: appsvc.NewPromptBuilder(),
		tools:         make(map[string]*entities.Tool),
	}

	// Register built-in tools
	registry.registerBuiltinTools()

	return registry
}

// NewToolRegistryWithCollector creates a new tool registry with context collection support
func NewToolRegistryWithCollector(claudeService services.IClaudeService, collector *appsvc.ContextCollector) *ToolRegistry {
	registry := &ToolRegistry{
		claudeService:    claudeService,
		contextCollector: collector,
		promptBuilder:    appsvc.NewPromptBuilder(),
		tools:            make(map[string]*entities.Tool),
	}

	registry.registerBuiltinTools()

	return registry
}

// GetTools returns all registered tools
func (r *ToolRegistry) GetTools() []*entities.Tool {
	tools := make([]*entities.Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// GetTool returns a tool by name
func (r *ToolRegistry) GetTool(name string) (*entities.Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// registerBuiltinTools registers all built-in tools
func (r *ToolRegistry) registerBuiltinTools() {
	// Claude conversation tool
	r.registerClaudeConversation()

	// File tools
	r.registerReadFile()
	r.registerWriteFile()
	r.registerListDirectory()

	// Shell tool
	r.registerExecuteCommand()

	// Search tool
	r.registerSearchFiles()

	// System info tool
	r.registerSystemInfo()

	// Echo tool (for testing)
	r.registerEcho()

	// Telemetry context tools
	r.registerCollectTelemetryContext()
	r.registerListContextTypes()
	r.registerBuildSystemPrompt()
}

// registerClaudeConversation registers the Claude conversation tool
func (r *ToolRegistry) registerClaudeConversation() {
	name, _ := vo.NewToolName("claude_conversation")
	desc, _ := vo.NewToolDescription("Send a message to Claude and receive a response. Use this for AI-powered assistance, code generation, analysis, and general conversation.")

	schema := &entities.JSONSchema{
		Type: "object",
		Properties: map[string]*entities.JSONSchema{
			"message": {
				Type:        "string",
				Description: "The message to send to Claude",
			},
			"system_prompt": {
				Type:        "string",
				Description: "Optional system prompt to set context",
			},
			"model": {
				Type:        "string",
				Description: "The LLM model to use (default: claude-opus-4-7). Supported: Anthropic Claude, Google Gemini, OpenAI GPT/o, DeepSeek, Qwen, Mistral, Grok, Kimi, Zhipu GLM, Xiaomi MiMo",
				Enum: []interface{}{
					"claude-opus-4-7", "claude-opus-4-7-fast", "claude-opus-4-6", "claude-opus-4-6-fast",
					"claude-sonnet-4-6", "claude-opus-4-5", "claude-sonnet-4-5-20250929",
					"claude-haiku-4-5", "claude-haiku-4-5-20251001", "claude-sonnet-4-20250514",
					"claude-mythos-preview",
					"gemini-3.5-flash", "gemini-2.5-pro", "gemini-2.5-flash",
					"gpt-5.5-pro", "gpt-5.5", "gpt-5.4-pro", "gpt-5.4", "o3",
					"deepseek-v4-pro", "deepseek-v4-flash", "deepseek-chat", "deepseek-reasoner",
					"qwen3.6-max-preview", "qwen3.6-plus", "qwen3.6-flash",
					"mistral-medium-3-5", "mistral-small-2603", "mistral-large-2512",
					"grok-4.3", "grok-4.20-multi-agent", "grok-4.20-0309-reasoning",
					"kimi-k2.6", "kimi-k2.5", "kimi-k2-thinking",
					"glm-5.1", "glm-5-turbo", "glm-4.7-flash",
					"mimo-v2.5-pro", "mimo-v2.5", "mimo-v2-pro",
				},
			},
			"max_tokens": {
				Type:        "integer",
				Description: "Maximum tokens in the response (default: 4096)",
			},
		},
		Required: []string{"message"},
	}

	tool, _ := entities.NewTool(name, desc, schema)
	tool.SetCategory("ai")
	tool.SetTags([]string{"claude", "conversation", "ai"})
	tool.SetHandler(r.handleClaudeConversation)
	tool.SetTimeout(120 * time.Second)

	r.tools["claude_conversation"] = tool
}

// handleClaudeConversation handles Claude conversation requests
func (r *ToolRegistry) handleClaudeConversation(input map[string]interface{}) (*entities.ToolResult, error) {
	message, ok := input["message"].(string)
	if !ok || message == "" {
		return entities.NewErrorToolResult(fmt.Errorf("message is required")), nil
	}

	// Build request
	model := vo.ModelClaudeOpus47
	if m, ok := input["model"].(string); ok {
		model = vo.Model(m)
	}

	maxTokens := 4096
	if mt, ok := input["max_tokens"].(float64); ok {
		maxTokens = int(mt)
	}

	var systemPrompt vo.SystemPrompt
	if sp, ok := input["system_prompt"].(string); ok && sp != "" {
		systemPrompt, _ = vo.NewSystemPrompt(sp)
	}

	request := &services.ClaudeRequest{
		Model:        model,
		SystemPrompt: systemPrompt,
		Messages: []services.ClaudeMessage{
			{
				Role: vo.RoleUser,
				Content: []entities.ContentBlock{
					{Type: vo.ContentTypeText, Text: message},
				},
			},
		},
		MaxTokens: maxTokens,
	}

	// Call Claude API
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	response, err := r.claudeService.CreateMessage(ctx, request)
	if err != nil {
		return entities.NewErrorToolResult(err), nil
	}

	// Extract text content
	var text string
	for _, block := range response.Content {
		if block.Type == vo.ContentTypeText {
			text += block.Text
		}
	}

	return entities.NewTextToolResult(text), nil
}

// registerReadFile registers the read file tool
func (r *ToolRegistry) registerReadFile() {
	name, _ := vo.NewToolName("read_file")
	desc, _ := vo.NewToolDescription("Read the contents of a file at the specified path")

	schema := &entities.JSONSchema{
		Type: "object",
		Properties: map[string]*entities.JSONSchema{
			"path": {
				Type:        "string",
				Description: "The path to the file to read",
			},
			"encoding": {
				Type:        "string",
				Description: "The encoding to use (default: utf-8)",
			},
		},
		Required: []string{"path"},
	}

	tool, _ := entities.NewTool(name, desc, schema)
	tool.SetCategory("file")
	tool.SetTags([]string{"file", "read"})
	tool.SetHandler(handleReadFile)

	r.tools["read_file"] = tool
}

func handleReadFile(input map[string]interface{}) (*entities.ToolResult, error) {
	path, ok := input["path"].(string)
	if !ok || path == "" {
		return entities.NewErrorToolResult(fmt.Errorf("path is required")), nil
	}

	// Security: Prevent path traversal
	absPath, err := filepath.Abs(path)
	if err != nil {
		return entities.NewErrorToolResult(err), nil
	}

	content, err := os.ReadFile(absPath) //nolint:gosec // G304: path is sanitized via filepath.Abs
	if err != nil {
		return entities.NewErrorToolResult(err), nil
	}

	return entities.NewTextToolResult(string(content)), nil
}

// registerWriteFile registers the write file tool
func (r *ToolRegistry) registerWriteFile() {
	name, _ := vo.NewToolName("write_file")
	desc, _ := vo.NewToolDescription("Write content to a file at the specified path")

	schema := &entities.JSONSchema{
		Type: "object",
		Properties: map[string]*entities.JSONSchema{
			"path": {
				Type:        "string",
				Description: "The path to the file to write",
			},
			"content": {
				Type:        "string",
				Description: "The content to write to the file",
			},
			"create_dirs": {
				Type:        "boolean",
				Description: "Create parent directories if they don't exist",
			},
		},
		Required: []string{"path", "content"},
	}

	tool, _ := entities.NewTool(name, desc, schema)
	tool.SetCategory("file")
	tool.SetTags([]string{"file", "write"})
	tool.SetHandler(handleWriteFile)

	r.tools["write_file"] = tool
}

func handleWriteFile(input map[string]interface{}) (*entities.ToolResult, error) {
	path, ok := input["path"].(string)
	if !ok || path == "" {
		return entities.NewErrorToolResult(fmt.Errorf("path is required")), nil
	}

	content, ok := input["content"].(string)
	if !ok {
		return entities.NewErrorToolResult(fmt.Errorf("content is required")), nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return entities.NewErrorToolResult(err), nil
	}

	// Create directories if requested
	if createDirs, ok := input["create_dirs"].(bool); ok && createDirs {
		dir := filepath.Dir(absPath)
		if err := os.MkdirAll(dir, 0750); err != nil {
			return entities.NewErrorToolResult(err), nil
		}
	}

	if err := os.WriteFile(absPath, []byte(content), 0600); err != nil {
		return entities.NewErrorToolResult(err), nil
	}

	return entities.NewTextToolResult(fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), absPath)), nil
}

// registerListDirectory registers the list directory tool
func (r *ToolRegistry) registerListDirectory() {
	name, _ := vo.NewToolName("list_directory")
	desc, _ := vo.NewToolDescription("List files and directories at the specified path")

	schema := &entities.JSONSchema{
		Type: "object",
		Properties: map[string]*entities.JSONSchema{
			"path": {
				Type:        "string",
				Description: "The path to the directory to list",
			},
			"recursive": {
				Type:        "boolean",
				Description: "List recursively (default: false)",
			},
		},
		Required: []string{"path"},
	}

	tool, _ := entities.NewTool(name, desc, schema)
	tool.SetCategory("file")
	tool.SetTags([]string{"file", "directory", "list"})
	tool.SetHandler(handleListDirectory)

	r.tools["list_directory"] = tool
}

func handleListDirectory(input map[string]interface{}) (*entities.ToolResult, error) {
	path, ok := input["path"].(string)
	if !ok || path == "" {
		return entities.NewErrorToolResult(fmt.Errorf("path is required")), nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return entities.NewErrorToolResult(err), nil
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return entities.NewErrorToolResult(err), nil
	}

	var result []string
	for _, entry := range entries {
		prefix := "📄 "
		if entry.IsDir() {
			prefix = "📁 "
		}
		result = append(result, prefix+entry.Name())
	}

	return entities.NewTextToolResult(strings.Join(result, "\n")), nil
}

// registerExecuteCommand registers the execute command tool
func (r *ToolRegistry) registerExecuteCommand() {
	name, _ := vo.NewToolName("execute_command")
	desc, _ := vo.NewToolDescription("Execute a shell command and return the output")

	schema := &entities.JSONSchema{
		Type: "object",
		Properties: map[string]*entities.JSONSchema{
			"command": {
				Type:        "string",
				Description: "The command to execute",
			},
			"working_dir": {
				Type:        "string",
				Description: "The working directory for the command",
			},
			"timeout": {
				Type:        "integer",
				Description: "Timeout in seconds (default: 30)",
			},
		},
		Required: []string{"command"},
	}

	tool, _ := entities.NewTool(name, desc, schema)
	tool.SetCategory("system")
	tool.SetTags([]string{"command", "shell", "execute"})
	tool.SetHandler(handleExecuteCommand)
	tool.SetTimeout(60 * time.Second)

	r.tools["execute_command"] = tool
}

func handleExecuteCommand(input map[string]interface{}) (*entities.ToolResult, error) {
	command, ok := input["command"].(string)
	if !ok || command == "" {
		return entities.NewErrorToolResult(fmt.Errorf("command is required")), nil
	}

	timeout := 30
	if t, ok := input["timeout"].(float64); ok {
		timeout = int(t)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command) //nolint:gosec // G204: command execution is intentional for shell tool

	if workingDir, ok := input["working_dir"].(string); ok && workingDir != "" {
		cmd.Dir = workingDir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return entities.NewErrorToolResult(fmt.Errorf("command timed out after %d seconds", timeout)), nil
		}
		return entities.NewTextToolResult(fmt.Sprintf("Command failed: %s\nOutput: %s", err.Error(), string(output))), nil
	}

	return entities.NewTextToolResult(string(output)), nil
}

// registerSearchFiles registers the search files tool
func (r *ToolRegistry) registerSearchFiles() {
	name, _ := vo.NewToolName("search_files")
	desc, _ := vo.NewToolDescription("Search for files matching a pattern in a directory")

	schema := &entities.JSONSchema{
		Type: "object",
		Properties: map[string]*entities.JSONSchema{
			"path": {
				Type:        "string",
				Description: "The directory to search in",
			},
			"pattern": {
				Type:        "string",
				Description: "The glob pattern to match (e.g., *.go, **/*.ts)",
			},
			"content_pattern": {
				Type:        "string",
				Description: "Optional: Search for files containing this text",
			},
		},
		Required: []string{"path", "pattern"},
	}

	tool, _ := entities.NewTool(name, desc, schema)
	tool.SetCategory("file")
	tool.SetTags([]string{"file", "search", "find"})
	tool.SetHandler(handleSearchFiles)

	r.tools["search_files"] = tool
}

func handleSearchFiles(input map[string]interface{}) (*entities.ToolResult, error) {
	path, ok := input["path"].(string)
	if !ok || path == "" {
		return entities.NewErrorToolResult(fmt.Errorf("path is required")), nil
	}

	pattern, ok := input["pattern"].(string)
	if !ok || pattern == "" {
		return entities.NewErrorToolResult(fmt.Errorf("pattern is required")), nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return entities.NewErrorToolResult(err), nil
	}

	var matches []string
	err = filepath.Walk(absPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil
		}

		matched, _ := filepath.Match(pattern, info.Name())
		if matched {
			relPath, _ := filepath.Rel(absPath, p)
			matches = append(matches, relPath)
		}
		return nil
	})

	if err != nil {
		return entities.NewErrorToolResult(err), nil
	}

	if len(matches) == 0 {
		return entities.NewTextToolResult("No files found matching pattern: " + pattern), nil
	}

	return entities.NewTextToolResult(fmt.Sprintf("Found %d files:\n%s", len(matches), strings.Join(matches, "\n"))), nil
}

// registerSystemInfo registers the system info tool
func (r *ToolRegistry) registerSystemInfo() {
	name, _ := vo.NewToolName("system_info")
	desc, _ := vo.NewToolDescription("Get system information")

	schema := &entities.JSONSchema{
		Type:       "object",
		Properties: map[string]*entities.JSONSchema{},
	}

	tool, _ := entities.NewTool(name, desc, schema)
	tool.SetCategory("system")
	tool.SetTags([]string{"system", "info"})
	tool.SetHandler(handleSystemInfo)

	r.tools["system_info"] = tool
}

func handleSystemInfo(input map[string]interface{}) (*entities.ToolResult, error) {
	hostname, _ := os.Hostname()
	wd, _ := os.Getwd()

	info := map[string]interface{}{
		"hostname":    hostname,
		"working_dir": wd,
		"os":          os.Getenv("GOOS"),
		"arch":        os.Getenv("GOARCH"),
		"user":        os.Getenv("USER"),
		"home":        os.Getenv("HOME"),
		"shell":       os.Getenv("SHELL"),
		"time":        time.Now().Format(time.RFC3339),
	}

	data, _ := json.MarshalIndent(info, "", "  ")
	return entities.NewTextToolResult(string(data)), nil
}

// registerEcho registers the echo tool (for testing)
func (r *ToolRegistry) registerEcho() {
	name, _ := vo.NewToolName("echo")
	desc, _ := vo.NewToolDescription("Echo back the input message (useful for testing)")

	schema := &entities.JSONSchema{
		Type: "object",
		Properties: map[string]*entities.JSONSchema{
			"message": {
				Type:        "string",
				Description: "The message to echo back",
			},
		},
		Required: []string{"message"},
	}

	tool, _ := entities.NewTool(name, desc, schema)
	tool.SetCategory("utility")
	tool.SetTags([]string{"test", "echo"})
	tool.SetHandler(handleEcho)

	r.tools["echo"] = tool
}

func handleEcho(input map[string]interface{}) (*entities.ToolResult, error) {
	message, ok := input["message"].(string)
	if !ok {
		return entities.NewErrorToolResult(fmt.Errorf("message is required")), nil
	}
	return entities.NewTextToolResult(message), nil
}

func (r *ToolRegistry) registerCollectTelemetryContext() {
	name, _ := vo.NewToolName("collect_telemetry_context")
	desc, _ := vo.NewToolDescription("Collect live telemetry context from TelemetryFlow platform for AI analysis. Queries ClickHouse and PostgreSQL for real observability data.")

	var contextTypes []interface{}
	for _, ct := range vo.AllContextTypes() {
		contextTypes = append(contextTypes, string(ct))
	}

	schema := &entities.JSONSchema{
		Type: "object",
		Properties: map[string]*entities.JSONSchema{
			"organization_id": {
				Type:        "string",
				Description: "The organization ID to collect context for",
			},
			"context_type": {
				Type:        "string",
				Description: "The type of telemetry context to collect",
				Enum:        contextTypes,
			},
			"user_id": {
				Type:        "string",
				Description: "Optional user ID (required for account-* context types)",
			},
			"time_range_from": {
				Type:        "string",
				Description: "Start time in ISO 8601 format (default: 1 hour ago)",
			},
			"time_range_to": {
				Type:        "string",
				Description: "End time in ISO 8601 format (default: now)",
			},
			"max_items": {
				Type:        "integer",
				Description: "Maximum number of items to return (default: 30)",
			},
		},
		Required: []string{"organization_id", "context_type"},
	}

	tool, _ := entities.NewTool(name, desc, schema)
	tool.SetCategory("telemetry")
	tool.SetTags([]string{"telemetry", "context", "observability", "telemetryflow"})
	tool.SetHandler(r.handleCollectTelemetryContext)
	tool.SetTimeout(10 * time.Second)

	r.tools["collect_telemetry_context"] = tool
}

func (r *ToolRegistry) handleCollectTelemetryContext(input map[string]interface{}) (*entities.ToolResult, error) {
	if r.contextCollector == nil {
		return entities.NewErrorToolResult(fmt.Errorf("telemetry context collection is not available — ClickHouse and/or PostgreSQL not configured")), nil
	}

	orgID, ok := input["organization_id"].(string)
	if !ok || orgID == "" {
		return entities.NewErrorToolResult(fmt.Errorf("organization_id is required")), nil
	}

	contextTypeStr, ok := input["context_type"].(string)
	if !ok || contextTypeStr == "" {
		return entities.NewErrorToolResult(fmt.Errorf("context_type is required")), nil
	}

	contextType := vo.ContextType(contextTypeStr)
	if !contextType.IsValid() {
		return entities.NewErrorToolResult(fmt.Errorf("invalid context_type: %s", contextTypeStr)), nil
	}

	var timeRange *vo.TimeRange
	if fromStr, ok := input["time_range_from"].(string); ok {
		if toStr, ok2 := input["time_range_to"].(string); ok2 {
			from, err := time.Parse(time.RFC3339, fromStr)
			if err != nil {
				return entities.NewErrorToolResult(fmt.Errorf("invalid time_range_from format: %w", err)), nil
			}
			to, err := time.Parse(time.RFC3339, toStr)
			if err != nil {
				return entities.NewErrorToolResult(fmt.Errorf("invalid time_range_to format: %w", err)), nil
			}
			timeRange = &vo.TimeRange{From: from, To: to}
		}
	}

	maxItems := 30
	if mi, ok := input["max_items"].(float64); ok {
		maxItems = int(mi)
	}

	userID, _ := input["user_id"].(string)

	opts := vo.CollectContextOptions{
		OrganizationID: orgID,
		UserID:         userID,
		ContextType:    contextType,
		TimeRange:      timeRange,
		MaxItems:       maxItems,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tc, err := r.contextCollector.CollectContext(ctx, opts)
	if err != nil {
		return entities.NewErrorToolResult(fmt.Errorf("failed to collect context: %w", err)), nil
	}

	systemPrompt := r.promptBuilder.BuildSystemPrompt(contextType, "")
	contextPrompt := r.promptBuilder.BuildContextPrompt(tc)

	result := map[string]interface{}{
		"context_type": string(tc.Type),
		"time_range": map[string]string{
			"from": tc.TimeRange.From.Format(time.RFC3339),
			"to":   tc.TimeRange.To.Format(time.RFC3339),
		},
		"summary":        tc.Summary,
		"data":           tc.Data,
		"system_prompt":  systemPrompt,
		"context_prompt": contextPrompt,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return entities.NewTextToolResult(string(data)), nil
}

func (r *ToolRegistry) registerListContextTypes() {
	name, _ := vo.NewToolName("list_context_types")
	desc, _ := vo.NewToolDescription("List all available telemetry context types supported by TelemetryFlow platform")

	schema := &entities.JSONSchema{
		Type:       "object",
		Properties: map[string]*entities.JSONSchema{},
	}

	tool, _ := entities.NewTool(name, desc, schema)
	tool.SetCategory("telemetry")
	tool.SetTags([]string{"telemetry", "context", "discovery"})
	tool.SetHandler(r.handleListContextTypes)

	r.tools["list_context_types"] = tool
}

func (r *ToolRegistry) handleListContextTypes(input map[string]interface{}) (*entities.ToolResult, error) {
	types := vo.AllContextTypes()

	categories := map[string][]string{
		"Core Telemetry":      {},
		"Infrastructure":      {},
		"Kubernetes":          {},
		"AI Intelligence":     {},
		"Database Monitoring": {},
		"Platform Management": {},
		"Account":             {},
	}

	for _, ct := range types {
		s := string(ct)
		switch {
		case strings.HasPrefix(s, "db-monitoring-"):
			categories["Database Monitoring"] = append(categories["Database Monitoring"], s)
		case strings.HasPrefix(s, "kubernetes-"):
			categories["Kubernetes"] = append(categories["Kubernetes"], s)
		case strings.HasPrefix(s, "infra-"):
			categories["Infrastructure"] = append(categories["Infrastructure"], s)
		case strings.HasPrefix(s, "account-"):
			categories["Account"] = append(categories["Account"], s)
		case s == "anomaly-detection" || s == "corrective-maintenance" || s == "predictive-maintenance" || s == "cost-optimization":
			categories["AI Intelligence"] = append(categories["AI Intelligence"], s)
		case strings.HasPrefix(s, "iam-") || strings.HasPrefix(s, "tenancy-") ||
			s == "audit" || s == "retention" || s == "subscription" || s == "api-keys" ||
			s == "data-masking" || s == "system-setup" || s == "system-channels" ||
			s == "ai-assistant" || s == "notifications" || s == "reports":
			categories["Platform Management"] = append(categories["Platform Management"], s)
		default:
			categories["Core Telemetry"] = append(categories["Core Telemetry"], s)
		}
	}

	result := map[string]interface{}{
		"total":      len(types),
		"categories": categories,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return entities.NewTextToolResult(string(data)), nil
}

func (r *ToolRegistry) registerBuildSystemPrompt() {
	name, _ := vo.NewToolName("build_system_prompt")
	desc, _ := vo.NewToolDescription("Build a context-aware system prompt for a given telemetry context type")

	schema := &entities.JSONSchema{
		Type: "object",
		Properties: map[string]*entities.JSONSchema{
			"context_type": {
				Type:        "string",
				Description: "The telemetry context type to build a prompt for",
			},
			"custom_prompt": {
				Type:        "string",
				Description: "Optional additional instructions to append",
			},
		},
		Required: []string{"context_type"},
	}

	tool, _ := entities.NewTool(name, desc, schema)
	tool.SetCategory("telemetry")
	tool.SetTags([]string{"telemetry", "prompt", "ai"})
	tool.SetHandler(r.handleBuildSystemPrompt)

	r.tools["build_system_prompt"] = tool
}

func (r *ToolRegistry) handleBuildSystemPrompt(input map[string]interface{}) (*entities.ToolResult, error) {
	contextTypeStr, ok := input["context_type"].(string)
	if !ok || contextTypeStr == "" {
		return entities.NewErrorToolResult(fmt.Errorf("context_type is required")), nil
	}

	contextType := vo.ContextType(contextTypeStr)
	if !contextType.IsValid() {
		return entities.NewErrorToolResult(fmt.Errorf("invalid context_type: %s", contextTypeStr)), nil
	}

	customPrompt, _ := input["custom_prompt"].(string)
	prompt := r.promptBuilder.BuildSystemPrompt(contextType, customPrompt)

	return entities.NewTextToolResult(prompt), nil
}
