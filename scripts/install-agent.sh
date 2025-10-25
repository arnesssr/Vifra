#!/bin/bash

# Installation script for VPS Monitor Agent

set -e

# Default values
INSTALL_DIR="/opt/vps-monitor-agent"
CONFIG_DIR="/etc/vps-monitor-agent"
SERVICE_NAME="vps-monitor-agent"

echo "Installing VPS Monitor Agent..."

# Create directories
echo "Creating directories..."
sudo mkdir -p $INSTALL_DIR
sudo mkdir -p $CONFIG_DIR

# Copy binary
echo "Copying agent binary..."
sudo cp bin/agent-linux-amd64 $INSTALL_DIR/agent
sudo chmod +x $INSTALL_DIR/agent

# Create configuration file
echo "Creating configuration file..."
sudo tee $CONFIG_DIR/config.env > /dev/null <<EOF
# VPS Monitor Agent Configuration
SERVER_URL=https://your-monitoring-server.com
AGENT_KEY=your-agent-key-here
SERVER_ID=your-server-id-here
COLLECTION_INTERVAL=30s
EOF

# Create systemd service file
echo "Creating systemd service..."
sudo tee /etc/systemd/system/$SERVICE_NAME.service > /dev/null <<EOF
[Unit]
Description=VPS Monitor Agent
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
EnvironmentFile=$CONFIG_DIR/config.env
ExecStart=$INSTALL_DIR/agent
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Set permissions
sudo chown -R root:root $INSTALL_DIR
sudo chown -R root:root $CONFIG_DIR
sudo chmod 644 /etc/systemd/system/$SERVICE_NAME.service

# Reload systemd
echo "Reloading systemd..."
sudo systemctl daemon-reload

echo "Installation completed!"
echo ""
echo "Next steps:"
echo "1. Edit the configuration file: sudo nano $CONFIG_DIR/config.env"
echo "2. Start the service: sudo systemctl start $SERVICE_NAME"
echo "3. Enable auto-start: sudo systemctl enable $SERVICE_NAME"
echo "4. Check status: sudo systemctl status $SERVICE_NAME"