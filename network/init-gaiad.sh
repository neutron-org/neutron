#!/bin/bash

BINARY=${BINARY:-gaiad}
CHAIN_DIR=./data
CHAINID=${CHAINID:-test-1}

STAKEDENOM=${STAKEDENOM:-stake}

echo "Creating and collecting gentx..."
$BINARY gentx val1 7000000000${STAKEDENOM} --home $CHAIN_DIR/$CHAINID --chain-id $CHAINID --keyring-backend test
$BINARY collect-gentxs --home $CHAIN_DIR/$CHAINID

sed -i -e 's/\"allow_messages\":.*/\"allow_messages\": [\"\/cosmos.bank.v1beta1.MsgSend\", \"\/cosmos.staking.v1beta1.MsgDelegate\", \"\/cosmos.staking.v1beta1.MsgUndelegate\"]/g' $CHAIN_DIR/$CHAINID/config/genesis.json
