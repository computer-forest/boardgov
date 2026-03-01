package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Storage  StorageConfig
	Security SecurityConfig
	Log      LogConfig
}

type ServerConfig struct {
	Host            string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	TrustedProxies  []string
}

type DatabaseConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
	Issuer        string
}

type StorageConfig struct {
	BasePath    string
	MaxFileSize int64
	AllowedMIME []string
}

type SecurityConfig struct {
	IPWhitelist        []string
	RateLimitRequests  int
	RateLimitPeriod    time.Duration
	RateLimitRPS       int
	RateLimitBurst     int
	CSRFSecret         string
	EncryptionKey      string
	CORSAllowedOrigins []string
}

type LogConfig struct {
	Level  string
	Format string
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Port:            getEnv("SERVER_PORT", "8080"),
			ReadTimeout:     getDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:     getDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: getDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
			TrustedProxies:  getList("SERVER_TRUSTED_PROXIES", "127.0.0.1"),
		},
		Database: DatabaseConfig{
			DSN:             mustEnv("DATABASE_DSN"),
			MaxOpenConns:    getInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getDuration("DB_CONN_MAX_IDLE_TIME", time.Minute),
		},
		JWT: JWTConfig{
			AccessSecret:  mustEnv("JWT_ACCESS_SECRET"),
			RefreshSecret: mustEnv("JWT_REFRESH_SECRET"),
			AccessExpiry:  getDuration("JWT_ACCESS_EXPIRY", 15*time.Minute),
			RefreshExpiry: getDuration("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
			Issuer:        getEnv("JWT_ISSUER", "boardgov"),
		},
		Storage: StorageConfig{
			BasePath:    getEnv("STORAGE_BASE_PATH", "/secure-storage"),
			MaxFileSize: int64(getInt("STORAGE_MAX_FILE_SIZE_MB", 25)) * 1024 * 1024,
			AllowedMIME: getList("STORAGE_ALLOWED_MIME", "application/pdf,application/vnd.openxmlformats-officedocument.wordprocessingml.document"),
		},
		Security: SecurityConfig{
			IPWhitelist:        getList("SECURITY_IP_WHITELIST", ""),
			RateLimitRequests:  getInt("RATE_LIMIT_REQUESTS", 100),
			RateLimitPeriod:    getDuration("RATE_LIMIT_PERIOD", time.Minute),
			RateLimitRPS:       getInt("RATE_LIMIT_RPS", 20),
			RateLimitBurst:     getInt("RATE_LIMIT_BURST", 40),
			CSRFSecret:         mustEnv("CSRF_SECRET"),
			EncryptionKey:      mustEnv("ENCRYPTION_KEY"),
			CORSAllowedOrigins: getList("CORS_ALLOWED_ORIGINS", "http://localhost:4200"),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}
	return cfg, nil
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required env var %s is not set", key))
	}
	return v
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func getList(key, def string) []string {
	v := getEnv(key, def)
	if v == "" {
		return []string{}
	}
	parts := strings.Split(v, ",")
	var result []string
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}
