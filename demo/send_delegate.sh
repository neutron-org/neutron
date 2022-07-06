#!/bin/bash

cd $(dirname "$0")

export OWNER=$(gaia-wasm-zoned keys show rly1 -a --keyring-backend test --home ../data/test-1) && echo $OWNER;

../build/gaia-wasm-zoned tx bank send cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs cosmos1lsgkk3q0xtz2wslumwvfwe487k2ksgsz202tsqqnxcsjt72uc3pq53fh9s 10000stake --chain-id test-2 --home ../data/test-2 --node tcp://localhost:26657 --keyring-backend test -y

sleep 3;
OWNER="cosmos14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr"

../build/gaia-wasm-zoned tx interchaintxs submit-tx connection-0 $OWNER \
    ./test_tx_delegate.json --from demowallet1 --gas 10000000 --gas-adjustment 1.4 --gas-prices 0.5stake \
    --broadcast-mode block --chain-id test-1 --keyring-backend test --home ../data/test-1 --node tcp://127.0.0.1:16657