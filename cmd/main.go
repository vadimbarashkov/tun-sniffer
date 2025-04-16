//go:build linux

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"slices"
	"strings"
	"sync"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/songgao/water"

	"github.com/vadimbarashkov/tun-sniffer/internal/config"
)

const (
	bufferSize = 2000
)

func main() {
	flag.Usage = config.Usage

	cfg, err := config.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

	logger := config.SetupLogger(os.Stdout, cfg.LogLevel, cfg.Env, cfg.LogHandler)

	ifce, err := setupTun()
	if err != nil {
		logger.Error("Failed to set up TUN interface", slog.Any("err", err))
		os.Exit(1)
	}
	defer func() {
		if err := ifce.Close(); err != nil {
			logger.Error("Failed to close TUN interface", slog.Any("err", err))
		}
	}()

	if err := configureTunInterface(ifce.Name(), cfg.TunIP, cfg.TunRoute); err != nil {
		logger.Error("Failed to configure TUN interface", slog.Any("err", err))
		os.Exit(1)
	}

	logger.Info("TUN interface configured", slog.String("name", ifce.Name()))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	readPackets(ctx, logger, ifce)
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
		return fmt.Errorf("failed to bring up: %w", err)
	}

	out, err := exec.Command("ip", "route", "add", tunRoute, "via", strings.Split(tunIP, "/")[0], "dev", name).CombinedOutput()
	if err != nil {
		if !bytes.Contains(out, []byte("File exists")) {
			return fmt.Errorf("failed to add route: %w", err)
		}
	}

	return nil
}

func readPackets(ctx context.Context, logger *slog.Logger, ifce *water.Interface) {
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
				logger.Error("Failed to read packet", slog.Any("err", err))
				continue
			}

			wg.Add(1)
			go func(data []byte) {
				defer wg.Done()
				packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)
				processPacket(logger, packet)
			}(slices.Clone(buff[:n]))
		}
	}
}

func processPacket(logger *slog.Logger, packet gopacket.Packet) {
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		logger.Debug("Non-IP packet recieved", slog.Any("layers", packet.Layers()))
		return
	}

	ip, ok := ipLayer.(*layers.IPv4)
	if !ok {
		logger.Warn("Failed to cast to IPv4 layer")
		return
	}

	switch ip.Protocol {
	case layers.IPProtocolTCP:
		processTCPLayer(logger, packet, ip)
	case layers.IPProtocolUDP:
		processUDPLayer(logger, packet, ip)
	default:
		logger.Debug("Non-TCP/UDP packet recieved", slog.String("protocol", ip.Protocol.String()))
	}
}

func processTCPLayer(logger *slog.Logger, packet gopacket.Packet, ip *layers.IPv4) {
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		logger.Debug("TCP layer not found", slog.Any("layers", packet.Layers()))
		return
	}

	tcp, ok := tcpLayer.(*layers.TCP)
	if !ok {
		logger.Warn("Failed to cast to TCP layer")
		return
	}

	logger.Info("Recieved TCP packet",
		slog.String("protocol", ip.Protocol.String()),
		slog.String("src_ip", ip.SrcIP.String()),
		slog.String("dst_ip", ip.DstIP.String()),
		slog.Int("src_port", int(tcp.SrcPort)),
		slog.Int("dst_port", int(tcp.DstPort)),
		slog.String("data", fmt.Sprintf("% x", packet.Data())),
	)
}

func processUDPLayer(logger *slog.Logger, packet gopacket.Packet, ip *layers.IPv4) {
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if udpLayer == nil {
		logger.Debug("UDP layer not found", slog.Any("layers", packet.Layers()))
		return
	}

	udp, ok := udpLayer.(*layers.UDP)
	if !ok {
		logger.Warn("Failed to cast to UDP layer")
		return
	}

	logger.Info("Recieved UDP packet",
		slog.String("protocol", ip.Protocol.String()),
		slog.String("src_ip", ip.SrcIP.String()),
		slog.String("dst_ip", ip.DstIP.String()),
		slog.Int("src_port", int(udp.SrcPort)),
		slog.Int("dst_port", int(udp.DstPort)),
		slog.String("data", fmt.Sprintf("% x", packet.Data())),
	)
}
