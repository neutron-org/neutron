#!/bin/bash
set -e

BINARY=${BINARY:-neutrond}
BASE_DIR=./data
CHAINID=${CHAINID:-test-1}
STAKEDENOM=${STAKEDENOM:-untrn}
CONTRACTS_BINARIES_DIR=${CONTRACTS_BINARIES_DIR:-./contracts}
THIRD_PARTY_CONTRACTS_DIR=${THIRD_PARTY_CONTRACTS_DIR:-./contracts_thirdparty}

# IMPORTANT! minimum_gas_prices should always contain at least one record, otherwise the chain will not start or halt
# ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2 denom is required by intgration tests (test:tokenomics)
MIN_GAS_PRICES_DEFAULT='[{"denom":"ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2","amount":"0"},{"denom":"untrn","amount":"0"}],'
MIN_GAS_PRICES=${MIN_GAS_PRICES:-"$MIN_GAS_PRICES_DEFAULT"}

BYPASS_MIN_FEE_MSG_TYPES_DEFAULT='["/ibc.core.channel.v1.Msg/RecvPacket", "/ibc.core.channel.v1.Msg/Acknowledgement", "/ibc.core.client.v1.Msg/UpdateClient"]'
BYPASS_MIN_FEE_MSG_TYPES=${BYPASS_MIN_FEE_MSG_TYPES:-"$BYPASS_MIN_FEE_MSG_TYPES_DEFAULT"}

MAX_TOTAL_BYPASS_MIN_FEE_MSG_GAS_USAGE_DEFAULT=1000000
MAX_TOTAL_BYPASS_MIN_FEE_MSG_GAS_USAGE=${MAX_TOTAL_BYPASS_MIN_FEE_MSG_GAS_USAGE:-"$MAX_TOTAL_BYPASS_MIN_FEE_MSG_GAS_USAGE_DEFAULT"}

CHAIN_DIR="$BASE_DIR/$CHAINID"
GENESIS_PATH="$CHAIN_DIR/config/genesis.json"

ADMIN_ADDRESS=$($BINARY keys show demowallet1 -a --home "$CHAIN_DIR" --keyring-backend test)
# MAIN_DAO
DAO_CONTRACT=$CONTRACTS_BINARIES_DIR/cwd_core.wasm
PRE_PROPOSAL_CONTRACT=$CONTRACTS_BINARIES_DIR/cwd_pre_propose_single.wasm
PRE_PROPOSAL_MULTIPLE_CONTRACT=$CONTRACTS_BINARIES_DIR/cwd_pre_propose_multiple.wasm
PRE_PROPOSAL_OVERRULE_CONTRACT=$CONTRACTS_BINARIES_DIR/cwd_pre_propose_overrule.wasm
PROPOSAL_CONTRACT=$CONTRACTS_BINARIES_DIR/cwd_proposal_single.wasm
PROPOSAL_MULTIPLE_CONTRACT=$CONTRACTS_BINARIES_DIR/cwd_proposal_multiple.wasm
VOTING_REGISTRY_CONTRACT=$CONTRACTS_BINARIES_DIR/neutron_voting_registry.wasm
# VAULTS
NEUTRON_VAULT_CONTRACT=$CONTRACTS_BINARIES_DIR/neutron_vault.wasm
# RESERVE
RESERVE_CONTRACT=$CONTRACTS_BINARIES_DIR/neutron_reserve.wasm
DISTRIBUTION_CONTRACT=$CONTRACTS_BINARIES_DIR/neutron_distribution.wasm
# SUBDAOS
SUBDAO_CORE_CONTRACT=$CONTRACTS_BINARIES_DIR/cwd_subdao_core.wasm
SUBDAO_TIMELOCK_CONTRACT=$CONTRACTS_BINARIES_DIR/cwd_subdao_timelock_single.wasm
SUBDAO_PRE_PROPOSE_CONTRACT=$CONTRACTS_BINARIES_DIR/cwd_subdao_pre_propose_single.wasm
SUBDAO_PROPOSAL_CONTRACT=$CONTRACTS_BINARIES_DIR/cwd_subdao_proposal_single.wasm
CW4_VOTING_CONTRACT=$THIRD_PARTY_CONTRACTS_DIR/cw4_voting.wasm
CW4_GROUP_CONTRACT=$THIRD_PARTY_CONTRACTS_DIR/cw4_group.wasm

echo "Add consumer section..."
$BINARY add-consumer-section --home "$CHAIN_DIR"
### PARAMETERS SECTION

## slashing params
SLASHING_SIGNED_BLOCKS_WINDOW=140000
SLASHING_MIN_SIGNED=0.050000000000000000
SLASHING_FRACTION_DOUBLE_SIGN=0.010000000000000000
SLASHING_FRACTION_DOWNTIME=0.000100000000000000

##pre propose single parameters
PRE_PROPOSAL_SINGLE_AMOUNT=1000
PRE_PROPOSAL_SINGLE_REFUND_POLICY="only_passed"
PRE_PROPOSAL_SINGLE_OPEN_PROPOSAL_SUBMISSION=false

