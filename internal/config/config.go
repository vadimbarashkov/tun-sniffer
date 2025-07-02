package config

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	flag.Usage = Usage
}

type Config struct {
	TunIP    string
	TunRoute string
}

func (c *Config) isValid() error {
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

	flag.StringVar(&cfg.TunIP, "tunIP", "10.0.0.1/24", "TUN interface IP")
	flag.StringVar(&cfg.TunRoute, "tunRoute", "10.0.0.0/24", "TUN interface route")
	flag.Parse()

	if err := cfg.isValid(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	return &cfg, nil
}

func Usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

func SetupLogger(w io.Writer, level zerolog.Level) {
	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(w).With().Timestamp().Logger()
	log.Logger = logger
}
