// Package config contains configuration management for the TelemetryFlow GO MCP service
//
// TelemetryFlow GO MCP Server - Community Enterprise Observability Platform
// Copyright (c) 2024-2026 Telemetri Data Indonesia. All rights reserved.
// Open Source Software built by Telemetri Data Indonesia.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the MCP server
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Claude     ClaudeConfig     `mapstructure:"claude"`
	MCP        MCPConfig        `mapstructure:"mcp"`
	Logging    LoggingConfig    `mapstructure:"logging"`
	Telemetry  TelemetryConfig  `mapstructure:"telemetry"`
	Security   SecurityConfig   `mapstructure:"security"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Clickhouse ClickHouseConfig `mapstructure:"clickhouse"`
}

type DatabaseConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	URL          string `mapstructure:"url"`
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	SSLMode      string `mapstructure:"sslmode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	AutoMigrate  bool   `mapstructure:"auto_migrate"`
	SeedData     bool   `mapstructure:"seed_data"`
}

type ClickHouseConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	URL         string `mapstructure:"url"`
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Database    string `mapstructure:"database"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	Compression string `mapstructure:"compression"`
	Secure      bool   `mapstructure:"secure"`
	AutoMigrate bool   `mapstructure:"auto_migrate"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`

	// Transport type: "stdio", "sse", "websocket"
	Transport string `mapstructure:"transport"`

	// Timeouts
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`

	// Debug mode
	Debug bool `mapstructure:"debug"`
}

// ClaudeConfig holds Claude API configuration
type ClaudeConfig struct {
	APIKey         string        `mapstructure:"api_key"`
	BaseURL        string        `mapstructure:"base_url"`
	DefaultModel   string        `mapstructure:"default_model"`
	MaxTokens      int           `mapstructure:"max_tokens"`
	Temperature    float64       `mapstructure:"temperature"`
	TopP           float64       `mapstructure:"top_p"`
	TopK           int           `mapstructure:"top_k"`
	Timeout        time.Duration `mapstructure:"timeout"`
	MaxRetries     int           `mapstructure:"max_retries"`
	RetryDelay     time.Duration `mapstructure:"retry_delay"`
	EnableBatching bool          `mapstructure:"enable_batching"`
}

// MCPConfig holds MCP protocol configuration
type MCPConfig struct {
	ProtocolVersion string `mapstructure:"protocol_version"`

	// Capabilities
	EnableTools     bool `mapstructure:"enable_tools"`
	EnableResources bool `mapstructure:"enable_resources"`
	EnablePrompts   bool `mapstructure:"enable_prompts"`
	EnableLogging   bool `mapstructure:"enable_logging"`
	EnableSampling  bool `mapstructure:"enable_sampling"`

	// Limits
	MaxToolsPerSession     int `mapstructure:"max_tools_per_session"`
	MaxResourcesPerSession int `mapstructure:"max_resources_per_session"`
	MaxPromptsPerSession   int `mapstructure:"max_prompts_per_session"`
	MaxConversations       int `mapstructure:"max_conversations"`
	MaxMessagesPerConv     int `mapstructure:"max_messages_per_conv"`

	// Tool execution
	ToolTimeout time.Duration `mapstructure:"tool_timeout"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"` // "json" or "text"
	Output     string `mapstructure:"output"` // "stdout", "stderr", or file path
	AddSource  bool   `mapstructure:"add_source"`
	TimeFormat string `mapstructure:"time_format"`
}