## proposal singe params
PROPOSAL_ALLOW_REVOTING=false # should be true for non-testing env
PROPOSAL_SINGLE_ONLY_MEMBERS_EXECUTE=false
PROPOSAL_SINGLE_ONLY_MAX_VOTING_PERIOD=1200 # seconds; should be 2 weeks in production
PROPOSAL_SINGLE_CLOSE_PROPOSAL_ON_EXECUTION_FAILURE=false
PROPOSAL_SINGLE_QUORUM=0.05 # quorum to consider proposal's result viable [float] < 1
PROPOSAL_SINGLE_THRESHOLD=0.5 # % of votes should vote for the proposal to pass [float] <1
PROPOSAL_SINGLE_LABEL="neutron.proposals.single"
PRE_PROPOSAL_SINGLE_LABEL="neutron.proposals.single.pre_propose"

## propose multiple params
PROPOSAL_MULTIPLE_ALLOW_REVOTING=false # should be true for non-testing env
PROPOSAL_MULTIPLE_ONLY_MEMBERS_EXECUTE=false
PROPOSAL_MULTIPLE_ONLY_MAX_VOTING_PERIOD=1200 # seconds; should be 2 weeks in production
PROPOSAL_MULTIPLE_CLOSE_PROPOSAL_ON_EXECUTION_FAILURE=false
PROPOSAL_MULTIPLE_QUORUM=0.05 # quorum to consider proposal's result viable [float] < 1
PROPOSAL_MULTIPLE_LABEL="neutron.proposals.multiple"
PRE_PROPOSAL_MULTIPLE_LABEL="neutron.proposals.multiple.pre_propose"

## Propose overrule params
PROPOSAL_OVERRULE_ALLOW_REVOTING=false
PROPOSAL_OVERRULE_ONLY_MEMBERS_EXECUTE=false
PROPOSAL_OVERRULE_ONLY_MAX_VOTING_PERIOD=1200 # seconds; should be 3 days in production
PROPOSAL_OVERRULE_CLOSE_PROPOSAL_ON_EXECUTION_FAILURE=false
PROPOSAL_OVERRULE_THRESHOLD=0.005 # around 10 times lower than for regular proposals
PROPOSAL_OVERRULE_LABEL="neutron.proposals.overrule"
PRE_PROPOSE_OVERRULE_LABEL="neutron.proposals.overrule.pre_propose"

## Voting registry
VOTING_REGISTRY_LABEL="neutron.voting"

## DAO
DAO_NAME="Neutron DAO"
DAO_DESCRIPTION="Neutron DAO is a DAO DAO-based governance of Neutron chain"
DAO_CORE_LABEL="neutron.core"

## Neutron vault
NEUTRON_VAULT_NAME="Neutron Vault"
NEUTRON_VAULT_DESCRIPTION="Vault to put NTRN tokens to get voting power"
NEUTRON_VAULT_LABEL="neutron.voting.vaults.neutron"

## Reserve
RESERVE_DISTRIBUTION_RATE=0
RESERVE_MIN_PERIOD=10
RESERVE_VESTING_DENOMINATOR=1
RESERVE_LABEL="reserve"

DISTRIBUTION_LABEL="distribution"

## Grants subdao
GRANTS_SUBDAO_CORE_NAME="Grants SubDAO"
GRANTS_SUBDAO_CORE_DESCRIPTION="SubDAO to distribute grants to projects"
GRANTS_SUBDAO_CORE_LABEL="neutron.subdaos.grants.core"
GRANTS_SUBDAO_PROPOSAL_LABEL="neutron.subdaos.grants.proposals.single"
GRANTS_SUBDAO_PRE_PROPOSE_LABEL="neutron.subdaos.grants.proposals.single.pre_propose"
GRANTS_SUBDAO_VOTING_MODULE_LABEL="neutron.subdaos.grants.voting"

## Timelock
GRANTS_SUBDAO_TIMELOCK_LABEL="neutron.subdaos.grants.proposals.single.pre_propose.timelock"

## Security subdao
SECURITY_SUBDAO_CORE_NAME="Security SubDAO"
SECURITY_SUBDAO_CORE_DESCRIPTION="SubDAO with power to pause specific Neutron DAO modules"
SECURITY_SUBDAO_CORE_LABEL="neutron.subdaos.security.core"
SECURITY_SUBDAO_PROPOSAL_LABEL="neutron.subdaos.security.proposals.single"
SECURITY_SUBDAO_PRE_PROPOSE_LABEL="neutron.subdaos.security.proposals.single.pre_propose"
SECURITY_SUBDAO_VOTE_LABEL="neutron.subdaos.security.voting"

echo "Initializing dao contract in genesis..."

