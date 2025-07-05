package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Config struct {
	LogLevel      zerolog.Level
	TunIP         string
	TunRoute      string
	MaxGoroutines int
}

type logLevelValue struct {
	Level *zerolog.Level
}

func (l *logLevelValue) String() string {
	if l.Level == nil {
		return ""
	}
	return l.Level.String()
}

func (l *logLevelValue) Set(s string) error {
	s = strings.ToLower(s)
	level, err := zerolog.ParseLevel(s)
	if err != nil {
		return fmt.Errorf("invalid log level: %s", err)
	}
	*l.Level = level
	return nil
}

func (c *Config) validate() error {
	var errs []error

	if _, _, err := net.ParseCIDR(c.TunIP); err != nil {
		errs = append(errs, fmt.Errorf("tunIP must be CIDR: %w", err))
	}
	if _, _, err := net.ParseCIDR(c.TunRoute); err != nil {
		errs = append(errs, fmt.Errorf("tunRoute must be CIDR: %w", err))
	}
	if c.MaxGoroutines <= 0 {
		errs = append(errs, fmt.Errorf("maxGoroutines must be positive: %d", c.MaxGoroutines))
	}

	return errors.Join(errs...)
}

func MustParse() *Config {
	var cfg Config

	flag.Var(&logLevelValue{Level: &cfg.LogLevel}, "logLevel", "Log level (trace|debug|info|warn|error|fatal|panic|disabled)")
	flag.StringVar(&cfg.TunIP, "tunIP", "10.0.0.1/24", "CIDR for TUN interface IP")
	flag.StringVar(&cfg.TunRoute, "tunRoute", "10.0.0.0/24", "CIRD for routing via TUN interface")
	flag.IntVar(&cfg.MaxGoroutines, "maxGoroutines", 100, "Maximum number of goroutines")
	flag.Parse()

	if err := cfg.validate(); err != nil {
		log.Fatal().
			Err(err).
			Msg("invalid configuration")
	}

	return &cfg
}

func SetupLogger(w io.Writer, level zerolog.Level) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.New(w).Level(level).With().Timestamp().Logger()
}
