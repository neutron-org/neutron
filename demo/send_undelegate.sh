#!/bin/bash

cd $(dirname "$0")

export OWNER=$(gaia-wasm-zoned keys show rly1 -a --keyring-backend test --home ../data/test-1) && echo $OWNER;

OWNER="cosmos14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr"

../build/gaia-wasm-zoned tx interchaintxs submit-tx connection-0 $OWNER \
    ./test_tx_undelegate.json send_undelegate_memo_12345 operation_undelegate \
    --from demowallet1 --gas 10000000 --gas-adjustment 1.4 --gas-prices 0.5stake \
    --broadcast-mode block --chain-id test-1 --keyring-backend test --home ../data/test-1 --node tcp://127.0.0.1:16657 -y
