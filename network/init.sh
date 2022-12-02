#!/bin/bash

ADMIN_ADDRESS=neutron14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s5c2epq
BINARY=neutrond
CHAIN_DIR=./data
CHAINID_1=test-1
CHAINID_2=test-2
DAO_CONTRACT=./contracts/cwd_core.wasm
PROPOSAL_CONTRACT=./contracts/cwd_proposal_single.wasm
VOTING_REGISTRY_CONTRACT=./contracts/neutron_voting_registry.wasm
VAULT_CONTRACT=./contracts/neutron_vault.wasm
VAL_MNEMONIC_1="clock post desk civil pottery foster expand merit dash seminar song memory figure uniform spice circle try happy obvious trash crime hybrid hood cushion"
VAL_MNEMONIC_2="angry twist harsh drastic left brass behave host shove marriage fall update business leg direct reward object ugly security warm tuna model broccoli choice"
DEMO_MNEMONIC_1="banner spread envelope side kite person disagree path silver will brother under couch edit food venture squirrel civil budget number acquire point work mass"
DEMO_MNEMONIC_2="veteran try aware erosion drink dance decade comic dawn museum release episode original list ability owner size tuition surface ceiling depth seminar capable only"
DEMO_MNEMONIC_3="obscure canal because tomorrow tribe sibling describe satoshi kiwi upgrade bless empty math trend erosion oblige donate label birth chronic hazard ensure wreck shine"
RLY_MNEMONIC_1="alley afraid soup fall idea toss can goose become valve initial strong forward bright dish figure check leopard decide warfare hub unusual join cart"
RLY_MNEMONIC_2="record gift you once hip style during joke field prize dust unique length more pencil transfer quit train device arrive energy sort steak upset"
P2PPORT_1=16656
P2PPORT_2=26656
RPCPORT_1=16657
RPCPORT_2=26657
RESTPORT_1=1316
RESTPORT_2=1317
ROSETTA_1=8080
ROSETTA_2=8081

# Stop if it is already running
if pgrep -x "$BINARY" >/dev/null; then
    echo "Terminating $BINARY..."
    killall $BINARY
fi

echo "Removing previous data..."
rm -rf $CHAIN_DIR/$CHAINID_1 &> /dev/null
rm -rf $CHAIN_DIR/$CHAINID_2 &> /dev/null

# Add directories for both chains, exit if an error occurs
if ! mkdir -p $CHAIN_DIR/$CHAINID_1 2>/dev/null; then
    echo "Failed to create chain folder. Aborting..."
    exit 1
fi

if ! mkdir -p $CHAIN_DIR/$CHAINID_2 2>/dev/null; then
    echo "Failed to create chain folder. Aborting..."
    exit 1
fi

echo "Initializing $CHAINID_1..."
echo "Initializing $CHAINID_2..."
$BINARY init test --home $CHAIN_DIR/$CHAINID_1 --chain-id=$CHAINID_1
$BINARY init test --home $CHAIN_DIR/$CHAINID_2 --chain-id=$CHAINID_2

echo "Adding genesis accounts..."
echo $VAL_MNEMONIC_1 | $BINARY keys add val1 --home $CHAIN_DIR/$CHAINID_1 --recover --keyring-backend=test
echo $VAL_MNEMONIC_2 | $BINARY keys add val2 --home $CHAIN_DIR/$CHAINID_2 --recover --keyring-backend=test
echo $DEMO_MNEMONIC_1 | $BINARY keys add demowallet1 --home $CHAIN_DIR/$CHAINID_1 --recover --keyring-backend=test
echo $DEMO_MNEMONIC_2 | $BINARY keys add demowallet2 --home $CHAIN_DIR/$CHAINID_2 --recover --keyring-backend=test
echo $DEMO_MNEMONIC_3 | $BINARY keys add demowallet3 --home $CHAIN_DIR/$CHAINID_1 --recover --keyring-backend=test
echo $RLY_MNEMONIC_1 | $BINARY keys add rly1 --home $CHAIN_DIR/$CHAINID_1 --recover --keyring-backend=test
echo $RLY_MNEMONIC_2 | $BINARY keys add rly2 --home $CHAIN_DIR/$CHAINID_2 --recover --keyring-backend=test

