./build/gaia-wasm-zoned init my-node --chain-id testnet

# Create a key to hold your validator account
./build/gaia-wasm-zoned keys add test

# Add that key into the genesis.app_state.accounts array in the genesis file
# NOTE: this command lets you set the number of coins. Make sure this account has some coins
# with the genesis.app_state.staking.params.bond_denom denom, the default is staking
./build/gaia-wasm-zoned add-genesis-account $(gaia-wasm-zoned keys show test -a) 1000000000000000stake,1000000000validatortoken

# Generate the transaction that creates your validator
./build/gaia-wasm-zoned gentx test 1000000000stake --chain-id testnet

# Add the generated bonding transaction to the genesis file
./build/gaia-wasm-zoned collect-gentxs

# Now its safe to start `gaiad`
./build/gaia-wasm-zoned start --minimum-gas-prices 0.25stake