// TelemetryConfig holds OpenTelemetry configuration
type TelemetryConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	ServiceName  string `mapstructure:"service_name"`
	Environment  string `mapstructure:"environment"`
	OTLPEndpoint string `mapstructure:"otlp_endpoint"`
	OTLPInsecure bool   `mapstructure:"otlp_insecure"`

	// Trace sampling
	TraceSampleRate float64 `mapstructure:"trace_sample_rate"`

	// Metrics
	MetricsEnabled  bool          `mapstructure:"metrics_enabled"`
	MetricsInterval time.Duration `mapstructure:"metrics_interval"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	// API Key validation
	RequireAPIKey  bool     `mapstructure:"require_api_key"`
	AllowedAPIKeys []string `mapstructure:"allowed_api_keys"`

	// Rate limiting
	RateLimitEnabled   bool `mapstructure:"rate_limit_enabled"`
	RateLimitPerMinute int  `mapstructure:"rate_limit_per_minute"`

	// CORS (for SSE transport)
	CORSEnabled        bool     `mapstructure:"cors_enabled"`
	CORSAllowedOrigins []string `mapstructure:"cors_allowed_origins"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Name:            "TelemetryFlow-MCP",
			Version:         "1.2.0",
			Host:            "localhost",
			Port:            8080,
			Transport:       "stdio",
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			ShutdownTimeout: 10 * time.Second,
			Debug:           false,
		},
		Claude: ClaudeConfig{
			BaseURL:        "https://api.anthropic.com",
			DefaultModel:   "claude-opus-4-7",
			MaxTokens:      4096,
			Temperature:    1.0,
			TopP:           1.0,
			TopK:           0,
			Timeout:        120 * time.Second,
			MaxRetries:     3,
			RetryDelay:     1 * time.Second,
			EnableBatching: false,
		},
		MCP: MCPConfig{
			ProtocolVersion:        "2024-11-05",
			EnableTools:            true,
			EnableResources:        true,
			EnablePrompts:          true,
			EnableLogging:          true,
			EnableSampling:         false,
			MaxToolsPerSession:     100,
			MaxResourcesPerSession: 100,
			MaxPromptsPerSession:   50,
			MaxConversations:       10,
			MaxMessagesPerConv:     1000,
			ToolTimeout:            30 * time.Second,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stderr",
			AddSource:  false,
			TimeFormat: time.RFC3339,
		},
		Telemetry: TelemetryConfig{
			Enabled:         true,
			ServiceName:     "telemetryflow-go-mcp",
			Environment:     "development",
			OTLPEndpoint:    "localhost:4317",
			OTLPInsecure:    true,
			TraceSampleRate: 1.0,
			MetricsEnabled:  true,
			MetricsInterval: 30 * time.Second,
		},
		Security: SecurityConfig{
			RequireAPIKey:      false,
			RateLimitEnabled:   true,
			RateLimitPerMinute: 100,
			CORSEnabled:        true,
			CORSAllowedOrigins: []string{"*"},
		},
		Database: DatabaseConfig{
			Enabled:      false,
			Host:         "localhost",
			Port:         5432,
			User:         "telemetryflow",
			Database:     "telemetryflow_mcp",
			SSLMode:      "disable",
			MaxOpenConns: 25,
			MaxIdleConns: 5,
			AutoMigrate:  true,
			SeedData:     true,
		},
		Clickhouse: ClickHouseConfig{
			Enabled:     false,
			Host:        "localhost",
			Port:        9000,
			Database:    "telemetryflow_analytics",
			Username:    "default",
			Compression: "lz4",
			Secure:      false,
			AutoMigrate: true,
		},
	}
}

