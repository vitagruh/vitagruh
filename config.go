// === FILE: config.go ===
package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LogLevel defines logging verbosity
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelError LogLevel = "error"
)

// BackendConfig represents a single ICAP backend server
type BackendConfig struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"` // host:port
	Service string `yaml:"service"` // ICAP service name (e.g., /icap/service)
}

// Config holds the entire relay configuration
type Config struct {
	ListenAddr   string        `yaml:"listen_addr"`
	BackendA     BackendConfig `yaml:"backend_a"`
	BackendB     BackendConfig `yaml:"backend_b"`
	TimeoutMs    int           `yaml:"timeout_ms"`
	Priority     string        `yaml:"priority"` // "A" or "B"
	LogLevel     LogLevel      `yaml:"log_level"`
	MaxConns     int           `yaml:"max_conns"`
	MaxBufferSize int          `yaml:"max_buffer_size"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		ListenAddr:    ":1344",
		TimeoutMs:     5000,
		Priority:      "A",
		LogLevel:      LogLevelInfo,
		MaxConns:      100,
		MaxBufferSize: 10 * 1024 * 1024, // 10MB
	}
}

// LoadConfig reads and validates configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate required fields
	if cfg.ListenAddr == "" {
		return nil, fmt.Errorf("listen_addr is required")
	}
	if cfg.BackendA.Address == "" {
		return nil, fmt.Errorf("backend_a.address is required")
	}
	if cfg.BackendB.Address == "" {
		return nil, fmt.Errorf("backend_b.address is required")
	}
	if cfg.TimeoutMs <= 0 {
		cfg.TimeoutMs = 5000
	}
	if cfg.Priority != "A" && cfg.Priority != "B" {
		cfg.Priority = "A"
	}
	if cfg.MaxConns <= 0 {
		cfg.MaxConns = 100
	}
	if cfg.MaxBufferSize <= 0 {
		cfg.MaxBufferSize = 10 * 1024 * 1024
	}

	return cfg, nil
}
