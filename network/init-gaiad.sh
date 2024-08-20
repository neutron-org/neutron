#!/bin/bash
set -e

BINARY=${BINARY:-gaiad}
BASE_DIR=./data
CHAINID=${CHAINID:-test-1}
CHAIN_DIR="$BASE_DIR/$CHAINID"

STAKEDENOM=${STAKEDENOM:-stake}

function set_genesis_param_jq() {
  param_path=$1
  param_value=$2
  jq "${param_path} = ${param_value}" > tmp_genesis_file.json < "$CHAIN_DIR/config/genesis.json" && mv tmp_genesis_file.json "$CHAIN_DIR/config/genesis.json"
}

echo "Creating and collecting gentx..."
$BINARY genesis gentx val1 "7000000000$STAKEDENOM" --home "$CHAIN_DIR" --chain-id "$CHAINID" --keyring-backend test
$BINARY genesis collect-gentxs --home "$CHAIN_DIR"

sed -i -e 's/\*/\/cosmos.bank.v1beta1.MsgSend\", \"\/cosmos.staking.v1beta1.MsgDelegate\", \"\/cosmos.staking.v1beta1.MsgUndelegate/g' "$CHAIN_DIR/config/genesis.json"

set_genesis_param_jq ".app_state.feemarket.params.enabled" "false"                      # feemarket
set_genesis_param_jq ".app_state.feemarket.params.fee_denom"       "\"uatom\""          # feemarket
set_genesis_param_jq ".app_state.feemarket.state.base_gas_price" "\"0.0025\""           # feemarket
set_genesis_param_jq ".app_state.feemarket.params.min_base_gas_price"    "\"0.0025\""   # feemarket

set_genesis_param_jq ".app_state.ibc.client_genesis.params.allowed_clients" "[\"*\"]"         # ibc