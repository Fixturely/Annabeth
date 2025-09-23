package config

func GetTestConfig(cfg *Config) {
	cfg.Environment = "test"
	// Allow overriding via environment variables; fall back to reasonable test defaults
	cfg.Database.Host = getEnvOrDefault("DATABASE_HOST", cfg.Database.Host)
	cfg.Database.Port = getEnvOrDefault("DATABASE_PORT", cfg.Database.Port)
	cfg.Database.User = getEnvOrDefault("DATABASE_USER", cfg.Database.User)
	cfg.Database.Password = getEnvOrDefault("DATABASE_PASSWORD", cfg.Database.Password)
	cfg.Database.Name = getEnvOrDefault("DATABASE_NAME", cfg.Database.Name)
	cfg.Database.SSLMode = getEnvOrDefault("DATABASE_SSL_MODE", false)
}
