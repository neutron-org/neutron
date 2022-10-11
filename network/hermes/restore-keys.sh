#!/bin/bash
set -e

# Load shell variables
. ./network/hermes/variables.sh

### Sleep is needed otherwise the relayer crashes when trying to init
sleep 1

### Restore Keys
echo "alley afraid soup fall idea toss can goose become valve initial strong forward bright dish figure check leopard decide warfare hub unusual join cart" > ./test_1_m.txt
$HERMES_BINARY --config ./network/hermes/config.toml keys delete --chain test-1 --key-name testkey_1
$HERMES_BINARY --config ./network/hermes/config.toml keys add --key-name testkey_1 --chain test-1 --mnemonic-file ./test_1_m.txt
rm ./test_1_m.txt
sleep 5

echo "record gift you once hip style during joke field prize dust unique length more pencil transfer quit train device arrive energy sort steak upset" > ./test_2_m.txt
$HERMES_BINARY --config ./network/hermes/config.toml keys delete --chain test-2 --key-name testkey_2
$HERMES_BINARY --config ./network/hermes/config.toml keys add --key-name testkey_2 --chain test-2 --mnemonic-file ./test_2_m.txt
rm ./test_2_m.txt
sleep 5
