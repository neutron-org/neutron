# Gaia Wasm Zone

## Requirments
* Go 1.18
* Ignite Cli
* Hermes IBC Relayer

### How to install Ignite CLI

```shell
curl https://get.ignite.com/cli! | bash
```

### How to install Hermes IBC Relayer

```shell
cargo install --version 0.14.1 ibc-relayer-cli --bin hermes --locked
```

## Build and Install Gaia Wasm Zone

```shell
make install
```

## Run local testnet node instances connected via IBC

### Bootstrap two chains and create an IBC connection

```shell
make init
```

### Start relayer

```shell
make start-rly
```

## Generate proto

```shell
ignite generate proto-go
```


# Testing with 2 lido-chains (easier for development)

### terminal 1

1. Start 2 lido-chains and an IBC relayer:
```
make init && make start-rly
```

### terminal 2
1. Register an interchain query to get delegations of delegator on remote chain:
```
gaia-wasm-zoned tx interchainqueries register-interchain-query test-2 connection-0 x/staking/DelegatorDelegations '{"delegator": "cosmos1qnk2n4nlkpw9xfqntladh74w6ujtulwn7j8za9"}' 1 --from demowallet1 --gas 10000000 --gas-adjustment 1.4 --gas-prices 0.5stake --broadcast-mode block --chain-id test-1 --keyring-backend test --home ./data/test-1 --node tcp://127.0.0.1:16657
```

2. Register an interchain query to search transactions on remote chain by some event (in this case it'll try to find all transactions for bank transfer):
```
gaia-wasm-zoned tx interchainqueries register-interchain-query test-2 connection-0 x/tx/RecipientTransactions '{"message.module": "bank"}' 5 --from demowallet1 --gas 10000000 --gas-adjustment 1.4 --gas-prices 0.5stake --broadcast-mode block --chain-id test-1 --keyring-backend test --home ./data/test-1 --node tcp://127.0.0.1:16657
```

3. Execute a bank transfer transaction on remote chain for the interchain query above:
```
gaia-wasm-zoned tx bank send $(gaia-wasm-zoned keys show demowallet2 -a --keyring-backend test --home ./data/test-2) cosmos1mjk79fjjgpplak5wq838w0yd982gzkyfrk07am 1000stake --from demowallet2 --gas 10000000 --gas-adjustment 1.4 --gas-prices 0.5stake --broadcast-mode block --chain-id test-2 --keyring-backend test --home ./data/test-2 --node tcp://127.0.0.1:26657
```

4. After relayer process query events, you can see submitted results:
* Updated info about registered queries:
```shell
gaia-wasm-zoned query interchainqueries registered-queries --node tcp://127.0.0.1:16657
```

* Result for KV storage query (in our case DelegatorDelegations and this query id is 1):
```shell
gaia-wasm-zoned query interchainqueries query-result 1 --node tcp://127.0.0.1:16657
```

* Result for transactions search query (in our case MsgBankSend and this query id is 2, and we set limit and offset to 1 and 100 respectively):
```shell
gaia-wasm-zoned query interchainqueries query-transactions-search-result 2 1 100 --node tcp://127.0.0.1:16657
```


### terminal 3

1. `git clone git@github.com:lidofinance/cosmos-query-relayer.git`
2. `cd cosmos-query-relayer`
3. `cp configs/dev.example.2-lido-chains.yml configs/dev.yml`
4. `make dev`
