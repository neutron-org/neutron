#!/bin/bash
set -e

BINARY=${BINARY:-neutrond}
BASE_DIR=./data
CHAINID=${CHAINID:-test-1}
STAKEDENOM=${STAKEDENOM:-untrn}
CONTRACTS_BINARIES_DIR=${CONTRACTS_BINARIES_DIR:-./contracts}

CHAIN_DIR="$BASE_DIR/$CHAINID"

ADMIN_ADDRESS=neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2
DAO_CONTRACT=$CONTRACTS_BINARIES_DIR/cwd_core.wasm
PRE_PROPOSAL_CONTRACT=$CONTRACTS_BINARIES_DIR/cwd_pre_propose_single.wasm
PROPOSAL_CONTRACT=$CONTRACTS_BINARIES_DIR/cwd_proposal_single.wasm
VOTING_REGISTRY_CONTRACT=$CONTRACTS_BINARIES_DIR/neutron_voting_registry.wasm
NEUTRON_VAULT_CONTRACT=$CONTRACTS_BINARIES_DIR/neutron_vault.wasm
LOCKDROP_VAULT_CONTRACT=$CONTRACTS_BINARIES_DIR/lockdrop_vault.wasm
PROPOSAL_MULTIPLE_CONTRACT=$CONTRACTS_BINARIES_DIR/cwd_proposal_multiple.wasm
PRE_PROPOSAL_MULTIPLE_CONTRACT=$CONTRACTS_BINARIES_DIR/cwd_pre_propose_multiple.wasm
TREASURY_CONTRACT=$CONTRACTS_BINARIES_DIR/neutron_treasury.wasm
DISTRIBUTION_CONTRACT=$CONTRACTS_BINARIES_DIR/neutron_distribution.wasm

echo "Add consumer section..."
$BINARY add-consumer-section --home "$CHAIN_DIR"

echo "Initializing dao contract in genesis..."

function store_binary() {
  CONTRACT_BINARY_PATH=$1
  $BINARY add-wasm-message store "$CONTRACT_BINARY_PATH" --output json --run-as ${ADMIN_ADDRESS} --keyring-backend=test --home "$CHAIN_DIR"
  BINARY_ID=$(jq -r "[.app_state.wasm.gen_msgs[] | select(.store_code != null)] | length" "$CHAIN_DIR/config/genesis.json")
  echo "$BINARY_ID"
}

# Upload the dao contracts

NEUTRON_VAULT_CONTRACT_BINARY_ID=$(store_binary         "$NEUTRON_VAULT_CONTRACT")
DAO_CONTRACT_BINARY_ID=$(store_binary                   "$DAO_CONTRACT")
PROPOSAL_CONTRACT_BINARY_ID=$(store_binary              "$PROPOSAL_CONTRACT")
VOTING_REGISTRY_CONTRACT_BINARY_ID=$(store_binary       "$VOTING_REGISTRY_CONTRACT")
PRE_PROPOSAL_CONTRACT_BINARY_ID=$(store_binary          "$PRE_PROPOSAL_CONTRACT")
PROPOSAL_MULTIPLE_CONTRACT_BINARY_ID=$(store_binary     "$PROPOSAL_MULTIPLE_CONTRACT")
PRE_PROPOSAL_MULTIPLE_CONTRACT_BINARY_ID=$(store_binary "$PRE_PROPOSAL_MULTIPLE_CONTRACT")
TREASURY_CONTRACT_BINARY_ID=$(store_binary              "$TREASURY_CONTRACT")
DISTRIBUTION_CONTRACT_BINARY_ID=$(store_binary          "$DISTRIBUTION_CONTRACT")
LOCKDROP_VAULT_CONTRACT_BINARY_ID=$(store_binary        "$LOCKDROP_VAULT_CONTRACT")

# PRE_PROPOSE_INIT_MSG will be put into the PROPOSAL_SINGLE_INIT_MSG and PROPOSAL_MULTIPLE_INIT_MSG
PRE_PROPOSE_INIT_MSG='{
   "deposit_info":{
      "denom":{
         "token":{
            "denom":{
               "native":"'"$STAKEDENOM"'"
            }
         }
      },
     "amount": "1000",
     "refund_policy":"always"
   },
   "open_proposal_submission":false
}'
PRE_PROPOSE_INIT_MSG_BASE64=$(echo "$PRE_PROPOSE_INIT_MSG" | base64 | tr -d "\n")

# -------------------- PROPOSE-SINGLE { PRE-PROPOSE } --------------------

