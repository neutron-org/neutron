#!/bin/bash
set -e

BINARY=${BINARY:-gaiad}
BASE_DIR=./data
CHAIN_ID=${CHAIN_ID:-test-1}
CHAIN_DIR="$BASE_DIR/$CHAIN_ID"

STAKEDENOM=${STAKEDENOM:-stake}

echo "Creating and collecting gentx..."
$BINARY gentx val1 "7000000000$STAKEDENOM" --home "$CHAIN_DIR" --chain-id "$CHAIN_ID" --keyring-backend test
$BINARY collect-gentxs --home "$CHAIN_DIR"

sed -i -e 's/\"allow_messages\":.*/\"allow_messages\": [\"\/cosmos.bank.v1beta1.MsgSend\", \"\/cosmos.staking.v1beta1.MsgDelegate\", \"\/cosmos.staking.v1beta1.MsgUndelegate\"]/g' "$CHAIN_DIR/config/genesis.json"
