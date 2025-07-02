//go:build linux

package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"slices"
	"strings"
	"sync"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/songgao/water"

	"github.com/vadimbarashkov/tun-sniffer/internal/config"
)

const (
	bufferSize = 2000
)

func main() {
	config.SetupLogger(os.Stdout, zerolog.InfoLevel)

	cfg, err := config.Parse()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("parse config")
	}

	ifce, err := setupTun()
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

	if err := configureTunInterface(ifce.Name(), cfg.TunIP, cfg.TunRoute); err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to configure TUN interface")
	}

	log.Info().
		Str("name", ifce.Name()).
		Msg("TUN interface configured")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	readPackets(ctx, ifce)
}

func setupTun() (*water.Interface, error) {
	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create TUN interface: %w", err)
	}
	return ifce, nil
}

func configureTunInterface(name, tunIP, tunRoute string) error {
	cmd := exec.Command("ip", "addr", "add", tunIP, "dev", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add IP address: %w", err)
	}

	cmd = exec.Command("ip", "link", "set", name, "up")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring interface up: %w", err)
	}

	out, err := exec.Command("ip", "route", "add", tunRoute, "via", strings.Split(tunIP, "/")[0], "dev", name).CombinedOutput()
	if err != nil {
		if !bytes.Contains(out, []byte("File exists")) {
			return fmt.Errorf("failed to add route: %w", err)
		}
	}

	return nil
}

func readPackets(ctx context.Context, ifce *water.Interface) {
	buff := make([]byte, bufferSize)
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return
		default:
			n, err := ifce.Read(buff)
			if err != nil {
				log.Error().
					Err(err).
					Msg("failed to read packet")
				continue
			}

			wg.Add(1)
			go func(data []byte) {
				defer wg.Done()
				packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)
				processPacket(packet)
			}(slices.Clone(buff[:n]))
		}
	}
}

func processPacket(packet gopacket.Packet) {
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		log.Debug().
			Any("layers", packet.Layers()).
			Msg("non-IP packet recieved")
		return
	}

	ip, ok := ipLayer.(*layers.IPv4)
	if !ok {
		log.Warn().
			Msg("failed to cast to IPv4 layer")
		return
	}

	switch ip.Protocol {
	case layers.IPProtocolTCP:
		processTCPLayer(packet, ip)
	case layers.IPProtocolUDP:
		processUDPLayer(packet, ip)
	default:
		log.Debug().
			Str("protocol", ip.Protocol.String()).
			Msg("non-TCP/UDP packet recieved")
	}
}

func processTCPLayer(packet gopacket.Packet, ip *layers.IPv4) {
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		log.Debug().
			Any("layers", packet.Layers()).
			Msg("TCP layer not found")
		return
	}

	tcp, ok := tcpLayer.(*layers.TCP)
	if !ok {
		log.Warn().
			Msg("failed to cast to TCP layer")
		return
	}

	log.Info().
		Str("protocol", ip.Protocol.String()).
		Str("src_ip", ip.SrcIP.String()).
		Str("dst_ip", ip.DstIP.String()).
		Int("src_port", int(tcp.SrcPort)).
		Int("dst_port", int(tcp.DstPort)).
		Str("data", fmt.Sprintf("% x", packet.Data())).
		Msg("reciebed TCP packer")
}

func processUDPLayer(packet gopacket.Packet, ip *layers.IPv4) {
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if udpLayer == nil {
		log.Debug().
			Any("layers", packet.Layers()).
			Msg("UDP layer not found")
		return
	}

	udp, ok := udpLayer.(*layers.UDP)
	if !ok {
		log.Warn().
			Msg("failed to cast to UDP layer")
		return
	}

	log.Info().
		Str("protocol", ip.Protocol.String()).
		Str("src_ip", ip.SrcIP.String()).
		Str("dst_ip", ip.DstIP.String()).
		Int("src_port", int(udp.SrcPort)).
		Int("dst_port", int(udp.DstPort)).
		Str("data", fmt.Sprintf("% x", packet.Data())).
		Msg("reciebed UDP packer")
}
