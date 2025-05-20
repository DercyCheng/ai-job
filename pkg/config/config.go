package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Worker   WorkerConfig   `yaml:"worker"`
	LLM      LLMConfig      `yaml:"llm"`
	Queue    QueueConfig    `yaml:"queue"`
	Logging  LoggingConfig  `yaml:"logging"`
	MCP      MCPConfig      `yaml:"mcp"`
}

// ServerConfig represents the API server configuration
type ServerConfig struct {
	Host           string        `yaml:"host"`
	Port           string        `yaml:"port"`
	Timeout        time.Duration `yaml:"timeout"`
	MaxRequestSize int64         `yaml:"max_request_size"`
}

// DatabaseConfig represents the database configuration
type DatabaseConfig struct {
	Driver          string        `yaml:"driver"`
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	Name            string        `yaml:"name"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

// WorkerConfig represents the worker configuration
type WorkerConfig struct {
	MaxWorkers        int           `yaml:"max_workers"`
	TaskTimeout       time.Duration `yaml:"task_timeout"`
	PollInterval      time.Duration `yaml:"poll_interval"`
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval"`
}

// LLMConfig represents the LLM configuration
type LLMConfig struct {
	Models []ModelConfig `yaml:"models"`
}

// ModelConfig represents an LLM model configuration
type ModelConfig struct {
	Name             string `yaml:"name"`
	Provider         string `yaml:"provider"`
	ModelPath        string `yaml:"model_path"`
	MaxContextLength int    `yaml:"max_context_length"`
	Quantization     string `yaml:"quantization"`
}

// QueueConfig represents the queue configuration
type QueueConfig struct {
	Driver   string        `yaml:"driver"`
	Address  string        `yaml:"address"`
	Password string        `yaml:"password"`
	MaxRetry int           `yaml:"max_retry"`
	JobTTL   time.Duration `yaml:"job_ttl"`
}

// LoggingConfig represents the logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

// MCPConfig represents the Model Context Protocol configuration
type MCPConfig struct {
	Enabled     bool          `yaml:"enabled"`
	ServerURL   string        `yaml:"server_url"`
	APIVersion  string        `yaml:"api_version"`
	MaxContexts int           `yaml:"max_contexts"`
	Timeout     time.Duration `yaml:"timeout"`
}

// Load loads the configuration from a file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply environment variable overrides if needed
	applyEnvironmentOverrides(&cfg)

	return &cfg, nil
}

// applyEnvironmentOverrides applies environment variable overrides to the configuration
func applyEnvironmentOverrides(cfg *Config) {
	// Database overrides
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		cfg.Database.Host = dbHost
	}
	if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
		fmt.Sscanf(dbPort, "%d", &cfg.Database.Port)
	}
	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		cfg.Database.User = dbUser
	}
	if dbPass := os.Getenv("DB_PASSWORD"); dbPass != "" {
		cfg.Database.Password = dbPass
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		cfg.Database.Name = dbName
	}

	// Server overrides
	if serverPort := os.Getenv("SERVER_PORT"); serverPort != "" {
		cfg.Server.Port = serverPort
	}
}
