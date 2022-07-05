#!/bin/bash

cd $(dirname "$0")

export DEMOWALLET_1=$(neutrond keys show demowallet1 -a --keyring-backend test --home ../data/test-1) && echo $DEMOWALLET_1;
export DEMOWALLET_2=$(neutrond keys show demowallet2 -a --keyring-backend test --home ../data/test-2) && echo $DEMOWALLET_2;
export OWNER=$(neutrond keys show rly1 -a --keyring-backend test --home ../data/test-1) && echo $OWNER;

../build/neutrond tx interchaintxs register-interchain-account connection-0 cosmos14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr \
    --from $DEMOWALLET_1 --chain-id test-1 --home ../data/test-1 --node tcp://localhost:16657 --keyring-backend test -y

sleep 5;

../build/neutrond tx wasm execute cosmos14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr '{"register_interchain_account": {"zone_id":"xxx", "connection_id":"connection-0", "staked_denomination": "stake", "staked_denomination_decimals": 6}}'  --chain-id test-1 --from cosmos1mjk79fjjgpplak5wq838w0yd982gzkyfrk07am --gas 20000000   --gas-adjustment 1.4 --gas-prices 0.5stake --keyring-backend test   --home ../data/test-1 --node tcp://127.0.0.1:16657   -y