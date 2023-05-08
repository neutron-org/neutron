#!/bin/bash
set -e

BINARY=${BINARY:-neutrond}
BASE_DIR=./data
CHAIN_ID=${CHAIN_ID:-test-1}
STAKEDENOM=${STAKEDENOM:-untrn}
IBCATOMDENOM=${IBCATOMDENOM:-uibcatom}
IBCUSDCDENOM=${IBCUSDCDENOM:-uibcusdc}
CHAIN_DIR="$BASE_DIR/$CHAIN_ID"

P2PPORT=${P2PPORT:-26656}
RPCPORT=${RPCPORT:-26657}
RESTPORT=${RESTPORT:-1317}
ROSETTA=${ROSETTA:-8081}

# Stop if it is already running
if pgrep -x "$BINARY" >/dev/null; then
    echo "Terminating $BINARY..."
    killall "$BINARY"
fi

echo "Removing previous data..."
rm -rf "$CHAIN_DIR" &> /dev/null

# Add directories for both chains, exit if an error occurs
if ! mkdir -p "$CHAIN_DIR" 2>/dev/null; then
    echo "Failed to create chain folder. Aborting..."
    exit 1
fi

echo "Initializing $CHAIN_ID..."
$BINARY init test --home "$CHAIN_DIR" --chain-id="$CHAIN_ID"

echo "Adding genesis accounts..."
$BINARY add-genesis-account "neutron1hfdndqcqfyrql53lzmmtunezufmz3h4335znl4" "0$STAKEDENOM"  --home "$CHAIN_DIR"

sed -i -e 's/timeout_commit = "5s"/timeout_commit = "1s"/g' "$CHAIN_DIR/config/config.toml"
sed -i -e 's/timeout_propose = "3s"/timeout_propose = "1s"/g' "$CHAIN_DIR/config/config.toml"
sed -i -e 's/index_all_keys = false/index_all_keys = true/g' "$CHAIN_DIR/config/config.toml"
sed -i -e 's/enable = false/enable = true/g' "$CHAIN_DIR/config/app.toml"
sed -i -e 's/swagger = false/swagger = true/g' "$CHAIN_DIR/config/app.toml"
#sed -i -e "s/minimum-gas-prices = \"\"/minimum-gas-prices = \"0.0025$STAKEDENOM,0.0025ibc\/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2\"/g" "$CHAIN_DIR/config/app.toml"
sed -i -e 's/enabled = false/enabled = true/g' "$CHAIN_DIR/config/app.toml"
sed -i -e 's/prometheus-retention-time = 0/prometheus-retention-time = 1000/g' "$CHAIN_DIR/config/app.toml"

sed -i -e 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:'"$P2PPORT"'"#g' "$CHAIN_DIR/config/config.toml"
sed -i -e 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:'"$RPCPORT"'"#g' "$CHAIN_DIR/config/config.toml"
sed -i -e 's#"tcp://0.0.0.0:1317"#"tcp://0.0.0.0:'"$RESTPORT"'"#g' "$CHAIN_DIR/config/app.toml"
sed -i -e 's#":8080"#":'"$ROSETTA_1"'"#g' "$CHAIN_DIR/config/app.toml"

GENESIS_FILE="$CHAIN_DIR/config/genesis.json"

sed -i -e "s/\"denom\": \"stake\",/\"denom\": \"$STAKEDENOM\",/g" "$GENESIS_FILE"
sed -i -e "s/\"mint_denom\": \"stake\",/\"mint_denom\": \"$STAKEDENOM\",/g" "$GENESIS_FILE"
sed -i -e "s/\"bond_denom\": \"stake\"/\"bond_denom\": \"$STAKEDENOM\"/g" "$GENESIS_FILE"
