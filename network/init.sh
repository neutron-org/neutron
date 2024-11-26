#!/bin/bash
set -e

BINARY=${BINARY:-neutrond}
BASE_DIR=./data
CHAINID=${CHAINID:-test-1}
STAKEDENOM=${STAKEDENOM:-untrn}
IBCATOMDENOM=${IBCATOMDENOM:-uibcatom}
IBCUSDCDENOM=${IBCUSDCDENOM:-uibcusdc}
CHAIN_DIR="$BASE_DIR/$CHAINID"

P2PPORT=${P2PPORT:-26656}
RPCPORT=${RPCPORT:-26657}
RESTPORT=${RESTPORT:-1317}
ROSETTA=${ROSETTA:-8081}

ORACLE_ADDRESS=${ORACLE_ADDRESS:-localhost:8080}
ORACLE_METRICS_ENABLED=${ORACLE_METRICS_ENABLED:-true}
ORACLE_CLIENT_TIMEOUT=${ORACLE_CLIENT_TIMEOUT:-500ms}

VAL_MNEMONIC_1="clock post desk civil pottery foster expand merit dash seminar song memory figure uniform spice circle try happy obvious trash crime hybrid hood cushion"
VAL_MNEMONIC_2="angry twist harsh drastic left brass behave host shove marriage fall update business leg direct reward object ugly security warm tuna model broccoli choice"
DEMO_MNEMONIC_1="banner spread envelope side kite person disagree path silver will brother under couch edit food venture squirrel civil budget number acquire point work mass"
DEMO_MNEMONIC_2="veteran try aware erosion drink dance decade comic dawn museum release episode original list ability owner size tuition surface ceiling depth seminar capable only"
DEMO_MNEMONIC_3="obscure canal because tomorrow tribe sibling describe satoshi kiwi upgrade bless empty math trend erosion oblige donate label birth chronic hazard ensure wreck shine"
RLY_MNEMONIC_1="alley afraid soup fall idea toss can goose become valve initial strong forward bright dish figure check leopard decide warfare hub unusual join cart"
RLY_MNEMONIC_2="record gift you once hip style during joke field prize dust unique length more pencil transfer quit train device arrive energy sort steak upset"

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

echo "Initializing $CHAINID..."
$BINARY init test --home "$CHAIN_DIR" --chain-id="$CHAINID"

echo "Adding genesis accounts..."
echo "$VAL_MNEMONIC_1" | $BINARY keys add val1 --home "$CHAIN_DIR" --recover --keyring-backend=test
echo "$VAL_MNEMONIC_2" | $BINARY keys add val2 --home "$CHAIN_DIR" --recover --keyring-backend=test
echo "$DEMO_MNEMONIC_1" | $BINARY keys add demowallet1 --home "$CHAIN_DIR" --recover --keyring-backend=test
echo "$DEMO_MNEMONIC_2" | $BINARY keys add demowallet2 --home "$CHAIN_DIR" --recover --keyring-backend=test
echo "$DEMO_MNEMONIC_3" | $BINARY keys add demowallet3 --home "$CHAIN_DIR" --recover --keyring-backend=test
echo "$RLY_MNEMONIC_1" | $BINARY keys add rly1 --home "$CHAIN_DIR" --recover --keyring-backend=test
echo "$RLY_MNEMONIC_2" | $BINARY keys add rly2 --home "$CHAIN_DIR" --recover --keyring-backend=test

# gaia v15+ has genesis prefix for some commands
if [[ "$BINARY" == "gaiad" ]]
then
  GENESIS_PREFIX="genesis"
fi;

