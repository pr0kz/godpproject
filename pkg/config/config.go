package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Role, Port, DBUser, DBPassword, DBHost, DBPort, DBName, RedisHost, RedisPort, RedisPassword, JWTSecret string
	KafkaBrokers                                                                                           []string
}

func Load(role string) (Config, error) {
	prefix := strings.ToUpper(role)
	get := func(name, fallback string) string {
		if v := strings.TrimSpace(os.Getenv(prefix + "_" + name)); v != "" {
			return v
		}
		if v := strings.TrimSpace(os.Getenv(name)); v != "" {
			return v
		}
		return fallback
	}
	c := Config{Role: role, Port: get("PORT", "8080"), DBUser: get("DB_USER", ""), DBPassword: get("DB_PASSWORD", ""), DBHost: get("DB_HOST", "localhost"), DBPort: get("DB_PORT", "3306"), DBName: get("DB_NAME", role+"_db"), RedisHost: get("REDIS_HOST", ""), RedisPort: get("REDIS_PORT", "6379"), RedisPassword: get("REDIS_PASSWORD", ""), JWTSecret: get("JWT_SECRET", ""), KafkaBrokers: csv(get("KAFKA_BROKERS", ""))}
	return c, c.Validate()
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
	if p, err := strconv.Atoi(c.Port); err != nil || p < 1 || p > 65535 {
		errs = append(errs, errors.New("PORT must be between 1 and 65535"))
	}
	if (c.Role == "shop" || c.Role == "order") && c.RedisHost == "" {
		errs = append(errs, errors.New("REDIS_HOST is required"))
	}
	if (c.Role == "review" || c.Role == "order" || c.Role == "shop") && len(c.KafkaBrokers) == 0 {
		errs = append(errs, errors.New("KAFKA_BROKERS is required"))
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s configuration: %w", c.Role, errors.Join(errs...))
	}
	return nil
}
func csv(v string) []string {
	var out []string
	for _, x := range strings.Split(v, ",") {
		if x = strings.TrimSpace(x); x != "" {
			out = append(out, x)
		}
	}
	return out
}
