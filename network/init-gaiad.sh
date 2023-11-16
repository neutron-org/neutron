#!/bin/bash
set -e

BINARY=${BINARY:-gaiad}
BASE_DIR=./data
CHAINID=${CHAINID:-test-1}
CHAIN_DIR="$BASE_DIR/$CHAINID"

STAKEDENOM=${STAKEDENOM:-stake}

echo "Creating and collecting gentx..."
$BINARY gentx val1 "7000000000$STAKEDENOM" --home "$CHAIN_DIR" --chain-id "$CHAINID" --keyring-backend test
$BINARY collect-gentxs --home "$CHAIN_DIR"

sed -i -e 's/\"allow_messages\":.*/\"allow_messages\": [\"\/cosmos.bank.v1beta1.MsgSend\", \"\/cosmos.staking.v1beta1.MsgDelegate\", \"\/cosmos.staking.v1beta1.MsgUndelegate\"]/g' "$CHAIN_DIR/config/genesis.json"
sed -i -e 's/\"max_deposit_period\":.*/\"max_deposit_period\": \"20s\"/g' "$CHAIN_DIR/config/genesis.json"
sed -i -e 's/\"voting_period\":.*/\"voting_period\": \"20s\"/g' "$CHAIN_DIR/config/genesis.json"
