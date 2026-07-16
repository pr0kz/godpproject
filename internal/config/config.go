package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Environment  string
	ServiceRole  string
	Port         string
	DBUser       string
	DBPassword   string
	DBHost       string
	DBPort       string
	DBName       string
	RedisHost    string
	RedisPort    string
	KafkaBrokers []string
	JWTSecret    string
}

func Load() (Config, error) {
	cfg := Config{
		Environment:  value("APP_ENV", "development"),
		ServiceRole:  strings.ToLower(strings.TrimSpace(os.Getenv("SERVICE_ROLE"))),
		Port:         value("PORT", "8080"),
		DBUser:       strings.TrimSpace(os.Getenv("DB_USER")),
		DBPassword:   os.Getenv("DB_PASSWORD"),
		DBHost:       value("DB_HOST", "localhost"),
		DBPort:       value("DB_PORT", "3306"),
		DBName:       value("DB_NAME", "ai_review_system"),
		RedisHost:    value("REDIS_HOST", "localhost"),
		RedisPort:    value("REDIS_PORT", "6379"),
		KafkaBrokers: splitCSV(value("KAFKA_BROKERS", "localhost:9092")),
		JWTSecret:    os.Getenv("JWT_SECRET"),
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	var errs []error
	if c.DBUser == "" {
		errs = append(errs, errors.New("DB_USER is required"))
	}
	if c.DBPassword == "" {
		errs = append(errs, errors.New("DB_PASSWORD is required"))
	}
	if len(c.JWTSecret) < 32 {
		errs = append(errs, errors.New("JWT_SECRET must contain at least 32 characters"))
	}
	port, err := strconv.Atoi(c.Port)
	if err != nil || port < 1 || port > 65535 {
		errs = append(errs, errors.New("PORT must be between 1 and 65535"))
	}
	if len(c.KafkaBrokers) == 0 {
		errs = append(errs, errors.New("KAFKA_BROKERS is required"))
	}
	if strings.EqualFold(c.Environment, "production") {
		if strings.EqualFold(c.DBUser, "root") {
			errs = append(errs, errors.New("DB_USER must not be root in production"))
		}
		if c.DBPassword == "root" {
			errs = append(errs, errors.New("default DB_PASSWORD is forbidden in production"))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("configuration error: %w", errors.Join(errs...))
	}
	return nil
}

func value(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func splitCSV(value string) []string {
	var result []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}
