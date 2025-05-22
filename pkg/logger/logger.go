package logger

import (
	"ai-job/pkg/config"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type Config struct {
	Level   string     `yaml:"level"`
	Format  string     `yaml:"format"`
	Outputs []string   `yaml:"outputs"`
	File    FileConfig `yaml:"file"`
	Loki    LokiConfig `yaml:"loki"`
}

type FileConfig struct {
	Path       string `yaml:"path"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
}

type LokiConfig struct {
	Enabled bool              `yaml:"enabled"`
	URL     string            `yaml:"url"`
	Labels  map[string]string `yaml:"labels"`
}

// ConvertConfig converts a config.LoggingConfig to logger.Config
func ConvertConfig(cfg config.LoggingConfig) Config {
	return Config{
		Level:   cfg.Level,
		Format:  cfg.Format,
		Outputs: []string{cfg.Output},
		File: FileConfig{
			Path:       "/var/log/app/ai-job.log",
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     7,
		},
		Loki: LokiConfig{
			Enabled: false,
		},
	}
}

func NewLogger(cfg Config) (*logrus.Logger, error) {
	log := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}
	log.SetLevel(level)

	// Set log format
	if cfg.Format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{})
	} else {
		log.SetFormatter(&logrus.TextFormatter{})
	}

	// Configure outputs
	var outputs []io.Writer
	for _, out := range cfg.Outputs {
		switch out {
		case "stdout":
			outputs = append(outputs, os.Stdout)
		case "file":
			if err := os.MkdirAll(filepath.Dir(cfg.File.Path), 0755); err != nil {
				return nil, err
			}
			fileOutput, err := NewFileOutput(cfg.File)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, fileOutput)
		}
	}

	if len(outputs) > 0 {
		log.SetOutput(io.MultiWriter(outputs...))
	}

	// Add Loki hook if enabled
	if cfg.Loki.Enabled {
		lokiHook := NewLokiHook(
			cfg.Loki.URL,
			cfg.Loki.Labels,
			10, // batch size
		)
		log.AddHook(lokiHook)
	}

	return log, nil
}

func NewFileOutput(cfg FileConfig) (io.Writer, error) {
	return os.OpenFile(cfg.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}
