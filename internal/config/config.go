package config

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
)

type Config struct {
	Env        string
	LogLevel   slog.Level
	LogHandler string
	TunIP      string
	TunRoute   string
}

func (c *Config) isValid() error {
	if c.LogHandler != "text" && c.LogHandler != "json" {
		return fmt.Errorf("invalid logHandler: %s", c.LogHandler)
	}

	if _, _, err := net.ParseCIDR(c.TunIP); err != nil {
		return fmt.Errorf("tunIP must be in CIDR notation: %w", err)
	}

	if _, _, err := net.ParseCIDR(c.TunRoute); err != nil {
		return fmt.Errorf("tunRoute must be in CIDR notation: %w", err)
	}

	return nil
}

func Parse() (*Config, error) {
	var cfg Config
	var logLevelStr string

	flag.StringVar(&cfg.Env, "env", "dev", "Environment (dev, prod)")
	flag.StringVar(&cfg.TunIP, "tunIP", "10.0.0.1/24", "TUN interface IP")
	flag.StringVar(&logLevelStr, "logLevel", "debug", "Log level (debug, info, warn, error)")
	flag.StringVar(&cfg.LogHandler, "logHandler", "text", "Log handler format (text, json)")
	flag.StringVar(&cfg.TunRoute, "tunRoute", "10.0.0.0/24", "TUN interface route")

	flag.Parse()

	switch strings.ToLower(logLevelStr) {
	case "debug":
		cfg.LogLevel = slog.LevelDebug
	case "info":
		cfg.LogLevel = slog.LevelInfo
	case "warn":
		cfg.LogLevel = slog.LevelWarn
	case "error":
		cfg.LogLevel = slog.LevelError
	default:
		return nil, fmt.Errorf("invalid logLevel: %s", logLevelStr)
	}

	if err := cfg.isValid(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	return &cfg, nil
}

func Usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

func SetupLogger(w io.Writer, level slog.Level, env, format string) *slog.Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: level,
	}

	switch format {
	case "json":
		handler = slog.NewJSONHandler(w, opts)
	default:
		handler = slog.NewTextHandler(w, opts)
	}

	return slog.New(handler).With(slog.String("env", env))
}
