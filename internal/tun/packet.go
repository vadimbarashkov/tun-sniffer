package tun

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/rs/zerolog/log"
	"github.com/songgao/water"
)

const bufferSize = 2000

func ReadPackets(ctx context.Context, ifce *water.Interface, maxGoroutines int) {
	buff := make([]byte, bufferSize)
	var wg sync.WaitGroup
	sema := make(chan struct{}, maxGoroutines)

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

			sema <- struct{}{}
			wg.Add(1)
			go func(data []byte) {
				defer wg.Done()
				defer func() { <-sema }()
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
		Msg("recieved TCP packet")
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
		Msg("recieved UDP packet")
}
