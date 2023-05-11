#!/bin/bash
# microtick and bitcanna contributed significantly here.
# Pebbledb state sync script.
# invoke like: bash contrib/statesync.bash

## USAGE RUNDOWN
# Not for use on live nodes
# For use when testing.

set -uxe

# Set Golang environment variables.
# ! Adapt as required, depending on your system configuration
#export GOPATH=~/go
#export PATH=$PATH:~/go/bin

# Install with pebbledb (uncomment for incredible performance)
go mod edit -replace github.com/tendermint/tm-db=github.com/baabeetaa/tm-db@pebble
go mod tidy

# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=pebbledb -X github.com/tendermint/tm-db.ForceSync=1' -tags pebbledb ./...

# install (comment if using pebble for incredible performance)
# go install ./...


go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=pebbledb -X github.com/tendermint/tm-db.ForceSync=1' -tags pebbledb ./...

# NOTE: ABOVE YOU CAN USE ALTERNATIVE DATABASES, HERE ARE THE EXACT COMMANDS
# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=rocksdb' -tags rocksdb ./...
# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=badgerdb' -tags badgerdb ./...
# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=boltdb' -tags boltdb ./...
# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=pebbledb -X github.com/tendermint/tm-db.ForceSync=1' -tags pebbledb ./...


# Initialize chain.
neutrond init test

# Get Genesis
wget -O ~/.neutrond/config/genesis.json https://raw.githubusercontent.com/neutron-org/mainnet-assets/main/neutron-1-genesis.json


# Get "trust_hash" and "trust_height".
INTERVAL=100
LATEST_HEIGHT=$(curl -s https://rpc-kralum.neutron-1.neutron.org/block | jq -r .result.block.header.height)
BLOCK_HEIGHT=$((LATEST_HEIGHT - INTERVAL))
TRUST_HASH=$(curl -s "https://rpc-kralum.neutron-1.neutron.org/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)

# Print out block and transaction hash from which to sync state.
echo "trust_height: $BLOCK_HEIGHT"
echo "trust_hash: $TRUST_HASH"

# Export state sync variables.
export NEUTROND_STATESYNC_ENABLE=true
export NEUTROND_P2P_MAX_NUM_OUTBOUND_PEERS=500
export NEUTROND_STATESYNC_RPC_SERVERS="https://rpc-kralum.neutron-1.neutron.org:443,https://rpc-kralum.neutron-1.neutron.org:443"
export NEUTROND_STATESYNC_TRUST_HEIGHT=$BLOCK_HEIGHT
export NEUTROND_STATESYNC_TRUST_HASH=$TRUST_HASH
export NEUTROND_P2P_LADDR=tcp://0.0.0.0:8010
export NEUTROND_RPC_LADDR=tcp://127.0.0.1:8011
export NEUTROND_GRPC_ADDRESS=127.0.0.1:8012
export NEUTROND_GRPC_WEB_ADDRESS=127.0.0.1:8014
export NEUTROND_API_ADDRESS=tcp://127.0.0.1:8013
export NEUTROND_NODE=tcp://127.0.0.1:8011
export NEUTROND_P2P_MAX_NUM_INBOUND_PEERS=500
export NEUTROND_RPC_PPROF_LADDR=127.0.0.1:6969

# Fetch and set list of seeds from chain registry.
NEUTROND_P2P_SEEDS=$(curl -s https://raw.githubusercontent.com/cosmos/chain-registry/master/neutron/chain.json | jq -r '[foreach .peers.seeds[] as $item (""; "\($item.id)@\($item.address)")] | join(",")')
export NEUTROND_P2P_SEEDS

# Start chain.
neutrond start --x-crisis-skip-assert-invariants --iavl-disable-fastnode false 
