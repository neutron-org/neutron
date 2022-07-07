#/bin/bash


ARTIFACTS_DIR=../../lido-interchain-staking-contracts/artifacts/

# VAL=cosmosvaloper18hl5c9xn5dze2g50uaw0l2mr02ew57zk0auktn

BIN=neutrond
ADMIN=neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2
CHAINID=test-1
HOME=../data/test-1/
HOME2=../data/test-2/
KEY=demowallet1

HUB_ADDRESS=neutron1wug8sewp6cedgkmrmvhl3lf3tulagm9hnvy8p0rppz9yjw0g4wtqvfcxh2

RES=$(${BIN} tx wasm store ${ARTIFACTS_DIR}lido_interchain_hub.wasm --from ${KEY} --gas 50000000  --chain-id ${CHAINID} --broadcast-mode=block --gas-prices 0.0025stake  -y --output json  --keyring-backend test --home ${HOME} --node tcp://127.0.0.1:16657)
HUB_CODE_ID=$(echo $RES | jq -r '.logs[0].events[1].attributes[0].value')
echo $RES
echo $HUB_CODE_ID

RES=$(${BIN} tx wasm store ${ARTIFACTS_DIR}lido_interchain_hub.wasm --from ${KEY} --gas 50000000  --chain-id ${CHAINID} --broadcast-mode=block --gas-prices 0.0025stake  -y --output json  --keyring-backend test --home ${HOME} --node tcp://127.0.0.1:16657)
HUB_CODE_ID=$(echo $RES | jq -r '.logs[0].events[1].attributes[0].value')
echo $RES
echo $HUB_CODE_ID

RES=$(${BIN} tx wasm store ${ARTIFACTS_DIR}lido_interchain_queries.wasm --from ${KEY} --gas 50000000  --chain-id ${CHAINID} --broadcast-mode=block --gas-prices 0.0025stake  -y --output json  --keyring-backend test --home ${HOME} --node tcp://127.0.0.1:16657)
QUERIES_CODE_ID=$(echo $RES | jq -r '.logs[0].events[1].attributes[0].value')
echo $RES
echo $QUERIES_CODE_ID

RES=$(${BIN} tx wasm store ${ARTIFACTS_DIR}lido_interchain_validators_registry.wasm --from ${KEY} --gas 50000000  --chain-id ${CHAINID} --broadcast-mode=block --gas-prices 0.0025stake  -y --output json  --keyring-backend test --home ${HOME} --node tcp://127.0.0.1:16657)
VALIDATORS_CODE_ID=$(echo $RES | jq -r '.logs[0].events[1].attributes[0].value')
echo $RES
echo $VALIDATORS_CODE_ID

INIT_HUB='{}'

RES=$(${BIN} tx wasm instantiate $HUB_CODE_ID "$INIT_HUB" --from ${KEY} --admin ${ADMIN} -y --chain-id ${CHAINID} --output json --broadcast-mode=block --label "init"  --keyring-backend test --gas-prices 0.0025stake --home ${HOME} --node tcp://127.0.0.1:16657)
echo $RES
HUB_ADDRESS=$(echo $RES | jq -r '.logs[0].events[0].attributes[0].value')
echo $HUB_ADDRESS

RES=$(${BIN} tx wasm instantiate $QUERIES_CODE_ID "{}" --from ${KEY} --admin ${ADMIN} -y --chain-id ${CHAINID} --output json --broadcast-mode=block --label "init"  --keyring-backend test --gas-prices 0.0025stake --home ${HOME} --node tcp://127.0.0.1:16657)
echo $RES
INTERCHAIN_ADDRESS=$(echo $RES | jq -r '.logs[0].events[0].attributes[0].value')
echo $INTERCHAIN_ADDRESS

UPDATE_CONFIG="{\"update_config\":{\"pool_distribution\":10,\"permissionless_limit\":10,\"validators_registry_code_id\":${VALIDATORS_CODE_ID},\"interchain_queries_contract\":\"${INTERCHAIN_ADDRESS}\"}}"

echo UPDATE_CONFIG
RES=$(${BIN} tx wasm execute $HUB_ADDRESS "$UPDATE_CONFIG" --from ${KEY}  -y --chain-id ${CHAINID} --output json --broadcast-mode=block --gas-prices 0.0025stake --gas 1000000 --keyring-backend test --home ${HOME} --node tcp://127.0.0.1:16657)
echo $RES


REGISTER_INTERCHAIN='{"register_interchain_account":{"zone_id":"test-2", "connection_id": "connection-0","transfer_channel": "transfer", "staked_denomination": "stake", "staked_denomination_decimals": 6, "stasset_denomination": "stasset"}}'

echo $REGISTER_INTERCHAIN
RES=$(${BIN} tx wasm execute $HUB_ADDRESS "$REGISTER_INTERCHAIN" --from ${KEY}  -y --chain-id ${CHAINID} --output json --broadcast-mode=block --gas-prices 0.0025stake --gas 1000000 --keyring-backend test --home ${HOME} --node tcp://127.0.0.1:16657)
echo $RES

# exit 0
sleep 15

INSTANTIATE_ZONE='{"instantiate_zone":{"zone_id":"test-2"}}'
echo $INSTANTIATE_ZONE
RES=$(${BIN} tx wasm execute $HUB_ADDRESS "$INSTANTIATE_ZONE" --from  ${KEY}  -y --chain-id ${CHAINID} --output json --broadcast-mode=block --gas-prices 0.0025stake --gas 1000000 --keyring-backend test --home ${HOME} --node tcp://127.0.0.1:16657)
echo $RES

INTERCHAIN_ACCOUNT=$(${BIN} q ibc channel channels --home ./data/test-1 --node tcp://localhost:16657 --output json | jq -r '.channels[0].version' | jq -r '.address')
${BIN} tx bank send demowallet2 $INTERCHAIN_ACCOUNT 10000stake --chain-id test-2 --home ${HOME2} --node tcp://localhost:26657 --keyring-backend test -y --gas-prices 0.0025stake --broadcast-mode=block
${BIN} tx bank send demowallet2 $INTERCHAIN_ACCOUNT 10000stake --chain-id test-2 --home ${HOME2} --node tcp://localhost:26657 --keyring-backend test -y --gas-prices 0.0025stake --broadcast-mode=block
echo "interchain account $INTERCHAIN_ACCOUNT"
echo "interchain query contract $INTERCHAIN_ADDRESS"
sleep 15
${BIN} q wasm contract-state smart  ${INTERCHAIN_ADDRESS} "{\"balance\":{\"zone_id\":\"test-2\",\"addr\":\"${INTERCHAIN_ACCOUNT}\",\"denom\":\"stake\"}}" --node tcp://localhost:16657
${BIN} q wasm contract-state smart ${INTERCHAIN_ADDRESS} "{\"get_transfers\":{\"zone_id\":\"test-2\",\"recipient\":\"${INTERCHAIN_ACCOUNT}\",\"start\":0,\"end\":10}}" --node tcp://localhost:16657