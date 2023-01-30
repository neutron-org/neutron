#!/bin/bash

ADMIN_ADDRESS=neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2
NEUTROND_BINARY=neutrond
GAIAD_BINARY=gaiad
CHAIN_DIR=./data
CHAINID_1=test-1
CHAINID_2=test-2
DAO_CONTRACT=./contracts/cwd_core.wasm
PRE_PROPOSAL_CONTRACT=./contracts/cwd_pre_propose_single.wasm
PROPOSAL_CONTRACT=./contracts/cwd_proposal_single.wasm
VOTING_REGISTRY_CONTRACT=./contracts/neutron_voting_registry.wasm
VAULT_CONTRACT=./contracts/neutron_vault.wasm
TREASURY_CONTRACT=./contracts/neutron_treasury.wasm
DISTRIBUTION_CONTRACT=./contracts/neutron_distribution.wasm
PROPOSAL_MULTIPLE_CONTRACT=./contracts/cwd_proposal_multiple.wasm
PRE_PROPOSAL_MULTIPLE_CONTRACT=./contracts/cwd_pre_propose_multiple.wasm
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
if pgrep -x "$NEUTROND_BINARY" >/dev/null; then
    echo "Terminating $NEUTROND_BINARY..."
    killall $NEUTROND_BINARY
fi

# Stop if it is already running
if pgrep -x "$GAIAD_BINARY" >/dev/null; then
    echo "Terminating $GAIAD_BINARY..."
    killall $GAIAD_BINARY
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
$NEUTROND_BINARY init test --home $CHAIN_DIR/$CHAINID_1 --chain-id=$CHAINID_1
$GAIAD_BINARY init test --home $CHAIN_DIR/$CHAINID_2 --chain-id=$CHAINID_2

echo "Adding genesis accounts..."
echo $VAL_MNEMONIC_1 | $NEUTROND_BINARY keys add val1 --home $CHAIN_DIR/$CHAINID_1 --recover --keyring-backend=test
echo $VAL_MNEMONIC_2 | $GAIAD_BINARY keys add val2 --home $CHAIN_DIR/$CHAINID_2 --recover --keyring-backend=test
echo $DEMO_MNEMONIC_1 | $NEUTROND_BINARY keys add demowallet1 --home $CHAIN_DIR/$CHAINID_1 --recover --keyring-backend=test
echo $DEMO_MNEMONIC_2 | $GAIAD_BINARY keys add demowallet2 --home $CHAIN_DIR/$CHAINID_2 --recover --keyring-backend=test
echo $DEMO_MNEMONIC_3 | $NEUTROND_BINARY keys add demowallet3 --home $CHAIN_DIR/$CHAINID_1 --recover --keyring-backend=test
echo $RLY_MNEMONIC_1 | $NEUTROND_BINARY keys add rly1 --home $CHAIN_DIR/$CHAINID_1 --recover --keyring-backend=test
echo $RLY_MNEMONIC_2 | $GAIAD_BINARY keys add rly2 --home $CHAIN_DIR/$CHAINID_2 --recover --keyring-backend=test

$NEUTROND_BINARY add-genesis-account $($NEUTROND_BINARY --home $CHAIN_DIR/$CHAINID_1 keys show val1 --keyring-backend test -a) 100000000000untrn  --home $CHAIN_DIR/$CHAINID_1
$GAIAD_BINARY add-genesis-account $($GAIAD_BINARY --home $CHAIN_DIR/$CHAINID_2 keys show val2 --keyring-backend test -a) 100000000000stake  --home $CHAIN_DIR/$CHAINID_2
$NEUTROND_BINARY add-genesis-account $($NEUTROND_BINARY --home $CHAIN_DIR/$CHAINID_1 keys show demowallet1 --keyring-backend test -a) 100000000000untrn  --home $CHAIN_DIR/$CHAINID_1
$GAIAD_BINARY add-genesis-account $($GAIAD_BINARY --home $CHAIN_DIR/$CHAINID_2 keys show demowallet2 --keyring-backend test -a) 100000000000stake  --home $CHAIN_DIR/$CHAINID_2
$NEUTROND_BINARY add-genesis-account $($NEUTROND_BINARY --home $CHAIN_DIR/$CHAINID_1 keys show demowallet3 --keyring-backend test -a) 100000000000untrn  --home $CHAIN_DIR/$CHAINID_1
$NEUTROND_BINARY add-genesis-account $($NEUTROND_BINARY --home $CHAIN_DIR/$CHAINID_1 keys show rly1 --keyring-backend test -a) 100000000000untrn  --home $CHAIN_DIR/$CHAINID_1
$GAIAD_BINARY add-genesis-account $($GAIAD_BINARY --home $CHAIN_DIR/$CHAINID_2 keys show rly2 --keyring-backend test -a) 100000000000stake  --home $CHAIN_DIR/$CHAINID_2