$BINARY add-genesis-account $($BINARY --home $CHAIN_DIR/$CHAINID_1 keys show val1 --keyring-backend test -a) 100000000000stake  --home $CHAIN_DIR/$CHAINID_1
$BINARY add-genesis-account $($BINARY --home $CHAIN_DIR/$CHAINID_2 keys show val2 --keyring-backend test -a) 100000000000stake  --home $CHAIN_DIR/$CHAINID_2
$BINARY add-genesis-account $($BINARY --home $CHAIN_DIR/$CHAINID_1 keys show demowallet1 --keyring-backend test -a) 100000000000stake  --home $CHAIN_DIR/$CHAINID_1
$BINARY add-genesis-account $($BINARY --home $CHAIN_DIR/$CHAINID_2 keys show demowallet2 --keyring-backend test -a) 100000000000stake  --home $CHAIN_DIR/$CHAINID_2
$BINARY add-genesis-account $($BINARY --home $CHAIN_DIR/$CHAINID_1 keys show demowallet3 --keyring-backend test -a) 100000000000stake  --home $CHAIN_DIR/$CHAINID_1
$BINARY add-genesis-account $($BINARY --home $CHAIN_DIR/$CHAINID_1 keys show rly1 --keyring-backend test -a) 100000000000stake  --home $CHAIN_DIR/$CHAINID_1
$BINARY add-genesis-account $($BINARY --home $CHAIN_DIR/$CHAINID_2 keys show rly2 --keyring-backend test -a) 100000000000stake  --home $CHAIN_DIR/$CHAINID_2

echo "Initializing dao contract in genesis..."
# Upload the dao contract
$BINARY add-wasm-message store ${VAULT_CONTRACT} --output json --run-as neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2 --keyring-backend=test --home $CHAIN_DIR/$CHAINID_1
$BINARY add-wasm-message store ${DAO_CONTRACT} --output json  --run-as neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2 --home $CHAIN_DIR/$CHAINID_1
$BINARY add-wasm-message store ${PROPOSAL_CONTRACT} --output json  --run-as neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2 --home $CHAIN_DIR/$CHAINID_1
$BINARY add-wasm-message store ${VOTING_REGISTRY_CONTRACT} --output json  --run-as neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2 --home $CHAIN_DIR/$CHAINID_1
# Instantiate the contract
#INIT_CONTRACT="$(printf '{"owner":"%s"}' "${ADMIN_ADDRESS}")"
INIT="$(printf '{"denom":"stake"}')"
#TODO: fix quotes issue
DAO_INIT="'""'"
#echo "Instantiate"
$BINARY add-wasm-message  instantiate-contract 1 ${INIT} --run-as neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2 --admin ${ADMIN_ADDRESS}  --label "DAO_Neutron_voting_vault"  --home $CHAIN_DIR/$CHAINID_1
$BINARY add-wasm-message  instantiate-contract 2 '{
                                                               "admin": null,
                                                               "automatically_add_cw20s": false,
                                                               "automatically_add_cw721s": false,
                                                               "description": "basic neutron dao",
                                                               "image_url": null,
                                                               "name": "Neutron",
                                                               "initial_items": null,
                                                               "proposal_modules_instantiate_info": [
                                                                 {
                                                                   "admin": null,
                                                                   "code_id": 3,
                                                                   "label": "DAO_Neutron_cw-proposal-single",
                                                                   "msg": "ewogICAgICAgICJhbGxvd19yZXZvdGluZyI6IGZhbHNlLAogICAgICAgICJwcmVfcHJvcG9zZV9pbmZvIjogewogICAgICAgICAgIkFueW9uZU1heVByb3Bvc2UiOiB7fQogICAgICAgIH0sCiAgICAgICAgImRlcG9zaXRfaW5mbyI6IG51bGwsCiAgICAgICAgImNsb3NlX3Byb3Bvc2FsX29uX2V4ZWN1dGlvbl9mYWlsdXJlIjogZmFsc2UsCiAgICAgICAgIm1heF92b3RpbmdfcGVyaW9kIjogewogICAgICAgICAgInRpbWUiOiA2MDQ4MDAKICAgICAgICB9LAogICAgICAgICJvbmx5X21lbWJlcnNfZXhlY3V0ZSI6IGZhbHNlLAogICAgICAgICJ0aHJlc2hvbGQiOiB7CiAgICAgICAgICAidGhyZXNob2xkX3F1b3J1bSI6IHsKICAgICAgICAgICAgInF1b3J1bSI6IHsKICAgICAgICAgICAgICAicGVyY2VudCI6ICIwLjIwIgogICAgICAgICAgICB9LAogICAgICAgICAgICAidGhyZXNob2xkIjogewogICAgICAgICAgICAgICJtYWpvcml0eSI6IHt9CiAgICAgICAgICAgIH0KICAgICAgICAgIH0KICAgICAgICB9CiAgICAgIH0="
                                                                 }
                                                               ],
                                                               "voting_module_instantiate_info": {
                                                                 "admin": null,
                                                                 "code_id": 4,
                                                                 "label": "DAO_Neutron_voting_registry",
                                                                 "msg": "ewogICAgICAibWFuYWdlciI6IG51bGwsCiAgICAgICJvd25lciI6IG51bGwsCiAgICAgICJzdGFraW5nIjogIm5ldXRyb24xNGhqMnRhdnE4ZnBlc2R3eHhjdTQ0cnR5M2hoOTB2aHVqcnZjbXN0bDR6cjN0eG1mdnc5czVjMmVwcSIKICAgIH0="
                                                               }
                                                       }' --run-as neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2 --admin ${ADMIN_ADDRESS}  --label "DAO"  --home $CHAIN_DIR/$CHAINID_1


