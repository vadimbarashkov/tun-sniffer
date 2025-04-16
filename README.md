# tun-sniffer

**tun-sniffer** is a simple CLI that creates a TUN interface, captures IP packets (TCP/UDP) and log information about them (protocol, source/destination IP addresses and ports, raw packets bytes).

## Requirements

- Go 1.20+
- Linux
- Docker

## Display available commands

```bash
make help
```

## Configuration

```bash
-env string
      Environment (dev, prod) (default "dev")
-logHandler string
      Log handler format (text, json) (default "text")
-logLevel string
      Log level (debug, info, warn, error) (default "debug")
-tunIP string
      TUN interface IP (default "10.0.0.1/24")
-tunRoute string
      TUN interface route (default "10.0.0.0/24")
```

## Build and Run

```bash
# Default run. Root access is required.
sudo make run

# You can override default arguments.
sudo make run TUN_IP=10.0.1.1/24 TUN_ROUTE=10.0.1.0/24
```

## Build and Run with Docker

```bash
# Default build and run.
make docker-run

# You can override default arguments.
make docker-run ENV=prod LOG_HANDLER=json LOG_LEVEL=info
```

## Project Structure

```bash
cmd/             # Main application entry point
internal/config/ # Configuration loading
```