echo "Initializing dao contract in genesis..."

function store_binary() {
  CONTRACT_BINARY_PATH=$1
  $NEUTROND_BINARY add-wasm-message store "$CONTRACT_BINARY_PATH" --output json --run-as ${ADMIN_ADDRESS} --keyring-backend=test --home $CHAIN_DIR/$CHAINID_1
  echo $(jq -r "[.app_state.wasm.gen_msgs[] | select(.store_code != null)] | length" $CHAIN_DIR/$CHAINID_1/config/genesis.json)
}

# Upload the dao contracts

VAULT_CONTRACT_BINARY_ID=$(store_binary ${VAULT_CONTRACT})
DAO_CONTRACT_BINARY_ID=$(store_binary ${DAO_CONTRACT})
PROPOSAL_CONTRACT_BINARY_ID=$(store_binary ${PROPOSAL_CONTRACT})
VOTING_REGISTRY_CONTRACT_BINARY_ID=$(store_binary ${VOTING_REGISTRY_CONTRACT})
PRE_PROPOSAL_CONTRACT_BINARY_ID=$(store_binary ${PRE_PROPOSAL_CONTRACT})
PROPOSAL_MULTIPLE_CONTRACT_BINARY_ID=$(store_binary ${PROPOSAL_MULTIPLE_CONTRACT})
PRE_PROPOSAL_MULTIPLE_CONTRACT_BINARY_ID=$(store_binary ${PRE_PROPOSAL_MULTIPLE_CONTRACT})
TREASURY_CONTRACT_BINARY_ID=$(store_binary ${TREASURY_CONTRACT})
DISTRIBUTION_CONTRACT_BINARY_ID=$(store_binary ${DISTRIBUTION_CONTRACT})

# PRE_PROPOSE_INIT_MSG will be put into the PROPOSAL_SINGLE_INIT_MSG and PROPOSAL_MULTIPLE_INIT_MSG
PRE_PROPOSE_INIT_MSG='{
   "deposit_info":{
      "denom":{
         "token":{
            "denom":{
               "native":"untrn"
            }
         }
      },
     "amount": "1000",
     "refund_policy":"always"
   },
   "open_proposal_submission":false
}'
PRE_PROPOSE_INIT_MSG_BASE64=$(echo ${PRE_PROPOSE_INIT_MSG} | base64 | tr -d "\n")

# -------------------- PROPOSE-SINGLE { PRE-PROPOSE } --------------------

PROPOSAL_SINGLE_INIT_MSG='{
   "allow_revoting":false,
   "pre_propose_info":{
      "ModuleMayPropose":{
         "info":{
            "code_id": '"${PRE_PROPOSAL_CONTRACT_BINARY_ID}"',
            "msg": "'"${PRE_PROPOSE_INIT_MSG_BASE64}"'",
            "label":"neutron"
         }
      }
   },
   "only_members_execute":false,
   "max_voting_period":{
      "time":604800
   },
   "close_proposal_on_execution_failure":false,
   "threshold":{
      "threshold_quorum":{
         "quorum":{
            "percent":"0.20"
         },
         "threshold":{
            "majority":{

            }
         }
      }
   }
}'
PROPOSAL_SINGLE_INIT_MSG_BASE64=$(echo ${PROPOSAL_SINGLE_INIT_MSG} | base64 | tr -d "\n")

# -------------------- PROPOSE-MULTIPLE { PRE-PROPOSE } --------------------

PROPOSAL_MULTIPLE_INIT_MSG='{
   "allow_revoting":false,
   "pre_propose_info":{
      "ModuleMayPropose":{
         "info":{
            "code_id": '"${PRE_PROPOSAL_MULTIPLE_CONTRACT_BINARY_ID}"',
            "msg": "'"${PRE_PROPOSE_INIT_MSG_BASE64}"'",
            "label":"neutron"
         }
      }
   },
   "only_members_execute":false,
   "max_voting_period":{
      "time":604800
   },
   "close_proposal_on_execution_failure":false,
   "voting_strategy":{
     "single_choice": {
        "quorum": {
          "majority": {
          }
        }
     }
   }
}'
PROPOSAL_MULTIPLE_INIT_MSG_BASE64=$(echo ${PROPOSAL_MULTIPLE_INIT_MSG} | base64 | tr -d "\n")

VOTING_REGISTRY_INIT_MSG='{
  "manager": null,
  "owner": null,
  "voting_vault": "neutron14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s5c2epq"
}'
VOTING_REGISTRY_INIT_MSG_BASE64=$(echo ${VOTING_REGISTRY_INIT_MSG} | base64 | tr -d "\n")

