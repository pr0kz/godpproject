package config

import (
	"strings"
	"testing"
)

func validConfig() Config {
	return Config{Environment: "development", Port: "8080", DBUser: "app", DBPassword: "password", KafkaBrokers: []string{"kafka:9092"}, JWTSecret: strings.Repeat("s", 32)}
}

func TestValidateAcceptsDevelopmentConfig(t *testing.T) {
	if err := validConfig().Validate(); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}
}

func TestValidateRequiresSecrets(t *testing.T) {
	cfg := validConfig()
	cfg.DBPassword = ""
	cfg.JWTSecret = "short"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected missing secret validation error")
	}
}

func TestValidateRejectsProductionRootUser(t *testing.T) {
	cfg := validConfig()
	cfg.Environment = "production"
	cfg.DBUser = "root"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected root user to be rejected in production")
	}
}

func TestValidateRejectsInvalidPort(t *testing.T) {
	cfg := validConfig()
	cfg.Port = "70000"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid port error")
	}
}
