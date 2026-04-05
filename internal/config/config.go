package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Config holds all application configuration sourced from environment variables.
type Config struct {
	Port         int
	Host         string
	StaticDir    string
	TemplateDir  string
	UploadDir    string
	MaxUploadMB  int64
	Environment  string
	LogLevel     string
	CORSOrigins  []string
	ReadTimeout  int
	WriteTimeout int
}

// Default values for configuration.
const (
	defaultPort         = 8080
	defaultHost         = "0.0.0.0"
	defaultStaticDir    = "web/static"
	defaultTemplateDir  = "web/templates"
	defaultUploadDir    = "uploads"
	defaultMaxUploadMB  = 10
	defaultEnvironment  = "development"
	defaultLogLevel     = "info"
	defaultReadTimeout  = 15
	defaultWriteTimeout = 15
)

// Load reads configuration from environment variables with sensible defaults.
// Relative paths for StaticDir, TemplateDir, and UploadDir are resolved against
// the executable's directory when they don't exist relative to the working directory.
// This ensures the server works on platforms like Vercel where the cwd may differ
// from the project root.
func Load() (*Config, error) {
	cfg := &Config{
		Port:         envInt("PORT", defaultPort),
		Host:         envStr("HOST", defaultHost),
		StaticDir:    envStr("STATIC_DIR", defaultStaticDir),
		TemplateDir:  envStr("TEMPLATE_DIR", defaultTemplateDir),
		UploadDir:    envStr("UPLOAD_DIR", defaultUploadDir),
		MaxUploadMB:  int64(envInt("MAX_UPLOAD_MB", defaultMaxUploadMB)),
		Environment:  envStr("ENVIRONMENT", defaultEnvironment),
		LogLevel:     envStr("LOG_LEVEL", defaultLogLevel),
		ReadTimeout:  envInt("READ_TIMEOUT", defaultReadTimeout),
		WriteTimeout: envInt("WRITE_TIMEOUT", defaultWriteTimeout),
	}

	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", cfg.Port)
	}

	cfg.StaticDir = resolveDir(cfg.StaticDir)
	cfg.TemplateDir = resolveDir(cfg.TemplateDir)
	cfg.UploadDir = resolveDir(cfg.UploadDir)

	return cfg, nil
}

// resolveDir returns the path as-is if it exists relative to the cwd.
// Otherwise it tries resolving relative to the executable's directory,
// which handles deployment platforms that run the binary from a different cwd.
func resolveDir(dir string) string {
	if filepath.IsAbs(dir) {
		return dir
	}
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return dir
	}
	exe, err := os.Executable()
	if err != nil {
		return dir
	}
	candidate := filepath.Join(filepath.Dir(exe), dir)
	if info, err := os.Stat(candidate); err == nil && info.IsDir() {
		return candidate
	}
	return dir
}

// Addr returns the listen address string.
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// IsProd returns true when running in production.
func (c *Config) IsProd() bool {
	return c.Environment == "production"
}

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return fallback
		}
		return n
	}
	return fallback
}