INIT='{
  "denom":"untrn",
  "description": "based neutron vault"
}'
DAO_INIT='{
  "description": "basic neutron dao",
  "name": "Neutron",
  "initial_items": null,
  "proposal_modules_instantiate_info": [
    {
      "code_id": '"${PROPOSAL_CONTRACT_BINARY_ID}"',
      "label": "DAO_Neutron_cw-proposal-single",
      "msg": "'"${PROPOSAL_SINGLE_INIT_MSG_BASE64}"'"
    },
    {
      "code_id": '"${PROPOSAL_MULTIPLE_CONTRACT_BINARY_ID}"',
      "label": "DAO_Neutron_cw-proposal-multiple",
      "msg": "'"${PROPOSAL_MULTIPLE_INIT_MSG_BASE64}"'"
    }
  ],
  "voting_registry_module_instantiate_info": {
    "code_id": '"${VOTING_REGISTRY_CONTRACT_BINARY_ID}"',
    "label": "DAO_Neutron_voting_registry",
    "msg": "'"${VOTING_REGISTRY_INIT_MSG_BASE64}"'"
  }
}'

# TODO: properly initialize treasury
DISTRIBUTION_CONTRACT_ADDRESS="neutron1vhndln95yd7rngslzvf6sax6axcshkxqpmpr886ntelh28p9ghuq56mwja"
TREASURY_INIT="$(printf '{
  "main_dao_address": "%s",
  "security_dao_address": "%s",
  "denom": "untrn",
  "distribution_rate": "0",
  "min_period": 10,
  "distribution_contract": "%s",
  "reserve_contract": "%s",
  "vesting_denominator": "1"
}' "$ADMIN_ADDRESS" "$ADMIN_ADDRESS" "$DISTRIBUTION_CONTRACT_ADDRESS" "$ADMIN_ADDRESS")"

DISTRIBUTION_INIT="$(printf '{
                           "main_dao_address": "%s",
                           "security_dao_address": "%s",
                           "denom": "untrn"
}' "$ADMIN_ADDRESS" "$ADMIN_ADDRESS")"

echo "Instantiate contracts"
$NEUTROND_BINARY add-wasm-message instantiate-contract "$VAULT_CONTRACT_BINARY_ID" "$INIT" --run-as ${ADMIN_ADDRESS} --admin ${ADMIN_ADDRESS}  --label "DAO_Neutron_voting_vault"  --home $CHAIN_DIR/$CHAINID_1
$NEUTROND_BINARY add-wasm-message instantiate-contract "$DAO_CONTRACT_BINARY_ID" "$DAO_INIT" --run-as ${ADMIN_ADDRESS} --admin ${ADMIN_ADDRESS}  --label "DAO"  --home $CHAIN_DIR/$CHAINID_1
$NEUTROND_BINARY add-wasm-message instantiate-contract "$TREASURY_CONTRACT_BINARY_ID" "$TREASURY_INIT" --run-as ${ADMIN_ADDRESS} --admin ${ADMIN_ADDRESS} --label "Treasury" --home $CHAIN_DIR/$CHAINID_1
$NEUTROND_BINARY add-wasm-message instantiate-contract "$DISTRIBUTION_CONTRACT_BINARY_ID" "$DISTRIBUTION_INIT" --run-as ${ADMIN_ADDRESS} --admin ${ADMIN_ADDRESS} --label "Distribution" --home $CHAIN_DIR/$CHAINID_1


echo "Add consumer section..."
$NEUTROND_BINARY add-consumer-section --home $CHAIN_DIR/$CHAINID_1

echo "Creating and collecting gaiad network gentx..."
$GAIAD_BINARY gentx val2 7000000000stake --home $CHAIN_DIR/$CHAINID_2 --chain-id $CHAINID_2 --keyring-backend test
$GAIAD_BINARY collect-gentxs --home $CHAIN_DIR/$CHAINID_2

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
sed -i -e 's/minimum-gas-prices = ""/minimum-gas-prices = "0.0025untrn,0.0025ibc\/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"/g' $CHAIN_DIR/$CHAINID_1/config/app.toml
sed -i -e 's/enabled = false/enabled = true/g' $CHAIN_DIR/$CHAINID_1/config/app.toml
sed -i -e 's/prometheus-retention-time = 0/prometheus-retention-time = 1000/g' $CHAIN_DIR/$CHAINID_1/config/app.toml


sed -i -e 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:'"$P2PPORT_2"'"#g' $CHAIN_DIR/$CHAINID_2/config/config.toml
sed -i -e 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:'"$RPCPORT_2"'"#g' $CHAIN_DIR/$CHAINID_2/config/config.toml
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
