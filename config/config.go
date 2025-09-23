package config

import (
	"os"
	"strconv"
	"strings"
	"sync"
)

type Config struct {
	Database    DatabaseConfig
	GoogleAuth  GoogleAuthConfig
	Frontend    FrontendConfig
	Environment string
	AwsEndpoint string
	ServerPort  string
	MikeAPIKey  string
}

type FrontendConfig struct {
	URL string
}

type GoogleAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  bool
}

var (
	config *Config
	once   sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		cfg := &Config{
			Environment: "development",
			ServerPort:  "8080",
			Database: DatabaseConfig{
				Host:     getEnvOrDefault("DATABASE_HOST", "localhost"),
				User:     getEnvOrDefault("DATABASE_USER", "admin"),
				Password: getEnvOrDefault("DATABASE_PASSWORD", "admin"),
				Name:     getEnvOrDefault("DATABASE_NAME", "annabeth-local-db"),
				Port:     getEnvOrDefault("DATABASE_PORT", 5413),
				SSLMode:  getEnvOrDefault("DATABASE_SSL_MODE", false),
			},
			Frontend: FrontendConfig{
				URL: getEnvOrDefault("FRONTEND_URL", "http://localhost:3000"),
			},
			MikeAPIKey: getEnvOrDefault("MIKE_API_KEY", "http://localhost:9000/v1"),
		}

		switch strings.ToLower(os.Getenv("ENV")) {
		case "test":
			GetTestConfig(cfg)
		default:
			GetDevelopmentConfig(cfg)
		}

		config = cfg
	})

	return config
}

func getEnvOrDefault[T string | int | bool](envVarName string, defaultVal T) T {
	val := os.Getenv(envVarName)
	if val == "" {
		return defaultVal
	}
	switch any(defaultVal).(type) {
	case string:
		return any(val).(T)
	case int:
		i, _ := strconv.Atoi(val)
		// don't error check cause we WANT it to blow up if it's not a parseable int
		return any(i).(T)
	case bool:
		b, _ := strconv.ParseBool(val)
		// don't error check cause we WANT it to blow up if it's not a parseable bool
		return any(b).(T)
	}
	return defaultVal
}
