#!/bin/bash
set -e

BINARY=${BINARY:-neutrond}
BASE_DIR=./data
CHAINID=${CHAINID:-test-1}
GRPCPORT=${GRPCPORT:-9090}
GRPCWEB=${GRPCWEB:-9091}
CHAIN_DIR="$BASE_DIR/$CHAINID"
NODES=${NODES:-10}


RUN_BACKGROUND=${RUN_BACKGROUND:-1}

echo "Starting $CHAINID in $CHAIN_DIR..."
echo "Creating log file at $CHAIN_DIR/$CHAINID.log"
if [ "$RUN_BACKGROUND" == 1 ]; then
  if [[ "$BINARY" == "gaiad" ]]
  then
    $BINARY start                           \
          --log_level debug                     \
          --log_format json                     \
          --home "$CHAIN_DIR"                   \
          --pruning=nothing                     \
          --grpc.address="0.0.0.0:$GRPCPORT"    \
          --trace > "$CHAIN_DIR/$CHAINID.log" 2>&1 &
  else
    for i in `seq 1 ${NODES}`; do
      $BINARY start                           \
        --log_level debug                     \
        --log_format json                     \
        --home "$CHAIN_DIR/node-${i}"                   \
        --pruning=nothing                     \
        --grpc.address="0.0.0.0:$GRPCPORT"    \
        --trace > "$CHAIN_DIR/node-${i}/$CHAINID.log" 2>&1 &
    done;
  fi;
else
  $BINARY start                           \
    --log_level debug                     \
    --log_format json                     \
    --home "$CHAIN_DIR"                   \
    --pruning=nothing                     \
    --grpc.address="0.0.0.0:$GRPCPORT"    \
    --trace 2>&1 | tee "$CHAIN_DIR/$CHAINID.log"
fi

