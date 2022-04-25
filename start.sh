interchain-adapterd init my-node --chain-id testnet

# Create a key to hold your validator account
interchain-adapterd keys add test

# Add that key into the genesis.app_state.accounts array in the genesis file
# NOTE: this command lets you set the number of coins. Make sure this account has some coins
# with the genesis.app_state.staking.params.bond_denom denom, the default is staking
interchain-adapterd add-genesis-account $(interchain-adapterd keys show test -a) 1000000000000000stake,1000000000validatortoken

# Generate the transaction that creates your validator
interchain-adapterd gentx test 1000000000stake --chain-id testnet

# Add the generated bonding transaction to the genesis file
interchain-adapterd collect-gentxs

# Now its safe to start `gaiad`
#interchain-adapterd start --mode validator