function store_binary() {
  CONTRACT_BINARY_PATH=$1
  $BINARY add-wasm-message store "$CONTRACT_BINARY_PATH" \
    --output json --run-as "${ADMIN_ADDRESS}" --keyring-backend=test --home "$CHAIN_DIR"
  BINARY_ID=$(jq -r "[.app_state.wasm.gen_msgs[] | select(.store_code != null)] | length" "$CHAIN_DIR/config/genesis.json")
  echo "$BINARY_ID"
}

# Upload the dao contracts
# MAIN_DAO
DAO_CONTRACT_BINARY_ID=$(store_binary                   "$DAO_CONTRACT")
PRE_PROPOSAL_CONTRACT_BINARY_ID=$(store_binary          "$PRE_PROPOSAL_CONTRACT")
PRE_PROPOSAL_MULTIPLE_CONTRACT_BINARY_ID=$(store_binary "$PRE_PROPOSAL_MULTIPLE_CONTRACT")
PRE_PROPOSAL_OVERRULE_CONTRACT_BINARY_ID=$(store_binary "$PRE_PROPOSAL_OVERRULE_CONTRACT")
PROPOSAL_CONTRACT_BINARY_ID=$(store_binary              "$PROPOSAL_CONTRACT")
PROPOSAL_MULTIPLE_CONTRACT_BINARY_ID=$(store_binary     "$PROPOSAL_MULTIPLE_CONTRACT")
VOTING_REGISTRY_CONTRACT_BINARY_ID=$(store_binary       "$VOTING_REGISTRY_CONTRACT")
# VAULTS
NEUTRON_VAULT_CONTRACT_BINARY_ID=$(store_binary         "$NEUTRON_VAULT_CONTRACT")
# RESERVE
DISTRIBUTION_CONTRACT_BINARY_ID=$(store_binary          "$DISTRIBUTION_CONTRACT")
RESERVE_CONTRACT_BINARY_ID=$(store_binary               "$RESERVE_CONTRACT")
# SUBDAOS
SUBDAO_CORE_BINARY_ID=$(store_binary                    "$SUBDAO_CORE_CONTRACT")
SUBDAO_TIMELOCK_BINARY_ID=$(store_binary                "$SUBDAO_TIMELOCK_CONTRACT")
SUBDAO_PRE_PROPOSE_BINARY_ID=$(store_binary             "$SUBDAO_PRE_PROPOSE_CONTRACT")
SUBDAO_PROPOSAL_BINARY_ID=$(store_binary                "$SUBDAO_PROPOSAL_CONTRACT")
CW4_VOTING_CONTRACT_BINARY_ID=$(store_binary            "$CW4_VOTING_CONTRACT")
CW4_GROUP_CONTRACT_BINARY_ID=$(store_binary             "$CW4_GROUP_CONTRACT")

# WARNING!
# The following code is needed to pre-generate the contract addresses
# Those addresses depend on the ORDER OF CONTRACTS INITIALIZATION
# Thus, this code section depends a lot on the order and content of the instantiate-contract commands at the end script
# It also depends on the implicitly initialized contracts (e.g. DAO core instantiation also instantiate proposals and stuff)
# If you're to do any changes, please do it consistently in both sections
# If you're to do add any implicitly initialized contracts in init messages, please reflect changes here

function genaddr() {
  CODE_ID=$1
  CONTRACT_ADDRESS=$($BINARY debug generate-contract-address "$INSTANCE_ID_COUNTER" "$CODE_ID")
  echo "$CONTRACT_ADDRESS"
}

INSTANCE_ID_COUNTER=1

# VAULTS
NEUTRON_VAULT_CONTRACT_ADDRESS=$(genaddr                "$NEUTRON_VAULT_CONTRACT_BINARY_ID") && (( INSTANCE_ID_COUNTER++ ))

# MAIN_DAO
DAO_CONTRACT_ADDRESS=$(genaddr                          "$DAO_CONTRACT_BINARY_ID") && (( INSTANCE_ID_COUNTER++ ))
PROPOSAL_SINGLE_CONTRACT_ADDRESS=$(genaddr              "$PROPOSAL_CONTRACT_BINARY_ID") && (( INSTANCE_ID_COUNTER++ ))
PRE_PROPOSAL_CONTRACT_ADDRESS=$(genaddr                 "$PRE_PROPOSAL_CONTRACT_BINARY_ID") && (( INSTANCE_ID_COUNTER++ ))
PROPOSAL_MULTIPLE_CONTRACT_ADDRESS=$(genaddr            "$PROPOSAL_MULTIPLE_CONTRACT_BINARY_ID") && (( INSTANCE_ID_COUNTER++ ))
PRE_PROPOSAL_MULTIPLE_CONTRACT_ADDRESS=$(genaddr        "$PRE_PROPOSAL_MULTIPLE_CONTRACT_BINARY_ID") && (( INSTANCE_ID_COUNTER++ ))
PROPOSAL_OVERRULE_CONTRACT_ADDRESS=$(genaddr            "$PROPOSAL_CONTRACT_BINARY_ID") && (( INSTANCE_ID_COUNTER++ ))
PRE_PROPOSAL_OVERRULE_CONTRACT_ADDRESS=$(genaddr        "$PRE_PROPOSAL_OVERRULE_CONTRACT_BINARY_ID") && (( INSTANCE_ID_COUNTER++ ))
VOTING_REGISTRY_CONTRACT_ADDRESS=$(genaddr              "$VOTING_REGISTRY_CONTRACT_BINARY_ID") && (( INSTANCE_ID_COUNTER++ ))

