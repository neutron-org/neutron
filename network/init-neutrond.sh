#!/bin/bash
set -e

BINARY=${BINARY:-neutrond}
BASE_DIR=./data
CHAINID=${CHAINID:-test-1}
STAKEDENOM=${STAKEDENOM:-untrn}
CONTRACTS_BINARIES_DIR=${CONTRACTS_BINARIES_DIR:-./contracts}
THIRD_PARTY_CONTRACTS_DIR=${THIRD_PARTY_CONTRACTS_DIR:-./contracts_thirdparty}
FEEMARKET_ENABLED=${FEEMARKET_ENABLED:-true}

# IMPORTANT! minimum_gas_prices should always contain at least one record, otherwise the chain will not start or halt
# ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2 denom is required by integration tests (test:tokenomics)
MIN_GAS_PRICES_DEFAULT='[{"denom":"ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2","amount":"0"},{"denom":"untrn","amount":"0"}]'
MIN_GAS_PRICES=${MIN_GAS_PRICES:-"$MIN_GAS_PRICES_DEFAULT"}

ADMIN_MODULE_ADDRESS="neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z"
GOV_MODULE_ADDRESS="neutron10d07y265gmmuvt4z0w9aw880jnsr700j7a68v5"

BYPASS_MIN_FEE_MSG_TYPES_DEFAULT='["/ibc.core.channel.v1.Msg/RecvPacket", "/ibc.core.channel.v1.Msg/Acknowledgement", "/ibc.core.client.v1.Msg/UpdateClient"]'
BYPASS_MIN_FEE_MSG_TYPES=${BYPASS_MIN_FEE_MSG_TYPES:-"$BYPASS_MIN_FEE_MSG_TYPES_DEFAULT"}

MAX_TOTAL_BYPASS_MIN_FEE_MSG_GAS_USAGE_DEFAULT=1000000
MAX_TOTAL_BYPASS_MIN_FEE_MSG_GAS_USAGE=${MAX_TOTAL_BYPASS_MIN_FEE_MSG_GAS_USAGE:-"$MAX_TOTAL_BYPASS_MIN_FEE_MSG_GAS_USAGE_DEFAULT"}

CHAIN_DIR="$BASE_DIR/$CHAINID"
GENESIS_PATH="$CHAIN_DIR/config/genesis.json"

ADMIN_ADDRESS=$($BINARY keys show demowallet1 -a --home "$CHAIN_DIR" --keyring-backend test)
SECOND_MULTISIG_ADDRESS=$($BINARY keys show demowallet2 -a --home "$CHAIN_DIR" --keyring-backend test)


# Slinky genesis configs
USE_CORE_MARKETS=${USE_CORE_MARKETS:-true}
USE_RAYDIUM_MARKETS=${USE_RAYDIUM_MARKETS:-false}
USE_UNISWAPV3_BASE_MARKETS=${USE_UNISWAPV3_BASE_MARKETS:-false}
USE_COINGECKO_MARKETS=${USE_COINGECKO_MARKETS:-false}

echo "Creating and collecting gentx..."
$BINARY  gentx val1 "1000000$STAKEDENOM" --home "$CHAIN_DIR" --chain-id "$CHAINID" --keyring-backend test
$BINARY  collect-gentxs --home "$CHAIN_DIR"
### PARAMETERS SECTION

## consensus params
CONSENSUS_BLOCK_MAX_GAS=1000000000
CONSENSUS_VOTE_EXTENSIONS_ENABLE_HEIGHT=1

## slashing params
SLASHING_SIGNED_BLOCKS_WINDOW=140000
SLASHING_MIN_SIGNED=0.050000000000000000
SLASHING_FRACTION_DOUBLE_SIGN=0.010000000000000000
SLASHING_FRACTION_DOWNTIME=0.000100000000000000


function store_binary() {
  CONTRACT_BINARY_PATH=$1
  $BINARY add-wasm-message store "$CONTRACT_BINARY_PATH" \
    --output json --run-as "${ADMIN_ADDRESS}" --keyring-backend=test --home "$CHAIN_DIR"
  BINARY_ID=$(jq -r "[.app_state.wasm.gen_msgs[] | select(.store_code != null)] | length" "$CHAIN_DIR/config/genesis.json")
  echo "$BINARY_ID"
}

function genaddr() {
  CODE_ID=$1
  CONTRACT_ADDRESS=$($BINARY debug generate-contract-address "$INSTANCE_ID_COUNTER" "$CODE_ID")
  echo "$CONTRACT_ADDRESS"
}

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
GOV_MODULE_ADDRESS_B64=$(convert_bech32_base64_esc "$GOV_MODULE_ADDRESS")
echo $GOV_MODULE_ADDRESS_B64

echo "Adding marketmap into genesis..."
go run network/slinky_genesis.go --use-core=$USE_CORE_MARKETS --use-raydium=$USE_RAYDIUM_MARKETS --use-uniswapv3-base=$USE_UNISWAPV3_BASE_MARKETS --use-coingecko=$USE_COINGECKO_MARKETS --temp-file=markets.json
MARKETS=$(cat markets.json)

NUM_MARKETS=$(echo "$MARKETS" | jq '.markets | length + 1')

