#!/bin/bash
set -e

# Load shell variables
. ./network/hermes/variables.sh

### Configure the clients and connection
echo "Initiating connection handshake..."
$HERMES_BINARY --config $CONFIG_DIR create connection  --a-chain test-1 --a-client 07-tendermint-0 --b-client 07-tendermint-0

sleep 2
$HERMES_BINARY --config $CONFIG_DIR create channel --a-chain test-1 --a-connection connection-0 --a-port consumer --b-port provider --order ordered --channel-version 1
#transfer channel creation is being initialized automatically by consumer module after "consumer" channel creation, you just need to run a relayer to finish transfer creation