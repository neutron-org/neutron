#!/bin/bash

set -x

BINARY=${BINARY:-neutrond}
CHAIN_DIR=./data
CHAINID=${CHAINID:-test-1}
GRPCPORT=${GRPCPORT:-9090}
GRPCWEB=${GRPCWEB:-9091}

RUN_BACKGROUND=${RUN_BACKGROUND:-1}

echo "Starting $CHAINID in $CHAIN_DIR..."
echo "Creating log file at $CHAIN_DIR/$CHAINID.log"
if [ "$RUN_BACKGROUND" == 1 ]; then
  $BINARY start                           \
    --log_level debug                     \
    --log_format json                     \
    --home $CHAIN_DIR/$CHAINID            \
    --pruning=nothing                     \
    --grpc.address="0.0.0.0:$GRPCPORT"    \
    --grpc-web.address="0.0.0.0:$GRPCWEB" \
    --trace > $CHAIN_DIR/$CHAINID.log 2>&1 &
else
  $BINARY start                           \
    --log_level debug                     \
    --log_format json                     \
    --home $CHAIN_DIR/$CHAINID            \
    --pruning=nothing                     \
    --grpc.address="0.0.0.0:$GRPCPORT"    \
    --grpc-web.address="0.0.0.0:$GRPCWEB" \
    --trace 2>&1 | tee $CHAIN_DIR/$CHAINID.log
fi

