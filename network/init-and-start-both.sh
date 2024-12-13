#!/bin/bash
set -e

script_full_path=$(dirname "$0")

export BINARY=neutrond
export CHAINID=test-1
export P2PPORT=26656
export RPCPORT=26657
export RESTPORT=1317
export ROSETTA=8080
export GRPCPORT=8090
export GRPCWEB=8091
export STAKEDENOM=untrn

"$script_full_path"/init.sh
"$script_full_path"/init-neutrond.sh
"$script_full_path"/start.sh

exit
export BINARY=gaiad
export CHAINID=test-2
export P2PPORT=16656
export RPCPORT=16657
export RESTPORT=1316
export ROSETTA=9080
export GRPCPORT=9090
export GRPCWEB=9091
export STAKEDENOM=uatom

"$script_full_path"/init.sh
"$script_full_path"/init-gaiad.sh
"$script_full_path"/start.sh