# RESERVE
RESERVE_CONTRACT_ADDRESS=$(genaddr                     "$RESERVE_CONTRACT_BINARY_ID") && (( INSTANCE_ID_COUNTER++ ))
DISTRIBUTION_CONTRACT_ADDRESS=$(genaddr                "$DISTRIBUTION_CONTRACT_BINARY_ID") && (( INSTANCE_ID_COUNTER++ ))
# SUBDAOS
SECURITY_SUBDAO_CORE_CONTRACT_ADDRESS=$(genaddr        "$SUBDAO_CORE_BINARY_ID") && (( INSTANCE_ID_COUNTER++ ))
SECURITY_SUBDAO_PROPOSAL_CONTRACT_ADDRESS=$(genaddr    "$SUBDAO_PROPOSAL_BINARY_ID") && (( INSTANCE_ID_COUNTER++ ))
SECURITY_SUBDAO_PRE_PROPOSE_CONTRACT_ADDRESS=$(genaddr "$SUBDAO_PROPOSAL_BINARY_ID") && (( INSTANCE_ID_COUNTER++ ))
SECURITY_SUBDAO_VOTING_CONTRACT_ADDRESS=$(genaddr      "$CW4_VOTING_CONTRACT_BINARY_ID") && (( INSTANCE_ID_COUNTER++ ))
SECURITY_SUBDAO_GROUP_CONTRACT_ADDRESS=$(genaddr       "$CW4_GROUP_CONTRACT_BINARY_ID") && (( INSTANCE_ID_COUNTER++ ))
GRANTS_SUBDAO_CORE_CONTRACT_ADDRESS=$(genaddr          "$SUBDAO_CORE_BINARY_ID") && (( INSTANCE_ID_COUNTER++ ))
GRANTS_SUBDAO_PROPOSAL_CONTRACT_ADDRESS=$(genaddr      "$SUBDAO_PROPOSAL_BINARY_ID") && (( INSTANCE_ID_COUNTER++ ))
GRANTS_SUBDAO_PRE_PROPOSE_CONTRACT_ADDRESS=$(genaddr   "$SUBDAO_PRE_PROPOSE_BINARY_ID") && (( INSTANCE_ID_COUNTER++ ))
GRANTS_SUBDAO_TIMELOCK_CONTRACT_ADDRESS=$(genaddr      "$SUBDAO_TIMELOCK_BINARY_ID") && (( INSTANCE_ID_COUNTER++ ))
GRANTS_SUBDAO_VOTING_CONTRACT_ADDRESS=$(genaddr        "$CW4_VOTING_CONTRACT_BINARY_ID") && (( INSTANCE_ID_COUNTER++ ))
GRANTS_SUBDAO_GROUP_CONTRACT_ADDRESS=$(genaddr         "$CW4_GROUP_CONTRACT_BINARY_ID") && (( INSTANCE_ID_COUNTER++ ))

function check_json() {
  MSG=$1
  if ! jq -e . >/dev/null 2>&1 <<<"$MSG"; then
      echo "Failed to parse JSON for $MSG" >&2
      exit 1
  fi
}

function json_to_base64() {
  MSG=$1
  check_json "$MSG"
  echo "$MSG" | base64 | tr -d "\n"
}

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
     "amount": "'"$PRE_PROPOSAL_SINGLE_AMOUNT"'",
     "refund_policy":"'"$PRE_PROPOSAL_SINGLE_REFUND_POLICY"'"
   },
   "open_proposal_submission": '"$PRE_PROPOSAL_SINGLE_OPEN_PROPOSAL_SUBMISSION"'
}'
PRE_PROPOSE_INIT_MSG_BASE64=$(json_to_base64 "$PRE_PROPOSE_INIT_MSG")

# -------------------- PROPOSE-SINGLE { PRE-PROPOSE } --------------------

PROPOSAL_SINGLE_INIT_MSG='{
   "allow_revoting":'"$PROPOSAL_ALLOW_REVOTING"',
   "pre_propose_info":{
      "module_may_propose":{
         "info":{
            "admin": {
              "core_module": {}
            },
            "code_id":  '"$PRE_PROPOSAL_CONTRACT_BINARY_ID"',
            "msg":      "'"$PRE_PROPOSE_INIT_MSG_BASE64"'",
            "label":    "'"$PRE_PROPOSAL_SINGLE_LABEL"'"
         }
      }
   },
   "only_members_execute":'"$PROPOSAL_SINGLE_ONLY_MEMBERS_EXECUTE"',
   "max_voting_period":{
      "time":'"$PROPOSAL_SINGLE_ONLY_MAX_VOTING_PERIOD"'
   },
   "close_proposal_on_execution_failure":'"$PROPOSAL_SINGLE_CLOSE_PROPOSAL_ON_EXECUTION_FAILURE"',
   "threshold":{
      "threshold_quorum":{
         "quorum":{
          "percent":"'"$PROPOSAL_SINGLE_QUORUM"'"
         },
         "threshold":{
            "percent":"'"$PROPOSAL_SINGLE_THRESHOLD"'"
         }
      }
   }
}'
PROPOSAL_SINGLE_INIT_MSG_BASE64=$(json_to_base64 "$PROPOSAL_SINGLE_INIT_MSG")

