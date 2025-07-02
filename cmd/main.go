//go:build linux

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"

	"github.com/vadimbarashkov/tun-sniffer/internal/config"
	"github.com/vadimbarashkov/tun-sniffer/internal/tun"
)

func main() {
	cfg := config.MustParse()
	config.SetupLogger(os.Stdout, cfg.LogLevel)

	ifce, err := tun.Setup()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to setup TUN interface")
	}
	defer func() {
		if err := ifce.Close(); err != nil {
			log.Error().
				Err(err).
				Msg("failed to close TUN interface")
		}
	}()

	if err := tun.Configure(ifce.Name(), cfg.TunIP, cfg.TunRoute); err != nil {
		log.Fatal().
			Str("name", ifce.Name()).
			Str("tun_ip", cfg.TunIP).
			Str("tun_route", cfg.TunRoute).
			Err(err).
			Msg("failed to configure TUN interface")
	}

	log.Info().
		Str("name", ifce.Name()).
		Str("tun_ip", cfg.TunIP).
		Str("tun_route", cfg.TunRoute).
		Msg("TUN interface configured")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGTERM)
	defer stop()

	tun.ReadPackets(ctx, ifce)
}