// Load loads configuration from files and environment
func Load(configPath string) (*Config, error) {
	config := DefaultConfig()

	v := viper.New()
	v.SetConfigType("yaml")

	// Set config file if provided
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Look for config in standard locations
		v.SetConfigName("config")
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
		v.AddConfigPath("/etc/telemetryflow-go-mcp")
		v.AddConfigPath("$HOME/.telemetryflow-go-mcp")
	}

	// Environment variable settings
	v.SetEnvPrefix("TELEMETRYFLOW_MCP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Bind specific environment variables
	bindEnvVars(v)

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found, use defaults and env vars
	}

	// Unmarshal config
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Override with environment variables for sensitive data
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.Claude.APIKey = apiKey
	}
	if apiKey := os.Getenv("TELEMETRYFLOW_MCP_CLAUDE_API_KEY"); apiKey != "" {
		config.Claude.APIKey = apiKey
	}

	if dbURL := os.Getenv("TELEMETRYFLOW_MCP_POSTGRES_URL"); dbURL != "" && config.Database.URL == "" {
		config.Database.URL = dbURL
		config.Database.Enabled = true
	}
	if chURL := os.Getenv("TELEMETRYFLOW_MCP_CLICKHOUSE_URL"); chURL != "" && config.Clickhouse.URL == "" {
		config.Clickhouse.URL = chURL
		config.Clickhouse.Enabled = true
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// bindEnvVars binds environment variables to config keys
func bindEnvVars(v *viper.Viper) {
	// Claude API (errors ignored as BindEnv only fails on empty key names)
	_ = v.BindEnv("claude.api_key", "ANTHROPIC_API_KEY", "TELEMETRYFLOW_MCP_CLAUDE_API_KEY")
	_ = v.BindEnv("claude.base_url", "TELEMETRYFLOW_MCP_CLAUDE_BASE_URL")
	_ = v.BindEnv("claude.default_model", "TELEMETRYFLOW_MCP_CLAUDE_DEFAULT_MODEL")

	// Server
	_ = v.BindEnv("server.host", "TELEMETRYFLOW_MCP_SERVER_HOST")
	_ = v.BindEnv("server.port", "TELEMETRYFLOW_MCP_SERVER_PORT")
	_ = v.BindEnv("server.transport", "TELEMETRYFLOW_MCP_SERVER_TRANSPORT")
	_ = v.BindEnv("server.debug", "TELEMETRYFLOW_MCP_DEBUG")

	// Logging
	_ = v.BindEnv("logging.level", "TELEMETRYFLOW_MCP_LOG_LEVEL")
	_ = v.BindEnv("logging.format", "TELEMETRYFLOW_MCP_LOG_FORMAT")

	// Telemetry
	_ = v.BindEnv("telemetry.enabled", "TELEMETRYFLOW_MCP_TELEMETRY_ENABLED")
	_ = v.BindEnv("telemetry.otlp_endpoint", "TELEMETRYFLOW_ENDPOINT", "TELEMETRYFLOW_MCP_OTLP_ENDPOINT")
	_ = v.BindEnv("telemetry.service_name", "TELEMETRYFLOW_SERVICE_NAME", "TELEMETRYFLOW_MCP_SERVICE_NAME")

	// Database (PostgreSQL)
	_ = v.BindEnv("database.enabled", "TELEMETRYFLOW_MCP_DATABASE_ENABLED")
	_ = v.BindEnv("database.url", "TELEMETRYFLOW_MCP_POSTGRES_URL")
	_ = v.BindEnv("database.host", "TELEMETRYFLOW_MCP_POSTGRES_HOST")
	_ = v.BindEnv("database.port", "TELEMETRYFLOW_MCP_POSTGRES_PORT")
	_ = v.BindEnv("database.user", "TELEMETRYFLOW_MCP_POSTGRES_USER")
	_ = v.BindEnv("database.password", "TELEMETRYFLOW_MCP_POSTGRES_PASSWORD")
	_ = v.BindEnv("database.database", "TELEMETRYFLOW_MCP_POSTGRES_DATABASE")
	_ = v.BindEnv("database.sslmode", "TELEMETRYFLOW_MCP_POSTGRES_SSLMODE")
	_ = v.BindEnv("database.auto_migrate", "TELEMETRYFLOW_MCP_POSTGRES_AUTO_MIGRATE")
	_ = v.BindEnv("database.seed_data", "TELEMETRYFLOW_MCP_POSTGRES_SEED_DATA")

	// ClickHouse
	_ = v.BindEnv("clickhouse.enabled", "TELEMETRYFLOW_MCP_CLICKHOUSE_ENABLED")
	_ = v.BindEnv("clickhouse.url", "TELEMETRYFLOW_MCP_CLICKHOUSE_URL")
	_ = v.BindEnv("clickhouse.host", "TELEMETRYFLOW_MCP_CLICKHOUSE_HOST")
	_ = v.BindEnv("clickhouse.port", "TELEMETRYFLOW_MCP_CLICKHOUSE_PORT")
	_ = v.BindEnv("clickhouse.database", "TELEMETRYFLOW_MCP_CLICKHOUSE_DATABASE")
	_ = v.BindEnv("clickhouse.username", "TELEMETRYFLOW_MCP_CLICKHOUSE_USERNAME")
	_ = v.BindEnv("clickhouse.password", "TELEMETRYFLOW_MCP_CLICKHOUSE_PASSWORD")
	_ = v.BindEnv("clickhouse.auto_migrate", "TELEMETRYFLOW_MCP_CLICKHOUSE_AUTO_MIGRATE")
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Claude.APIKey == "" {
		return errors.New("claude.api_key is required (set ANTHROPIC_API_KEY environment variable)")
	}

	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return errors.New("server.port must be between 1 and 65535")
	}

	validTransports := map[string]bool{"stdio": true, "sse": true, "websocket": true}
	if !validTransports[c.Server.Transport] {
		return errors.New("server.transport must be 'stdio', 'sse', or 'websocket'")
	}

	if c.Claude.MaxTokens < 1 {
		return errors.New("claude.max_tokens must be positive")
	}

	if c.Claude.Temperature < 0 || c.Claude.Temperature > 2 {
		return errors.New("claude.temperature must be between 0 and 2")
	}

	if c.Telemetry.TraceSampleRate < 0 || c.Telemetry.TraceSampleRate > 1 {
		return errors.New("telemetry.trace_sample_rate must be between 0 and 1")
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Telemetry.Environment == "development" || c.Server.Debug
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Telemetry.Environment == "production"
}