# -------------------- PROPOSE-MULTIPLE { PRE-PROPOSE } --------------------

PROPOSAL_MULTIPLE_INIT_MSG='{
   "allow_revoting":'"$PROPOSAL_MULTIPLE_ALLOW_REVOTING"',
   "pre_propose_info":{
      "module_may_propose":{
         "info":{
            "admin": {
              "core_module": {}
            },
            "code_id":  '"$PRE_PROPOSAL_MULTIPLE_CONTRACT_BINARY_ID"',
            "msg":      "'"$PRE_PROPOSE_INIT_MSG_BASE64"'",
            "label":    "'"$PRE_PROPOSAL_MULTIPLE_LABEL"'"
         }
      }
   },
   "only_members_execute":'"$PROPOSAL_MULTIPLE_ONLY_MEMBERS_EXECUTE"',
   "max_voting_period":{
      "time":'"$PROPOSAL_MULTIPLE_ONLY_MAX_VOTING_PERIOD"'
   },
   "close_proposal_on_execution_failure": '"$PROPOSAL_MULTIPLE_CLOSE_PROPOSAL_ON_EXECUTION_FAILURE"',
   "voting_strategy":{
     "single_choice": {
        "quorum": {
          "percent": "'"$PROPOSAL_MULTIPLE_QUORUM"'"
        }
     }
   }
}'
PROPOSAL_MULTIPLE_INIT_MSG_BASE64=$(json_to_base64 "$PROPOSAL_MULTIPLE_INIT_MSG")

# PRE_PROPOSE_OVERRULE_INIT_MSG will be put into the PROPOSAL_OVERRULE_INIT_MSG
PRE_PROPOSE_OVERRULE_INIT_MSG='{}'
PRE_PROPOSE_OVERRULE_INIT_MSG_BASE64=$(json_to_base64 "$PRE_PROPOSE_OVERRULE_INIT_MSG")


# -------------------- PROPOSE-OVERRULE { PRE-PROPOSE-OVERRULE } --------------------

PROPOSAL_OVERRULE_INIT_MSG='{
   "allow_revoting":'"$PROPOSAL_OVERRULE_ALLOW_REVOTING"',
   "pre_propose_info":{
      "module_may_propose":{
         "info":{
            "admin": {
              "core_module": {}
            },
            "code_id":  '"$PRE_PROPOSAL_OVERRULE_CONTRACT_BINARY_ID"',
            "msg":      "'"$PRE_PROPOSE_OVERRULE_INIT_MSG_BASE64"'",
            "label":    "'"$PRE_PROPOSE_OVERRULE_LABEL"'"
         }
      }
   },
   "only_members_execute": '"$PROPOSAL_OVERRULE_ONLY_MEMBERS_EXECUTE"',
   "max_voting_period":{
      "time": '"$PROPOSAL_OVERRULE_ONLY_MAX_VOTING_PERIOD"'
   },
   "close_proposal_on_execution_failure": '"$PROPOSAL_OVERRULE_CLOSE_PROPOSAL_ON_EXECUTION_FAILURE"',
   "threshold":{
       "absolute_percentage":{
          "percentage":{
            "percent": "'"$PROPOSAL_OVERRULE_THRESHOLD"'"
          }
       }
   }
}'
PROPOSAL_OVERRULE_INIT_MSG_BASE64=$(json_to_base64 "$PROPOSAL_OVERRULE_INIT_MSG")

VOTING_REGISTRY_INIT_MSG='{
  "owner": "'"$DAO_CONTRACT_ADDRESS"'",
  "voting_vaults": [
    "'"$NEUTRON_VAULT_CONTRACT_ADDRESS"'"
  ]
}'
VOTING_REGISTRY_INIT_MSG_BASE64=$(json_to_base64 "$VOTING_REGISTRY_INIT_MSG")

