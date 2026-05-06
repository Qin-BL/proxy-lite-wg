package config

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppAddr                string
	DBPath                 string
	AdminUsername          string
	AdminPassword          string
	AdminPasswordHash      string
	AdminSessionSecret     string
	AdminSessionSecretPath string
	AdminSessionTTL        time.Duration
	PublicHost             string
	PublicPort             int
	VLESSWSPath            string
	XrayConfigPath         string
	XrayContainerName      string
	XrayInboundPort        int
	XrayLogLevel           string
	DockerSocketPath       string
}

func Load() (Config, error) {
	_ = loadDotEnv(".env")

	cfg := Config{
		AppAddr:                envOrDefault("APP_ADDR", "127.0.0.1:8080"),
		DBPath:                 envOrDefault("APP_DB_PATH", "./data/proxy-lite-control.db"),
		AdminUsername:          strings.TrimSpace(os.Getenv("ADMIN_USERNAME")),
		AdminPassword:          os.Getenv("ADMIN_PASSWORD"),
		AdminPasswordHash:      strings.TrimSpace(os.Getenv("ADMIN_PASSWORD_HASH")),
		AdminSessionSecret:     strings.TrimSpace(os.Getenv("ADMIN_SESSION_SECRET")),
		PublicHost:             strings.TrimSpace(os.Getenv("PUBLIC_HOST")),
		PublicPort:             envOrDefaultInt("PUBLIC_PORT", 443),
		VLESSWSPath:            envOrDefault("VLESS_WS_PATH", "/proxy-lite-ws"),
		XrayConfigPath:         envOrDefault("XRAY_CONFIG_PATH", "./runtime/xray-config.json"),
		XrayContainerName:      strings.TrimSpace(os.Getenv("XRAY_CONTAINER_NAME")),
		XrayInboundPort:        envOrDefaultInt("XRAY_INBOUND_PORT", 10000),
		XrayLogLevel:           envOrDefault("XRAY_LOG_LEVEL", "warning"),
		DockerSocketPath:       envOrDefault("DOCKER_SOCKET_PATH", "/var/run/docker.sock"),
		AdminSessionSecretPath: envOrDefault("ADMIN_SESSION_SECRET_PATH", filepath.Join(filepath.Dir(envOrDefault("APP_DB_PATH", "./data/proxy-lite-control.db")), "session_secret")),
		AdminSessionTTL:        time.Duration(envOrDefaultInt("ADMIN_SESSION_TTL_HOURS", 12)) * time.Hour,
	}

	if cfg.AdminUsername == "" {
		return Config{}, errors.New("ADMIN_USERNAME is required")
	}
	if cfg.AdminPassword == "" && cfg.AdminPasswordHash == "" {
		return Config{}, errors.New("either ADMIN_PASSWORD or ADMIN_PASSWORD_HASH is required")
	}
	if cfg.PublicHost == "" {
		return Config{}, errors.New("PUBLIC_HOST is required")
	}
	if !strings.HasPrefix(cfg.VLESSWSPath, "/") {
		return Config{}, errors.New("VLESS_WS_PATH must start with /")
	}
	if cfg.XrayConfigPath == "" {
		return Config{}, errors.New("XRAY_CONFIG_PATH is required")
	}
	if cfg.PublicPort <= 0 {
		return Config{}, errors.New("PUBLIC_PORT must be positive")
	}
	if cfg.XrayInboundPort <= 0 {
		return Config{}, errors.New("XRAY_INBOUND_PORT must be positive")
	}
	if cfg.AdminSessionTTL <= 0 {
		return Config{}, errors.New("ADMIN_SESSION_TTL_HOURS must be positive")
	}

	return cfg, nil
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, value)
	}

	return scanner.Err()
}

func (c Config) DataDir() string {
	return filepath.Dir(c.DBPath)
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envOrDefaultInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