PROPOSAL_SINGLE_INIT_MSG='{
   "allow_revoting":false,
   "pre_propose_info":{
      "module_may_propose":{
         "info":{
            "code_id": '"$PRE_PROPOSAL_CONTRACT_BINARY_ID"',
            "msg": "'"$PRE_PROPOSE_INIT_MSG_BASE64"'",
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
PROPOSAL_SINGLE_INIT_MSG_BASE64=$(echo "$PROPOSAL_SINGLE_INIT_MSG" | base64 | tr -d "\n")

# -------------------- PROPOSE-MULTIPLE { PRE-PROPOSE } --------------------

PROPOSAL_MULTIPLE_INIT_MSG='{
   "allow_revoting":false,
   "pre_propose_info":{
      "module_may_propose":{
         "info":{
            "code_id": '"$PRE_PROPOSAL_MULTIPLE_CONTRACT_BINARY_ID"',
            "msg": "'"$PRE_PROPOSE_INIT_MSG_BASE64"'",
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
PROPOSAL_MULTIPLE_INIT_MSG_BASE64=$(echo "$PROPOSAL_MULTIPLE_INIT_MSG" | base64 | tr -d "\n")

VOTING_REGISTRY_INIT_MSG='{
  "manager": null,
  "owner": null,
  "voting_vaults": [
    "neutron14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s5c2epq",
    "neutron13we0myxwzlpx8l5ark8elw5gj5d59dl6cjkzmt80c5q5cv5rt54qvzkv2a"
  ]
}'
VOTING_REGISTRY_INIT_MSG_BASE64=$(echo "$VOTING_REGISTRY_INIT_MSG" | base64 | tr -d "\n")

DAO_INIT='{
  "description": "basic neutron dao",
  "name": "Neutron",
  "initial_items": null,
  "proposal_modules_instantiate_info": [
    {
      "code_id": '"$PROPOSAL_CONTRACT_BINARY_ID"',
      "label": "DAO_Neutron_cw-proposal-single",
      "msg": "'"$PROPOSAL_SINGLE_INIT_MSG_BASE64"'"
    },
    {
      "code_id": '"$PROPOSAL_MULTIPLE_CONTRACT_BINARY_ID"',
      "label": "DAO_Neutron_cw-proposal-multiple",
      "msg": "'"$PROPOSAL_MULTIPLE_INIT_MSG_BASE64"'"
    }
  ],
  "voting_registry_module_instantiate_info": {
    "code_id": '"$VOTING_REGISTRY_CONTRACT_BINARY_ID"',
    "label": "DAO_Neutron_voting_registry",
    "msg": "'"$VOTING_REGISTRY_INIT_MSG_BASE64"'"
  }
}'

# TODO: properly initialize treasury
DISTRIBUTION_CONTRACT_ADDRESS="neutron1vhndln95yd7rngslzvf6sax6axcshkxqpmpr886ntelh28p9ghuq56mwja"
TREASURY_INIT='{
  "main_dao_address": "'"$ADMIN_ADDRESS"'",
  "security_dao_address": "'"$ADMIN_ADDRESS"'",
  "denom": "'"$STAKEDENOM"'",
  "distribution_rate": "0",
  "min_period": 10,
  "distribution_contract": "'"$DISTRIBUTION_CONTRACT_ADDRESS"'",
  "reserve_contract": "'"$ADMIN_ADDRESS"'",
  "vesting_denominator": "1"
}'

DISTRIBUTION_INIT='{
  "main_dao_address": "'"$ADMIN_ADDRESS"'",
  "security_dao_address": "'"$ADMIN_ADDRESS"'",
  "denom": "'"$STAKEDENOM"'"
}'

NEUTRON_VAULT_INIT='{
  "name": "voting vault",
  "denom": "'"$STAKEDENOM"'",
  "description": "a simple voting vault for testing purposes"
}'
# since the lockdrop_contract is still a mock, the address is a random valid one just to pass instantiation
LOCKDROP_VAULT_INIT='{
  "name": "lockdrop vault",
  "description": "a lockdrop vault for testing purposes",
  "lockdrop_contract": "neutron17zayzl5d0daqa89csvv8kqayxzke6jd6zh00tq"
}'

echo "Instantiate contracts"
$BINARY add-wasm-message instantiate-contract "$NEUTRON_VAULT_CONTRACT_BINARY_ID" "$NEUTRON_VAULT_INIT" --run-as ${ADMIN_ADDRESS} --admin ${ADMIN_ADDRESS}  --label "DAO_Neutron_voting_vault"  --home "$CHAIN_DIR"
$BINARY add-wasm-message instantiate-contract "$DAO_CONTRACT_BINARY_ID" "$DAO_INIT" --run-as ${ADMIN_ADDRESS} --admin ${ADMIN_ADDRESS}  --label "DAO"  --home "$CHAIN_DIR"
$BINARY add-wasm-message instantiate-contract "$TREASURY_CONTRACT_BINARY_ID" "$TREASURY_INIT" --run-as ${ADMIN_ADDRESS} --admin ${ADMIN_ADDRESS} --label "Treasury" --home "$CHAIN_DIR"
$BINARY add-wasm-message instantiate-contract "$DISTRIBUTION_CONTRACT_BINARY_ID" "$DISTRIBUTION_INIT" --run-as ${ADMIN_ADDRESS} --admin ${ADMIN_ADDRESS} --label "Distribution" --home "$CHAIN_DIR"
$BINARY add-wasm-message instantiate-contract "$LOCKDROP_VAULT_CONTRACT_BINARY_ID" "$LOCKDROP_VAULT_INIT" --run-as ${ADMIN_ADDRESS} --admin ${ADMIN_ADDRESS}  --label "DAO_Neutron_lockdrop_vault"  --home "$CHAIN_DIR"

sed -i -e 's/\"admins\":.*/\"admins\": [\"neutron1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrqcd0mrx\"]/g' "$CHAIN_DIR/config/genesis.json"
sed -i -e 's/\"treasury_address\":.*/\"treasury_address\":\"neutron1vguuxez2h5ekltfj9gjd62fs5k4rl2zy5hfrncasykzw08rezpfsd2rhm7\"/g' "$CHAIN_DIR/config/genesis.json"