DAO_INIT='{
  "description": "'"$DAO_DESCRIPTION"'",
  "name": "'"$DAO_NAME"'",
  "proposal_modules_instantiate_info": [
    {
      "admin": {
        "core_module": {}
      },
      "code_id":  '"$PROPOSAL_CONTRACT_BINARY_ID"',
      "label":    "'"$PROPOSAL_SINGLE_LABEL"'",
      "msg":      "'"$PROPOSAL_SINGLE_INIT_MSG_BASE64"'"
    },
    {
      "admin": {
        "core_module": {}
      },
      "code_id":  '"$PROPOSAL_MULTIPLE_CONTRACT_BINARY_ID"',
      "label":    "'"$PROPOSAL_MULTIPLE_LABEL"'",
      "msg":      "'"$PROPOSAL_MULTIPLE_INIT_MSG_BASE64"'"
    },
    {
      "admin": {
        "core_module": {}
      },
      "code_id":  '"$PROPOSAL_CONTRACT_BINARY_ID"',
      "label":    "'"$PROPOSAL_OVERRULE_LABEL"'",
      "msg":      "'"$PROPOSAL_OVERRULE_INIT_MSG_BASE64"'"
    }
  ],
  "voting_registry_module_instantiate_info": {
    "admin": {
      "core_module": {}
    },
    "code_id":  '"$VOTING_REGISTRY_CONTRACT_BINARY_ID"',
    "label":    "'"$VOTING_REGISTRY_LABEL"'",
    "msg":      "'"$VOTING_REGISTRY_INIT_MSG_BASE64"'"
  }
}'

# RESERVE
RESERVE_INIT='{
  "main_dao_address":       "'"$DAO_CONTRACT_ADDRESS"'",
  "security_dao_address":   "'"$SECURITY_SUBDAO_CORE_CONTRACT_ADDRESS"'",
  "denom":                  "'"$STAKEDENOM"'",
  "distribution_rate":      "'"$RESERVE_DISTRIBUTION_RATE"'",
  "min_period":             '"$RESERVE_MIN_PERIOD"',
  "distribution_contract":  "'"$DISTRIBUTION_CONTRACT_ADDRESS"'",
  "treasury_contract":      "'"$DAO_CONTRACT_ADDRESS"'",
  "vesting_denominator":    "'"$RESERVE_VESTING_DENOMINATOR"'"
}'

DISTRIBUTION_INIT='{
  "main_dao_address":     "'"$DAO_CONTRACT_ADDRESS"'",
  "security_dao_address": "'"$SECURITY_SUBDAO_CORE_CONTRACT_ADDRESS"'",
  "denom":                "'"$STAKEDENOM"'"
}'

# VAULTS

NEUTRON_VAULT_INIT='{
  "owner": "'"$DAO_CONTRACT_ADDRESS"'",
  "name":         "'"$NEUTRON_VAULT_NAME"'",
  "denom":        "'"$STAKEDENOM"'",
  "description":  "'"$NEUTRON_VAULT_DESCRIPTION"'"
}'

# CW4 MODULES FOR SUBDAOS

CW4_VOTE_INIT_MSG='{
  "cw4_group_code_id": '"$CW4_GROUP_CONTRACT_BINARY_ID"',
  "initial_members": [
    {
      "addr": "'"$ADMIN_ADDRESS"'",
      "weight": 1
    }
  ]
}'
CW4_VOTE_INIT_MSG_BASE64=$(json_to_base64 "$CW4_VOTE_INIT_MSG")

# SECURITY_SUBDAO

# SECURITY_SUBDAO_PRE_PROPOSE_INIT_MSG will be put into the SECURITY_SUBDAO_PROPOSAL_INIT_MSG
SECURITY_SUBDAO_PRE_PROPOSE_INIT_MSG='{
   "open_proposal_submission": false
}'
SECURITY_SUBDAO_PRE_PROPOSE_INIT_MSG_BASE64=$(json_to_base64 "$SECURITY_SUBDAO_PRE_PROPOSE_INIT_MSG")

SECURITY_SUBDAO_PROPOSAL_INIT_MSG='{
   "allow_revoting": false,
   "pre_propose_info":{
         "module_may_propose":{
            "info":{
               "admin": {
                     "address": {
                       "addr": "'"$DAO_CONTRACT_ADDRESS"'"
                     }
               },
               "code_id": '"$PRE_PROPOSAL_CONTRACT_BINARY_ID"',
               "msg":     "'"$SECURITY_SUBDAO_PRE_PROPOSE_INIT_MSG_BASE64"'",
               "label":   "'"$SECURITY_SUBDAO_PRE_PROPOSE_LABEL"'"
            }
         }
      },
   "only_members_execute":false,
   "max_voting_period":{
      "height": 1000000000000
   },
   "close_proposal_on_execution_failure":false,
   "threshold":{
      "absolute_count":{
         "threshold": "1"
      }
   }
}'
SECURITY_SUBDAO_PROPOSAL_INIT_MSG_BASE64=$(json_to_base64 "$SECURITY_SUBDAO_PROPOSAL_INIT_MSG")

