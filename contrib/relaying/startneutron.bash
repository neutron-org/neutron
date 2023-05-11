export NEUTROND_P2P_MAX_NUM_OUTBOUND_PEERS=500
export NEUTROND_STATESYNC_RPC_SERVERS="https://rpc-kralum.neutron-1.neutron.org:443,https://rpc-kralum.neutron-1.neutron.org:443"
export NEUTROND_STATESYNC_TRUST_HEIGHT=$BLOCK_HEIGHT
export NEUTROND_STATESYNC_TRUST_HASH=$TRUST_HASH
export NEUTROND_P2P_LADDR=tcp://0.0.0.0:8010
export NEUTROND_RPC_LADDR=tcp://127.0.0.1:8011
export NEUTROND_GRPC_ADDRESS=127.0.0.1:8012
export NEUTROND_GRPC_WEB_ADDRESS=127.0.0.1:8014
export NEUTROND_API_ADDRESS=tcp://127.0.0.1:8013
export NEUTROND_NODE=tcp://127.0.0.1:8011
export NEUTROND_P2P_MAX_NUM_INBOUND_PEERS=500
export NEUTROND_RPC_PPROF_LADDR=127.0.0.1:6969

# Fetch and set list of seeds from chain registry.
# NEUTROND_P2P_SEEDS=$(curl -s https://raw.githubusercontent.com/cosmos/chain-registry/master/neutron/chain.json | jq -r '[foreach .peers.seeds[] as $item (""; "\($item.id)@\($item.address)")] | join(",")')
export NEUTROND_P2P_SEEDS
#
# # Start chain.
neutrond start --x-crisis-skip-assert-invariants --iavl-disable-fastnode false