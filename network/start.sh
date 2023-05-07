#!/bin/bash
set -e

BINARY=${BINARY:-neutrond}
BASE_DIR=./data
CHAIN_ID=${CHAIN_ID:-test-1}
GRPCPORT=${GRPCPORT:-9090}
GRPCWEB=${GRPCWEB:-9091}
CHAIN_DIR="$BASE_DIR/$CHAIN_ID"

RUN_BACKGROUND=${RUN_BACKGROUND:-1}

echo "Starting $CHAIN_ID in $CHAIN_DIR..."
echo "Creating log file at $CHAIN_DIR/$CHAIN_ID.log"
if [ "$RUN_BACKGROUND" == 1 ]; then
  $BINARY start                           \
    --log_level debug                     \
    --log_format json                     \
    --home "$CHAIN_DIR"                   \
    --pruning=nothing                     \
    --grpc.address="0.0.0.0:$GRPCPORT"    \
    --grpc-web.address="0.0.0.0:$GRPCWEB" \
    --trace > "$CHAIN_DIR/$CHAIN_ID.log" 2>&1 &
else
  $BINARY start                           \
    --log_level debug                     \
    --log_format json                     \
    --home "$CHAIN_DIR"                   \
    --pruning=nothing                     \
    --grpc.address="0.0.0.0:$GRPCPORT"    \
    --grpc-web.address="0.0.0.0:$GRPCWEB" \
    --trace 2>&1 | tee "$CHAIN_DIR/$CHAIN_ID.log"
fi

