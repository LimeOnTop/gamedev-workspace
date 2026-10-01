package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	DevMode   bool
	PublicURL string
	DatabaseConfig
	RedisConfig
	StorageConfig
}

type DatabaseConfig struct {
	DatabaseURL string
}

type RedisConfig struct {
	RedisURL     string
	TreeCacheTTL time.Duration
}

type StorageConfig struct {
	UploadDir     string
	MaxUploadSize int64
	// MaxModelSize limits 3D model uploads, which are much larger than references.
	MaxModelSize int64
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Port:      getEnv("PORT", "8080"),
		DevMode:   getEnvBool("DEV_MODE", false),
		PublicURL: getEnv("PUBLIC_URL", "http://localhost:3300"),
		DatabaseConfig: DatabaseConfig{
			DatabaseURL: getEnv("DATABASE_URL", "postgres://workspace:workspace@localhost:55432/workspace?sslmode=disable"),
		},
		RedisConfig: RedisConfig{
			RedisURL:     getEnv("REDIS_URL", "redis://localhost:56379/0"),
			TreeCacheTTL: getEnvDuration("TREE_CACHE_TTL", 10*time.Minute),
		},
		StorageConfig: StorageConfig{
			UploadDir:     getEnv("UPLOAD_DIR", "./uploads"),
			MaxUploadSize: getEnvInt64("MAX_UPLOAD_MB", 25) << 20,
			MaxModelSize:  getEnvInt64("MAX_MODEL_UPLOAD_MB", 100) << 20,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func getEnvInt64(key string, defaultValue int64) int64 {
	parsed, err := strconv.ParseInt(os.Getenv(key), 10, 64)
	if err != nil || parsed <= 0 {
		return defaultValue
	}
	return parsed
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	parsed, err := time.ParseDuration(os.Getenv(key))
	if err != nil || parsed <= 0 {
		return defaultValue
	}
	return parsed
}