$BINARY $GENESIS_PREFIX add-genesis-account "$($BINARY --home "$CHAIN_DIR" keys show val1 --keyring-backend test -a --home "$CHAIN_DIR")" "100000000000000$STAKEDENOM"  --home "$CHAIN_DIR"
$BINARY $GENESIS_PREFIX add-genesis-account "$($BINARY --home "$CHAIN_DIR" keys show val2 --keyring-backend test -a --home "$CHAIN_DIR")" "100000000000000$STAKEDENOM"  --home "$CHAIN_DIR"
$BINARY $GENESIS_PREFIX add-genesis-account "$($BINARY --home "$CHAIN_DIR" keys show demowallet1 --keyring-backend test -a --home "$CHAIN_DIR")" "100000000000000$STAKEDENOM,100000000000000$IBCATOMDENOM,100000000000000$IBCUSDCDENOM"  --home "$CHAIN_DIR"
$BINARY $GENESIS_PREFIX add-genesis-account "$($BINARY --home "$CHAIN_DIR" keys show demowallet2 --keyring-backend test -a --home "$CHAIN_DIR")" "100000000000000$STAKEDENOM,100000000000000$IBCATOMDENOM,100000000000000$IBCUSDCDENOM"  --home "$CHAIN_DIR"
$BINARY $GENESIS_PREFIX add-genesis-account "$($BINARY --home "$CHAIN_DIR" keys show demowallet3 --keyring-backend test -a --home "$CHAIN_DIR")" "100000000000000$STAKEDENOM,100000000000000$IBCATOMDENOM,100000000000000$IBCUSDCDENOM"  --home "$CHAIN_DIR"
$BINARY $GENESIS_PREFIX add-genesis-account "$($BINARY --home "$CHAIN_DIR" keys show rly1 --keyring-backend test -a --home "$CHAIN_DIR")" "100000000000000$STAKEDENOM"  --home "$CHAIN_DIR"
$BINARY $GENESIS_PREFIX add-genesis-account "$($BINARY --home "$CHAIN_DIR" keys show rly2 --keyring-backend test -a --home "$CHAIN_DIR")" "100000000000000$STAKEDENOM"  --home "$CHAIN_DIR"


sed -i -e 's/timeout_commit = "5s"/timeout_commit = "1s"/g' "$CHAIN_DIR/config/config.toml"
sed -i -e 's/timeout_propose = "3s"/timeout_propose = "1s"/g' "$CHAIN_DIR/config/config.toml"
sed -i -e 's/index_all_keys = false/index_all_keys = true/g' "$CHAIN_DIR/config/config.toml"
sed -i -e 's/enable = false/enable = true/g' "$CHAIN_DIR/config/app.toml"
sed -i -e 's/swagger = false/swagger = true/g' "$CHAIN_DIR/config/app.toml"
sed -i -e "s/minimum-gas-prices = \"\"/minimum-gas-prices = \"0.0025$STAKEDENOM,0.0025ibc\/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2\"/g" "$CHAIN_DIR/config/app.toml"
sed -i -e 's/enabled = false/enabled = true/g' "$CHAIN_DIR/config/app.toml"
sed -i -e 's/prometheus-retention-time = 0/prometheus-retention-time = 1000/g' "$CHAIN_DIR/config/app.toml"

sed -i -e 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:'"$P2PPORT"'"#g' "$CHAIN_DIR/config/config.toml"
sed -i -e 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:'"$RPCPORT"'"#g' "$CHAIN_DIR/config/config.toml"
sed -i -e 's#"tcp://localhost:1317"#"tcp://0.0.0.0:'"$RESTPORT"'"#g' "$CHAIN_DIR/config/app.toml"
sed -i -e 's#"tcp://0.0.0.0:1317"#"tcp://0.0.0.0:'"$RESTPORT"'"#g' "$CHAIN_DIR/config/app.toml"
sed -i -e 's#":8080"#":'"$ROSETTA_1"'"#g' "$CHAIN_DIR/config/app.toml"
sed -i -e 's#cors_allowed_origins = \[\]#cors_allowed_origins = ["*"]#g' "$CHAIN_DIR/config/config.toml"

sed -i -e 's/oracle_address = "localhost:8080"/oracle_address = '\""$ORACLE_ADDRESS"\"'/g' "$CHAIN_DIR/config/app.toml"
sed -i -e 's/client_timeout = "2s"/client_timeout = '\""$ORACLE_CLIENT_TIMEOUT"\"'/g' "$CHAIN_DIR/config/app.toml"
sed -i -e 's/metrics_enabled = true/metrics_enabled = '\""$ORACLE_METRICS_ENABLED"\"'/g' "$CHAIN_DIR/config/app.toml"


GENESIS_FILE="$CHAIN_DIR/config/genesis.json"

sed -i -e "s/\"denom\": \"stake\",/\"denom\": \"$STAKEDENOM\",/g" "$GENESIS_FILE"
sed -i -e "s/\"mint_denom\": \"stake\",/\"mint_denom\": \"$STAKEDENOM\",/g" "$GENESIS_FILE"
sed -i -e "s/\"bond_denom\": \"stake\"/\"bond_denom\": \"$STAKEDENOM\"/g" "$GENESIS_FILE"
sed -i -e 's/enabled-unsafe-cors = false/enabled-unsafe-cors = true/g' "$CHAIN_DIR/config/app.toml"