echo "Add consumer section..."
$BINARY add-consumer-section --home $CHAIN_DIR/$CHAINID_1
$BINARY add-consumer-section --home $CHAIN_DIR/$CHAINID_2

echo "Changing defaults and ports in app.toml and config.toml files..."
sed -i -e 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:'"$P2PPORT_1"'"#g' $CHAIN_DIR/$CHAINID_1/config/config.toml
sed -i -e 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:'"$RPCPORT_1"'"#g' $CHAIN_DIR/$CHAINID_1/config/config.toml
sed -i -e 's/timeout_commit = "5s"/timeout_commit = "1s"/g' $CHAIN_DIR/$CHAINID_1/config/config.toml
sed -i -e 's/timeout_propose = "3s"/timeout_propose = "1s"/g' $CHAIN_DIR/$CHAINID_1/config/config.toml
sed -i -e 's/index_all_keys = false/index_all_keys = true/g' $CHAIN_DIR/$CHAINID_1/config/config.toml
sed -i -e 's/enable = false/enable = true/g' $CHAIN_DIR/$CHAINID_1/config/app.toml
sed -i -e 's/swagger = false/swagger = true/g' $CHAIN_DIR/$CHAINID_1/config/app.toml
sed -i -e 's#"tcp://0.0.0.0:1317"#"tcp://0.0.0.0:'"$RESTPORT_1"'"#g' $CHAIN_DIR/$CHAINID_1/config/app.toml
sed -i -e 's#":8080"#":'"$ROSETTA_1"'"#g' $CHAIN_DIR/$CHAINID_1/config/app.toml
sed -i -e 's/minimum-gas-prices = ""/minimum-gas-prices = "0.0025stake"/g' $CHAIN_DIR/$CHAINID_1/config/app.toml
sed -i -e 's/enabled = false/enabled = true/g' $CHAIN_DIR/$CHAINID_1/config/app.toml
sed -i -e 's/prometheus-retention-time = 0/prometheus-retention-time = 1000/g' $CHAIN_DIR/$CHAINID_1/config/app.toml


sed -i -e 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:'"$P2PPORT_2"'"#g' $CHAIN_DIR/$CHAINID_2/config/config.toml
sed -i -e 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:'"$RPCPORT_2"'"#g' $CHAIN_DIR/$CHAINID_2/config/config.toml
sed -i -e 's/timeout_commit = "5s"/timeout_commit = "1s"/g' $CHAIN_DIR/$CHAINID_2/config/config.toml
sed -i -e 's/timeout_propose = "3s"/timeout_propose = "1s"/g' $CHAIN_DIR/$CHAINID_2/config/config.toml
sed -i -e 's/index_all_keys = false/index_all_keys = true/g' $CHAIN_DIR/$CHAINID_2/config/config.toml
sed -i -e 's/enable = false/enable = true/g' $CHAIN_DIR/$CHAINID_2/config/app.toml
sed -i -e 's/swagger = false/swagger = true/g' $CHAIN_DIR/$CHAINID_2/config/app.toml
sed -i -e 's#"tcp://0.0.0.0:1317"#"tcp://0.0.0.0:'"$RESTPORT_2"'"#g' $CHAIN_DIR/$CHAINID_2/config/app.toml
sed -i -e 's#":8080"#":'"$ROSETTA_2"'"#g' $CHAIN_DIR/$CHAINID_2/config/app.toml
sed -i -e 's/minimum-gas-prices = ""/minimum-gas-prices = "0.0025stake"/g' $CHAIN_DIR/$CHAINID_2/config/app.toml
sed -i -e 's/enabled = false/enabled = true/g' $CHAIN_DIR/$CHAINID_2/config/app.toml
sed -i -e 's/prometheus-retention-time = 0/prometheus-retention-time = 1000/g' $CHAIN_DIR/$CHAINID_2/config/app.toml

# Update host chain genesis to allow x/bank/MsgSend ICA tx execution
sed -i -e 's/\"allow_messages\":.*/\"allow_messages\": [\"\/cosmos.bank.v1beta1.MsgSend\", \"\/cosmos.staking.v1beta1.MsgDelegate\", \"\/cosmos.staking.v1beta1.MsgUndelegate\"]/g' $CHAIN_DIR/$CHAINID_2/config/genesis.json
sed -i -e 's/\"admins\":.*/\"admins\": [\"neutron1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrqcd0mrx\"]/g' $CHAIN_DIR/$CHAINID_1/config/genesis.json