SECURITY_SUBDAO_CORE_INIT_MSG='{
  "name":         "'"$SECURITY_SUBDAO_CORE_NAME"'",
  "description":  "'"$SECURITY_SUBDAO_CORE_DESCRIPTION"'",
  "vote_module_instantiate_info": {
    "admin": {
      "address": {
        "addr": "'"$DAO_CONTRACT_ADDRESS"'"
      }
    },
    "code_id":  '"$CW4_VOTING_CONTRACT_BINARY_ID"',
    "label":    "'"$SECURITY_SUBDAO_VOTE_LABEL"'",
    "msg":      "'"$CW4_VOTE_INIT_MSG_BASE64"'"
  },
  "proposal_modules_instantiate_info": [
    {
      "admin": {
        "address": {
          "addr": "'"$DAO_CONTRACT_ADDRESS"'"
        }
      },
      "code_id":  '"$SUBDAO_PROPOSAL_BINARY_ID"',
      "label":    "'"$SECURITY_SUBDAO_PROPOSAL_LABEL"'",
      "msg":      "'"$SECURITY_SUBDAO_PROPOSAL_INIT_MSG_BASE64"'"
    }
  ],
  "main_dao":     "'"$DAO_CONTRACT_ADDRESS"'",
  "security_dao": "'"$SECURITY_SUBDAO_CORE_CONTRACT_ADDRESS"'"
}'

# GRANTS_SUBDAO

GRANTS_SUBDAO_TIMELOCK_INIT_MSG='{
  "overrule_pre_propose": "'"$PROPOSAL_OVERRULE_CONTRACT_ADDRESS"'"
}'
GRANTS_SUBDAO_TIMELOCK_INIT_MSG_BASE64=$(json_to_base64 "$GRANTS_SUBDAO_TIMELOCK_INIT_MSG")

GRANTS_SUBDAO_PRE_PROPOSE_INIT_MSG='{
  "open_proposal_submission": true,
  "timelock_module_instantiate_info": {
    "admin": {
      "address": {
        "addr": "'"$DAO_CONTRACT_ADDRESS"'"
      }
    },
    "code_id":  '"$SUBDAO_TIMELOCK_BINARY_ID"',
    "label":    "'"$GRANTS_SUBDAO_TIMELOCK_LABEL"'",
    "msg":      "'"$GRANTS_SUBDAO_TIMELOCK_INIT_MSG_BASE64"'"
  }
}'
GRANTS_SUBDAO_PRE_PROPOSE_INIT_MSG_BASE64=$(json_to_base64 "$GRANTS_SUBDAO_PRE_PROPOSE_INIT_MSG")

GRANTS_SUBDAO_PROPOSAL_INIT_MSG='{
   "allow_revoting": false,
   "pre_propose_info":{
      "module_may_propose":{
         "info":{
            "admin": {
              "address": {
                "addr": "'"$DAO_CONTRACT_ADDRESS"'"
              }
            },
            "code_id":  '"$SUBDAO_PRE_PROPOSE_BINARY_ID"',
            "msg":      "'"$GRANTS_SUBDAO_PRE_PROPOSE_INIT_MSG_BASE64"'",
            "label":    "'"$GRANTS_SUBDAO_PRE_PROPOSE_LABEL"'"
         }
      }
   },
   "only_members_execute":false,
   "max_voting_period":{
      "height": 1000000000000
   },
   "close_proposal_on_execution_failure":false,
   "threshold":{
      "absolute_count":{
         "threshold": "1"
      }
   }
}'
GRANTS_SUBDAO_PROPOSAL_INIT_MSG_BASE64=$(json_to_base64 "$GRANTS_SUBDAO_PROPOSAL_INIT_MSG")

GRANTS_SUBDAO_CORE_INIT_MSG='{
  "name":         "'"$GRANTS_SUBDAO_CORE_NAME"'",
  "description":  "'"$GRANTS_SUBDAO_CORE_DESCRIPTION"'",
  "vote_module_instantiate_info": {
    "admin": {
      "address": {
        "addr": "'"$DAO_CONTRACT_ADDRESS"'"
      }
    },
    "code_id":  '"$CW4_VOTING_CONTRACT_BINARY_ID"',
    "label":    "'"$GRANTS_SUBDAO_VOTING_MODULE_LABEL"'",
    "msg":      "'"$CW4_VOTE_INIT_MSG_BASE64"'"
  },
  "proposal_modules_instantiate_info": [
    {
      "admin": {
        "address": {
          "addr": "'"$DAO_CONTRACT_ADDRESS"'"
        }
      },
      "code_id":  '"$SUBDAO_PROPOSAL_BINARY_ID"',
      "label":    "'"$GRANTS_SUBDAO_PROPOSAL_LABEL"'",
      "msg":      "'"$GRANTS_SUBDAO_PROPOSAL_INIT_MSG_BASE64"'"
    }
  ],
  "main_dao":     "'"$DAO_CONTRACT_ADDRESS"'",
  "security_dao": "'"$SECURITY_SUBDAO_CORE_CONTRACT_ADDRESS"'"
}'

echo "Instantiate contracts"

