package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
)

type Config struct {
	Port           string
	OpenGolfAPIKey string
}

func Load() Config {
	return Config{
		Port:           getEnv("PORT", "8080"),
		OpenGolfAPIKey: getEnv("OPENGOLF_API_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func GetLogger() *slog.Logger {
	env := os.Getenv("APP_ENV")
	addSource := os.Getenv("LOG_ADD_SOURCE") == "true"
	levelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))

	// Parse log level
	var level slog.Level
	switch levelStr {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Production → JSON logs
	if env == "production" {
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: addSource,
			Level:     level,
		}))
	}

	// Development → pretty logs (tint)
	return slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level:     level,
		AddSource: addSource,
		NoColor:   false,
	}))
}
