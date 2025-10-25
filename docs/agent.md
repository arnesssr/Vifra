# VPS Monitor Agent

## Overview

The VPS Monitor Agent is a lightweight daemon that runs on each monitored VPS server to collect system metrics and submit them to the Dashboard Server.

## Features

- **Lightweight**: Minimal resource usage (< 1% CPU, < 50MB RAM)
- **Secure**: Uses AgentKey authentication for secure communication
- **Reliable**: Automatic retry mechanisms and error handling
- **Cross-platform**: Supports Linux, Windows, and macOS
- **Configurable**: Adjustable collection intervals and settings

## System Requirements

- **Operating Systems**: Linux (Ubuntu, CentOS, Debian), Windows, macOS
- **Architecture**: x86_64, ARM64
- **Resources**: < 1% CPU, < 50MB RAM

## Installation

### Building from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/username/vps-monitor.git
   cd vps-monitor
   ```

2. Build the agent:
   ```bash
   ./scripts/build-agent.sh
   ```

3. The binaries will be created in the `bin/` directory.

### Installation on Linux

1. Run the installation script:
   ```bash
   sudo ./scripts/install-agent.sh
   ```

2. Configure the agent:
   ```bash
   sudo nano /etc/vps-monitor-agent/config.env
   ```

3. Start the service:
   ```bash
   sudo systemctl start vps-monitor-agent
   ```

4. Enable auto-start:
   ```bash
   sudo systemctl enable vps-monitor-agent
   ```

## Configuration

The agent is configured using environment variables. Create a configuration file at `/etc/vps-monitor-agent/config.env`:

```env
# Dashboard Server URL
SERVER_URL=https://your-monitoring-server.com

# Agent authentication key (provided during server registration)
AGENT_KEY=your-agent-key-here

# Server ID (provided during server registration)
SERVER_ID=123

# Metrics collection interval (e.g., 30s, 1m, 5m)
COLLECTION_INTERVAL=30s
```

## Metrics Collected

The agent collects the following system metrics:

- **CPU Usage**: Overall CPU utilization percentage
- **Memory Usage**: Used and total memory in bytes
- **Disk Usage**: Used and total disk space in bytes
- **Network I/O**: Incoming and outgoing network traffic (planned)
- **Load Average**: System load average (1-minute)

## Security

- **Authentication**: Uses AgentKey for secure authentication with the Dashboard Server
- **Encryption**: All communication is encrypted using HTTPS/TLS
- **Isolation**: Each agent has unique credentials
- **Minimal Permissions**: Runs with minimal required privileges

## Troubleshooting

### Check Agent Status

```bash
sudo systemctl status vps-monitor-agent
```

### View Logs

```bash
sudo journalctl -u vps-monitor-agent -f
```

### Common Issues

1. **Connection Refused**: Verify the SERVER_URL is correct and the Dashboard Server is running
2. **Authentication Failed**: Check that AGENT_KEY and SERVER_ID are correct
3. **Permission Denied**: Ensure the agent has necessary permissions to read system metrics

## Development

### Building

To build the agent for different platforms:

```bash
./scripts/build-agent.sh
```

### Running Locally

```bash
SERVER_URL=http://localhost:8080 \
AGENT_KEY=your-agent-key \
SERVER_ID=1 \
COLLECTION_INTERVAL=10s \
go run cmd/agent/main.go
```

## Architecture

The agent follows a simple collection and submission pattern:

1. **Initialization**: Load configuration from environment variables
2. **Collection Loop**: 
   - Collect system metrics at regular intervals
   - Format metrics according to API specification
   - Submit metrics to Dashboard Server via HTTPS
3. **Error Handling**: Log errors and continue operation
4. **Graceful Shutdown**: Handle interrupt signals properly

## API Integration

The agent communicates with the Dashboard Server using the following endpoint:

```
POST /api/v1/metrics
```

Request body:
```json
{
  "server_id": 123,
  "cpu_usage": 25.5,
  "memory_used": 1073741824,
  "memory_total": 2147483648,
  "disk_used": 5368709120,
  "disk_total": 10737418240,
  "network_in": 102400,
  "network_out": 51200,
  "load_avg": 0.75,
  "timestamp": "2023-01-01T12:00:00Z"
}
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request