NUM_MARKETS=$NUM_MARKETS; jq --arg num "$NUM_MARKETS" '.app_state["oracle"]["next_id"] = $num' "$GENESIS_PATH" > genesis_tmp.json && mv genesis_tmp.json "$GENESIS_PATH"
MARKETS=$MARKETS; jq --arg markets "$MARKETS" '.app_state["marketmap"]["market_map"] = ($markets | fromjson)' "$GENESIS_PATH" > genesis_tmp.json && mv genesis_tmp.json "$GENESIS_PATH"
MARKETS=$MARKETS; jq --arg markets "$MARKETS" '.app_state["oracle"]["currency_pair_genesis"] += [$markets | fromjson | .markets | values | .[].ticker.currency_pair | {"currency_pair": {"Base": .Base, "Quote": .Quote}, "currency_pair_price": null, "nonce": 0} ]' "$GENESIS_PATH" > genesis_tmp.json && mv genesis_tmp.json "$GENESIS_PATH"
MARKETS=$MARKETS; jq --arg markets "$MARKETS" '.app_state["oracle"]["currency_pair_genesis"] |= (to_entries | map(.value += {id: (.key + 1)} | .value))' "$GENESIS_PATH" > genesis_tmp.json && mv genesis_tmp.json "$GENESIS_PATH"

rm markets.json

BANK_DENOMS_METADATA='
[{
    "description": "The native staking token of the Neutron network",
    "denom_units": [
      {
        "denom": "untrn",
        "exponent": 0,
        "aliases": [
          "microntrn"
        ]
      },
      {
        "denom": "ntrn",
        "exponent": 6,
        "aliases": [
          "NTRN"
        ]
      }
    ],
    "base": "untrn",
    "display": "ntrn",
    "name": "Neutron",
    "symbol": "NTRN",
    "uri": "",
    "uri_hash": ""
  }]
'

echo "Setting the rest of Neutron genesis params..."
set_genesis_param fee_collector_address                  "\"$GOV_MODULE_ADDRESS\","                      # tokenfactory
set_genesis_param_jq ".app_state.cron.params.security_address" "\"$GOV_MODULE_ADDRESS\"" # cron
set_genesis_param limit                                  5                                                # cron
set_genesis_param signed_blocks_window                   "\"$SLASHING_SIGNED_BLOCKS_WINDOW\","            # slashing
set_genesis_param min_signed_per_window                  "\"$SLASHING_MIN_SIGNED\","                      # slashing
set_genesis_param slash_fraction_double_sign             "\"$SLASHING_FRACTION_DOUBLE_SIGN\","            # slashing
set_genesis_param slash_fraction_downtime                "\"$SLASHING_FRACTION_DOWNTIME\""                # slashing
set_genesis_param minimum_gas_prices                     "$MIN_GAS_PRICES,"                               # globalfee
set_genesis_param max_total_bypass_min_fee_msg_gas_usage "\"$MAX_TOTAL_BYPASS_MIN_FEE_MSG_GAS_USAGE\""    # globalfee
set_genesis_param_jq ".app_state.globalfee.params.bypass_min_fee_msg_types" "$BYPASS_MIN_FEE_MSG_TYPES"   # globalfee
set_genesis_param proposer_fee                          "\"0.25\""                                        # builder(POB)
set_genesis_param escrow_account_address                "\"$GOV_MODULE_ADDRESS_B64\","                  # builder(POB)
set_genesis_param sudo_call_gas_limit                   "\"1000000\""                                     # contractmanager
set_genesis_param max_gas                               "\"$CONSENSUS_BLOCK_MAX_GAS\""                    # consensus_params
set_genesis_param vote_extensions_enable_height         "\"$CONSENSUS_VOTE_EXTENSIONS_ENABLE_HEIGHT\""    # consensus_params
set_genesis_param_jq ".app_state.marketmap.params.admin" "\"$GOV_MODULE_ADDRESS\""                      # marketmap
set_genesis_param_jq ".app_state.marketmap.params.market_authorities" "[\"$GOV_MODULE_ADDRESS\"]"       # marketmap
set_genesis_param_jq ".app_state.feemarket.params.max_block_utilization" "\"$CONSENSUS_BLOCK_MAX_GAS\""   # feemarket
set_genesis_param_jq ".app_state.feemarket.params.min_base_gas_price"    "\"0.0025\""                     # feemarket
set_genesis_param_jq ".app_state.feemarket.params.fee_denom"       "\"untrn\""                            # feemarket
set_genesis_param_jq ".app_state.feemarket.params.max_learning_rate" "\"0.5\""                            # feemarket
set_genesis_param_jq ".app_state.feemarket.params.enabled" "$FEEMARKET_ENABLED"                           # feemarket
set_genesis_param_jq ".app_state.feemarket.params.distribute_fees" "true"                                 # feemarket
set_genesis_param_jq ".app_state.feemarket.state.base_gas_price" "\"0.0025\""                             # feemarket
set_genesis_param_jq ".app_state.bank.denom_metadata" "$BANK_DENOMS_METADATA"                                    # bank metadata

if ! jq -e . "$GENESIS_PATH" >/dev/null 2>&1; then
    echo "genesis appears to become incorrect json" >&2
    exit 1
fi