function init_contract() {
  BINARY_ID=$1
  INIT_MSG=$2
  LABEL=$3
  check_json "$INIT_MSG"
  $BINARY add-wasm-message instantiate-contract "$BINARY_ID" "$INIT_MSG" --label "$LABEL" \
    --run-as "$DAO_CONTRACT_ADDRESS" --admin "$DAO_CONTRACT_ADDRESS" --home "$CHAIN_DIR"
}

# WARNING!
# The following code is to add contracts instantiations messages to genesis
# It affects the section of predicting contracts addresses at the beginning of the script
# If you're to do any changes, please do it consistently in both sections
init_contract "$NEUTRON_VAULT_CONTRACT_BINARY_ID"   "$NEUTRON_VAULT_INIT"             "$NEUTRON_VAULT_LABEL"
init_contract "$DAO_CONTRACT_BINARY_ID"             "$DAO_INIT"                       "$DAO_CORE_LABEL"
init_contract "$RESERVE_CONTRACT_BINARY_ID"         "$RESERVE_INIT"                   "$RESERVE_LABEL"
init_contract "$DISTRIBUTION_CONTRACT_BINARY_ID"    "$DISTRIBUTION_INIT"              "$DISTRIBUTION_LABEL"
init_contract "$SUBDAO_CORE_BINARY_ID"              "$SECURITY_SUBDAO_CORE_INIT_MSG"  "$SECURITY_SUBDAO_CORE_LABEL"
init_contract "$SUBDAO_CORE_BINARY_ID"              "$GRANTS_SUBDAO_CORE_INIT_MSG"    "$GRANTS_SUBDAO_CORE_LABEL"

ADD_SUBDAOS_MSG='{
  "update_sub_daos": {
    "to_add": [
      {
        "addr": "'"$SECURITY_SUBDAO_CORE_CONTRACT_ADDRESS"'"
      },
      {
        "addr": "'"$GRANTS_SUBDAO_CORE_CONTRACT_ADDRESS"'"
      }
    ],
    "to_remove": []
  }
}'
check_json "$ADD_SUBDAOS_MSG"

$BINARY add-wasm-message execute "$DAO_CONTRACT_ADDRESS" "$ADD_SUBDAOS_MSG" \
  --run-as "$DAO_CONTRACT_ADDRESS" --home "$CHAIN_DIR"

function set_genesis_param() {
  param_name=$1
  param_value=$2
  sed -i -e "s;\"$param_name\":.*;\"$param_name\": $param_value;g" "$GENESIS_PATH"
}

function set_genesis_param_jq() {
  param_path=$1
  param_value=$2
  jq "${param_path} = ${param_value}" > tmp_genesis_file.json < "$GENESIS_PATH" && mv tmp_genesis_file.json "$GENESIS_PATH"
}

function convert_bech32_base64_esc() {
  $BINARY keys parse $1 --output json | jq .bytes | xxd -r -p | base64 | sed -e 's/\//\\\//g'
}
DAO_CONTRACT_ADDRESS_B64=$(convert_bech32_base64_esc "$DAO_CONTRACT_ADDRESS")
echo $DAO_CONTRACT_ADDRESS_B64

set_genesis_param admins                                 "[\"$DAO_CONTRACT_ADDRESS\"]"                 # admin module
set_genesis_param treasury_address                       "\"$DAO_CONTRACT_ADDRESS\""                   # feeburner
set_genesis_param fee_collector_address                  "\"$DAO_CONTRACT_ADDRESS\""                   # tokenfactory
set_genesis_param security_address                       "\"$SECURITY_SUBDAO_CORE_CONTRACT_ADDRESS\"," # cron
set_genesis_param limit                                  5                                             # cron
#set_genesis_param allow_messages                        "[\"*\"]"                                     # interchainaccounts
set_genesis_param signed_blocks_window                   "\"$SLASHING_SIGNED_BLOCKS_WINDOW\","         # slashing
set_genesis_param min_signed_per_window                  "\"$SLASHING_MIN_SIGNED\","                   # slashing
set_genesis_param slash_fraction_double_sign             "\"$SLASHING_FRACTION_DOUBLE_SIGN\","         # slashing
set_genesis_param slash_fraction_downtime                "\"$SLASHING_FRACTION_DOWNTIME\""             # slashing
set_genesis_param minimum_gas_prices                     "$MIN_GAS_PRICES,"                            # globalfee
set_genesis_param max_total_bypass_min_fee_msg_gas_usage "\"$MAX_TOTAL_BYPASS_MIN_FEE_MSG_GAS_USAGE\"" # globalfee

set_genesis_param_jq ".app_state.globalfee.params.bypass_min_fee_msg_types" "$BYPASS_MIN_FEE_MSG_TYPES" # globalfee

set_genesis_param proposer_fee                "\"0.25\""                                    # builder(POB)
set_genesis_param escrow_account_address      "\"$DAO_CONTRACT_ADDRESS_B64\","              # builder(POB)

if ! jq -e . "$GENESIS_PATH" >/dev/null 2>&1; then
    echo "genesis appears to become incorrect json" >&2
    exit 1
fi

echo "DAO $DAO_CONTRACT_ADDRESS"
