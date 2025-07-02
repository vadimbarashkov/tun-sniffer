package tun

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/songgao/water"
)

func Setup() (*water.Interface, error) {
	ifce, err := water.New(water.Config{DeviceType: water.TUN})
	if err != nil {
		return nil, fmt.Errorf("failed to create TUN interface: %w", err)
	}
	return ifce, nil
}

func Configure(name, tunIP, tunRoute string) error {
	if err := exec.Command("ip", "addr", "add", tunIP, "dev", name).Run(); err != nil {
		return fmt.Errorf("failed to add IP address: %w", err)
	}
	if err := exec.Command("ip", "link", "set", name, "up").Run(); err != nil {
		return fmt.Errorf("failed to bring interface up: %w", err)
	}

	out, err := exec.Command("ip", "route", "add", tunRoute, "via", strings.Split(tunIP, "/")[0], "dev", name).CombinedOutput()
	if err != nil && !bytes.Contains(out, []byte("File exists")) {
		if !bytes.Contains(out, []byte("File exists")) {
			return fmt.Errorf("failed to add route: %w", err)
		}
	}

	return nil
}
