#!/bin/bash

cd $(dirname "$0")

OWNER=$(gaia-wasm-zoned keys show rly1 -a --keyring-backend test --home ../data/test-1);

echo "Owner" $OWNER
echo "Depoy Hub:" 
TX_HASH=`../build/gaia-wasm-zoned tx wasm store ./../../lido-interchain-staking-contracts/artifacts/lido_interchain_hub.wasm --chain-id test-1 --from demowallet1 --gas 20000000 --gas-adjustment 1.4 --gas-prices 0.5stake --keyring-backend test --home ../data/test-1 --node tcp://127.0.0.1:16657 -y | grep txhash | cut -d " " -f 2`
echo "Tx hash:" $TX_HASH
sleep 5;
CODE_ID=`../build/gaia-wasm-zoned query tx $TX_HASH --chain-id test-1 --home ../data/test-1 --node tcp://127.0.0.1:16657 --output json | jq -r '.logs | .[] | .events | .[] | select(.type=="store_code") | .attributes | .[] | .value'`
echo "Code id:" $CODE_ID


TX_HASH=`../build/gaia-wasm-zoned tx wasm instantiate $CODE_ID '{"zone_id":"xxx", "owner":"'$OWNER'"}' \
     --admin $OWNER --label=hub --chain-id test-1 --from $OWNER --gas 20000000 \
     --gas-adjustment 1.4 --gas-prices 0.5stake --keyring-backend test \
     --home ../data/test-1 --node tcp://127.0.0.1:16657 \
     -y | grep txhash | cut -d " " -f 2`

echo "instantiate Tx hash: " $TX_HASH
sleep 3;
HUB_CONTRACT_ADDRESS=`../build/gaia-wasm-zoned query tx $TX_HASH --chain-id test-1 --home ../data/test-1 --node tcp://127.0.0.1:16657 --output json | jq -r '.logs | .[] | .events | .[] |select(.type=="instantiate") | .attributes | .[] | select(.key=="_contract_address") | .value'`
echo "Hub Contract address: " $HUB_CONTRACT_ADDRESS

echo ""
echo "Depoy validator registry" 
TX_HASH=`../build/gaia-wasm-zoned tx wasm store ./../../lido-interchain-staking-contracts/artifacts/lido_interchain_validators_registry.wasm --chain-id test-1 --from demowallet1 --gas 20000000 --gas-adjustment 1.4 --gas-prices 0.5stake --keyring-backend test --home ../data/test-1 --node tcp://127.0.0.1:16657 -y | grep txhash | cut -d " " -f 2`
echo "Tx hash:" $TX_HASH
sleep 2;
VALIDATOR_CODE_ID=`../build/gaia-wasm-zoned query tx $TX_HASH --chain-id test-1 --home ../data/test-1 --node tcp://127.0.0.1:16657 --output json | jq -r '.logs | .[] | .events | .[] | select(.type=="store_code") | .attributes | .[] | .value'`
echo "Validator registry code id:" $VALIDATOR_CODE_ID

echo ""
echo "Depoy interchain queries" 
TX_HASH=`../build/gaia-wasm-zoned tx wasm store ./../../lido-interchain-staking-contracts/artifacts/lido_interchain_queries.wasm --chain-id test-1 --from demowallet1 --gas 20000000 --gas-adjustment 1.4 --gas-prices 0.5stake --keyring-backend test --home ../data/test-1 --node tcp://127.0.0.1:16657 -y | grep txhash | cut -d " " -f 2`
echo "Tx hash:" $TX_HASH
sleep 2;
CODE_ID=`../build/gaia-wasm-zoned query tx $TX_HASH --chain-id test-1 --home ../data/test-1 --node tcp://127.0.0.1:16657 --output json | jq -r '.logs | .[] | .events | .[] | select(.type=="store_code") | .attributes | .[] | .value'`
echo "Code id:" $VALIDATOR_CODE_ID

TX_HASH=`../build/gaia-wasm-zoned tx wasm instantiate $CODE_ID '{}' \
     --admin $OWNER --label=interchain --chain-id test-1 --from $OWNER --gas 20000000 \
     --gas-adjustment 1.4 --gas-prices 0.5stake --keyring-backend test \
     --home ../data/test-1 --node tcp://127.0.0.1:16657 \
     -y | grep txhash | cut -d " " -f 2`

echo "instantiate Tx hash: " $TX_HASH
sleep 3;
IQ_CONTRACT_ADDRESS=`../build/gaia-wasm-zoned query tx $TX_HASH --chain-id test-1 --home ../data/test-1 --node tcp://127.0.0.1:16657 --output json | jq -r '.logs | .[] | .events | .[] |select(.type=="instantiate") | .attributes | .[] | select(.key=="_contract_address") | .value'`
echo "Interchain queries Contract address: " $IQ_CONTRACT_ADDRESS


echo "update hub config";
gaia-wasm-zoned tx wasm execute $HUB_CONTRACT_ADDRESS '{"update_config": {"interchain_queries_contract":"'$IQ_CONTRACT_ADDRESS'", "permissionless_limit": 70, "pool_distribution": 70, "validators_registry_code_id": '$VALIDATOR_CODE_ID' }}'\
  --chain-id test-1 --from $OWNER --gas 20000000 \
  --gas-adjustment 1.4 --gas-prices 0.5stake --keyring-backend test \
  --home ../data/test-1 --node tcp://127.0.0.1:16657 \
  -y
sleep 3;
