#!/bin/bash


BINARY=${BINARY:-gaiad}
BASE_DIR=./data
CHAINID=${CHAINID:-test-2}
CHAIN_DIR="$BASE_DIR/$CHAINID"

STAKEDENOM=${STAKEDENOM:-stake}

echo "Waiting for a first block..."
while ! curl -f http://127.0.0.1:1316/blocks/1 >/dev/null 2>&1; do
  sleep 1
done


${BINARY} tx staking delegate $(${BINARY} q staking validators --node http://127.0.0.1:16657 --output json | jq .validators[0].operator_address -r) 10000000000${STAKEDENOM} --from demowallet1 --chain-id test-2 --home ./data/test-2 --keyring-backend test --node http://127.0.0.1:16657 --gas 500000 --gas-prices 0.5uatom -y  --broadcast-mode block
SPAWN_TIME=$(date -Iseconds -u --date='50 seconds')
jq  '.spawn_time="'${SPAWN_TIME}'"' ./network/proposal_template.json > ./network/proposal.json
${BINARY} tx gov submit-proposal consumer-addition ./network/proposal.json --from demowallet1 --chain-id test-2 --broadcast-mode block --gas 500000  --home ./data/test-2 --keyring-backend test --node http://127.0.0.1:16657 --gas 500000 --gas-prices 0.5uatom -y
echo "12345678" | ${BINARY} tx gov vote 1 yes --chain-id test-2 --broadcast-mode block -y --from demowallet1  --home ./data/test-2 --keyring-backend test --node http://127.0.0.1:16657 --gas 500000 --gas-prices 0.5uatom -y
W=50
echo "wating to the end of proposal for ${W}sec"
sleep ${W}
${BINARY} query provider consumer-genesis test-1 -o json --node http://127.0.0.1:16657 | jq '.params |= . + {"soft_opt_out_threshold": "0.05", "provider_reward_denoms": ["uatom"], "reward_denoms": ["untrn"], "distribution_transmission_channel": "channel-1"}' > ccv-state.json