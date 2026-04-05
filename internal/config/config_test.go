package config

import (
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != defaultPort {
		t.Errorf("expected port %d, got %d", defaultPort, cfg.Port)
	}
	if cfg.Host != defaultHost {
		t.Errorf("expected host %q, got %q", defaultHost, cfg.Host)
	}
	if cfg.Environment != defaultEnvironment {
		t.Errorf("expected env %q, got %q", defaultEnvironment, cfg.Environment)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("ENVIRONMENT", "production")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
	if cfg.Environment != "production" {
		t.Errorf("expected production, got %q", cfg.Environment)
	}
}

func TestLoadInvalidPort(t *testing.T) {
	t.Setenv("PORT", "99999")

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid port")
	}
}

func TestAddr(t *testing.T) {
	cfg := &Config{Host: "localhost", Port: 3000}
	if got := cfg.Addr(); got != "localhost:3000" {
		t.Errorf("expected localhost:3000, got %q", got)
	}
}

func TestIsProd(t *testing.T) {
	cfg := &Config{Environment: "production"}
	if !cfg.IsProd() {
		t.Error("expected IsProd to return true")
	}
	cfg.Environment = "development"
	if cfg.IsProd() {
		t.Error("expected IsProd to return false")
	}
}

func TestLoadNonNumericPort(t *testing.T) {
	t.Setenv("PORT", "not-a-number")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != defaultPort {
		t.Errorf("expected default port %d for non-numeric, got %d", defaultPort, cfg.Port)
	}
}

func TestLoadAllEnvVars(t *testing.T) {
	t.Setenv("PORT", "3000")
	t.Setenv("HOST", "127.0.0.1")
	t.Setenv("STATIC_DIR", "/tmp/static")
	t.Setenv("TEMPLATE_DIR", "/tmp/templates")
	t.Setenv("UPLOAD_DIR", "/tmp/uploads")
	t.Setenv("MAX_UPLOAD_MB", "25")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("READ_TIMEOUT", "30")
	t.Setenv("WRITE_TIMEOUT", "45")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 3000 {
		t.Errorf("Port: expected 3000, got %d", cfg.Port)
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("Host: expected 127.0.0.1, got %q", cfg.Host)
	}
	if cfg.StaticDir != "/tmp/static" {
		t.Errorf("StaticDir: expected /tmp/static, got %q", cfg.StaticDir)
	}
	if cfg.TemplateDir != "/tmp/templates" {
		t.Errorf("TemplateDir: got %q", cfg.TemplateDir)
	}
	if cfg.UploadDir != "/tmp/uploads" {
		t.Errorf("UploadDir: got %q", cfg.UploadDir)
	}
	if cfg.MaxUploadMB != 25 {
		t.Errorf("MaxUploadMB: expected 25, got %d", cfg.MaxUploadMB)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel: expected debug, got %q", cfg.LogLevel)
	}
	if cfg.ReadTimeout != 30 {
		t.Errorf("ReadTimeout: expected 30, got %d", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 45 {
		t.Errorf("WriteTimeout: expected 45, got %d", cfg.WriteTimeout)
	}
}

func TestLoadPortBoundaries(t *testing.T) {
	t.Setenv("PORT", "0")
	_, err := Load()
	if err == nil {
		t.Error("expected error for port 0")
	}
}

func TestAddrFormat(t *testing.T) {
	cases := []struct {
		host string
		port int
		want string
	}{
		{"0.0.0.0", 8080, "0.0.0.0:8080"},
		{"localhost", 3000, "localhost:3000"},
		{"", 80, ":80"},
	}
	for _, tc := range cases {
		cfg := &Config{Host: tc.host, Port: tc.port}
		if got := cfg.Addr(); got != tc.want {
			t.Errorf("Addr(%q, %d) = %q, want %q", tc.host, tc.port, got, tc.want)
		}
	}
}
