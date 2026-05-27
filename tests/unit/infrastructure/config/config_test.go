package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	assert.Equal(t, "TelemetryFlow-MCP", cfg.Server.Name)
	assert.Equal(t, "1.2.0", cfg.Server.Version)
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "stdio", cfg.Server.Transport)
	assert.Equal(t, "claude-opus-4-7", cfg.Claude.DefaultModel)
	assert.Equal(t, 4096, cfg.Claude.MaxTokens)
	assert.Equal(t, "2024-11-05", cfg.MCP.ProtocolVersion)
	assert.True(t, cfg.MCP.EnableTools)
	assert.True(t, cfg.MCP.EnableResources)
	assert.True(t, cfg.MCP.EnablePrompts)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.True(t, cfg.Telemetry.Enabled)
	assert.False(t, cfg.Security.RequireAPIKey)
	assert.False(t, cfg.Database.Enabled)
	assert.False(t, cfg.Clickhouse.Enabled)
}

func TestConfig_Validate_MissingAPIKey(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Claude.APIKey = ""
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "claude.api_key")
}

func TestConfig_Validate_InvalidPort(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Claude.APIKey = "test-key"
	cfg.Server.Port = 0
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "server.port")
}

func TestConfig_Validate_InvalidTransport(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Claude.APIKey = "test-key"
	cfg.Server.Transport = "invalid"
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "server.transport")
}

func TestConfig_Validate_InvalidMaxTokens(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Claude.APIKey = "test-key"
	cfg.Claude.MaxTokens = 0
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "claude.max_tokens")
}

func TestConfig_Validate_InvalidTemperature(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Claude.APIKey = "test-key"
	cfg.Claude.Temperature = 3.0
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "claude.temperature")
}

func TestConfig_Validate_InvalidTraceSampleRate(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Claude.APIKey = "test-key"
	cfg.Telemetry.TraceSampleRate = 2.0
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "telemetry.trace_sample_rate")
}

func TestConfig_Validate_Success(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Claude.APIKey = "sk-ant-test-key"
	err := cfg.Validate()
	require.NoError(t, err)
}

func TestConfig_IsDevelopment(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Telemetry.Environment = "development"
	assert.True(t, cfg.IsDevelopment())

	cfg.Telemetry.Environment = "production"
	cfg.Server.Debug = true
	assert.True(t, cfg.IsDevelopment())

	cfg.Server.Debug = false
	assert.False(t, cfg.IsDevelopment())
}

func TestConfig_IsProduction(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Telemetry.Environment = "production"
	assert.True(t, cfg.IsProduction())

	cfg.Telemetry.Environment = "staging"
	assert.False(t, cfg.IsProduction())
}

func TestConfig_Load(t *testing.T) {
	t.Run("no config file uses defaults with env", func(t *testing.T) {
		t.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key")
		cfg, err := config.Load("")
		require.NoError(t, err)
		assert.Equal(t, "sk-ant-test-key", cfg.Claude.APIKey)
	})

	t.Run("from yaml file", func(t *testing.T) {
		dir := t.TempDir()
		cfgPath := filepath.Join(dir, "config.yaml")
		content := []byte("claude:\n  api_key: sk-from-file\nserver:\n  port: 9090\n")
		require.NoError(t, os.WriteFile(cfgPath, content, 0644))
		cfg, err := config.Load(cfgPath)
		require.NoError(t, err)
		assert.Equal(t, "sk-from-file", cfg.Claude.APIKey)
		assert.Equal(t, 9090, cfg.Server.Port)
	})

	t.Run("invalid config file", func(t *testing.T) {
		dir := t.TempDir()
		cfgPath := filepath.Join(dir, "config.yaml")
		content := []byte("claude:\n  api_key: ''\nserver:\n  port: 0\n")
		require.NoError(t, os.WriteFile(cfgPath, content, 0644))
		_, err := config.Load(cfgPath)
		require.Error(t, err)
	})
}

