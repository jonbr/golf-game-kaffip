package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
)

type Config struct {
	Port               string
	OpenGolfAPIKey     string
	CORSAllowedOrigins []string
}

func Load() Config {
	return Config{
		Port:               getEnv("PORT", "8080"),
		OpenGolfAPIKey:     getEnv("OPENGOLF_API_KEY", ""),
		CORSAllowedOrigins: getEnvList("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getEnvList reads a comma-separated env var into a string slice, e.g.
// CORS_ALLOWED_ORIGINS="http://localhost:5173,https://kaffip.app".
// Use "*" to allow any origin.
func getEnvList(key string, fallback []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
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
