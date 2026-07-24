// Package config reads runtime configuration from the environment.
package config

import "os"

type Config struct {
	Env         string
	Port        string
	DatabaseURL string
	JWTSecret   string
}

func Load() Config {
	return Config{
		Env:         getenv("APP_ENV", "development"),
		Port:        getenv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
	}
}

func (c Config) IsProduction() bool { return c.Env == "production" }

// MissingProductionSecrets names the secrets production requires but does not have.
// Failing at boot is better than serving traffic with an insecure default.
func (c Config) MissingProductionSecrets() []string {
	if !c.IsProduction() {
		return nil
	}
	var missing []string
	if c.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if c.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	return missing
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