func TestConfig_Load_EnvOverrides(t *testing.T) {
	t.Run("ANTHROPIC_API_KEY sets claude api key", func(t *testing.T) {
		t.Setenv("ANTHROPIC_API_KEY", "sk-ant-from-env")
		cfg, err := config.Load("")
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if cfg.Claude.APIKey != "sk-ant-from-env" {
			t.Errorf("expected sk-ant-from-env, got %s", cfg.Claude.APIKey)
		}
	})

	t.Run("TELEMETRYFLOW_MCP_CLAUDE_API_KEY overrides ANTHROPIC_API_KEY", func(t *testing.T) {
		t.Setenv("ANTHROPIC_API_KEY", "sk-first")
		t.Setenv("TELEMETRYFLOW_MCP_CLAUDE_API_KEY", "sk-second")
		cfg, err := config.Load("")
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if cfg.Claude.APIKey != "sk-second" {
			t.Errorf("expected sk-second, got %s", cfg.Claude.APIKey)
		}
	})

	t.Run("TELEMETRYFLOW_MCP_POSTGRES_URL sets database url via binding", func(t *testing.T) {
		t.Setenv("ANTHROPIC_API_KEY", "sk-test")
		t.Setenv("TELEMETRYFLOW_MCP_POSTGRES_URL", "postgres://localhost/testdb")
		cfg, err := config.Load("")
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if cfg.Database.URL != "postgres://localhost/testdb" {
			t.Errorf("expected postgres URL, got %s", cfg.Database.URL)
		}
	})

	t.Run("TELEMETRYFLOW_MCP_CLICKHOUSE_URL sets clickhouse url via binding", func(t *testing.T) {
		t.Setenv("ANTHROPIC_API_KEY", "sk-test")
		t.Setenv("TELEMETRYFLOW_MCP_CLICKHOUSE_URL", "clickhouse://localhost/analytics")
		cfg, err := config.Load("")
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if cfg.Clickhouse.URL != "clickhouse://localhost/analytics" {
			t.Errorf("expected clickhouse URL, got %s", cfg.Clickhouse.URL)
		}
	})

	t.Run("config file values can be set", func(t *testing.T) {
		dir := t.TempDir()
		cfgPath := filepath.Join(dir, "config.yaml")
		content := []byte("claude:\n  api_key: sk-from-file\nserver:\n  port: 9090\n  transport: sse\n")
		if err := os.WriteFile(cfgPath, content, 0644); err != nil {
			t.Fatalf("write: %v", err)
		}
		cfg, err := config.Load(cfgPath)
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if cfg.Claude.APIKey != "sk-from-file" {
			t.Errorf("expected sk-from-file, got %s", cfg.Claude.APIKey)
		}
		if cfg.Server.Port != 9090 {
			t.Errorf("expected 9090, got %d", cfg.Server.Port)
		}
		if cfg.Server.Transport != "sse" {
			t.Errorf("expected sse, got %s", cfg.Server.Transport)
		}
	})

	t.Run("bad yaml syntax returns error", func(t *testing.T) {
		dir := t.TempDir()
		cfgPath := filepath.Join(dir, "config.yaml")
		content := []byte("server:\n  port: [invalid\n")
		if err := os.WriteFile(cfgPath, content, 0644); err != nil {
			t.Fatalf("write: %v", err)
		}
		_, err := config.Load(cfgPath)
		if err == nil {
			t.Error("expected error for invalid YAML")
		}
	})

	t.Run("env override for server host and port", func(t *testing.T) {
		t.Setenv("ANTHROPIC_API_KEY", "sk-test")
		t.Setenv("TELEMETRYFLOW_MCP_SERVER_HOST", "0.0.0.0")
		t.Setenv("TELEMETRYFLOW_MCP_SERVER_PORT", "3000")
		cfg, err := config.Load("")
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if cfg.Server.Host != "0.0.0.0" {
			t.Errorf("expected 0.0.0.0, got %s", cfg.Server.Host)
		}
		if cfg.Server.Port != 3000 {
			t.Errorf("expected 3000, got %d", cfg.Server.Port)
		}
	})

	t.Run("env override for logging level", func(t *testing.T) {
		t.Setenv("ANTHROPIC_API_KEY", "sk-test")
		t.Setenv("TELEMETRYFLOW_MCP_LOG_LEVEL", "debug")
		cfg, err := config.Load("")
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if cfg.Logging.Level != "debug" {
			t.Errorf("expected debug, got %s", cfg.Logging.Level)
		}
	})

	t.Run("env override for telemetry", func(t *testing.T) {
		t.Setenv("ANTHROPIC_API_KEY", "sk-test")
		t.Setenv("TELEMETRYFLOW_MCP_TELEMETRY_ENABLED", "false")
		cfg, err := config.Load("")
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if cfg.Telemetry.Enabled {
			t.Error("expected telemetry disabled")
		}
	})
}

func TestConfig_Validate_InvalidPortHigh(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Claude.APIKey = "test-key"
	cfg.Server.Port = 70000
	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for port > 65535")
	}
}

func TestConfig_Validate_InvalidTemperatureLow(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Claude.APIKey = "test-key"
	cfg.Claude.Temperature = -0.5
	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for negative temperature")
	}
}

func TestConfig_Validate_InvalidTraceSampleRateLow(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Claude.APIKey = "test-key"
	cfg.Telemetry.TraceSampleRate = -0.1
	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for negative trace sample rate")
	}
}
