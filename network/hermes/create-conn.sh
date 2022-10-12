#!/bin/bash
set -e

# Load shell variables
. ./network/hermes/variables.sh

### Configure the clients and connection
echo "Initiating connection handshake..."
$HERMES_BINARY -c $CONFIG_DIR create connection test-1 test-2

sleep 2
hermes -c ./network/hermes/config.toml create channel --port-a transfer --port-b transfer test-1 connection-0
