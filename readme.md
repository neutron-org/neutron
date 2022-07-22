# Neutron Zone

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

## Build and Install Neutron Zone

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


# Testing with 2 neutron-chains (easier for development)

### terminal 1

1. Start 2 neutron-chains and an IBC relayer:
```
make init && make start-rly
```

### terminal 2
1. Register an interchain query to get a delegation from delegator `neutron1qnk2n4nlkpw9xfqntladh74w6ujtulwn6dwq8z` to validator `neutronvaloper1qnk2n4nlkpw9xfqntladh74w6ujtulwnqshepx` on remote chain.
```
neutrond tx interchainqueries register-interchain-query test-2 connection-0 5 kv staking/311404eca9d67fb05c5324135ffadbfaaed724be7dd31404eca9d67fb05c5324135ffadbfaaed724be7dd3 --from demowallet1 --gas 10000000 --gas-adjustment 1.4 --gas-prices 0.5stake --broadcast-mode block --chain-id test-1 --keyring-backend test --home ./data/test-1 --node tcp://127.0.0.1:16657
```
<details>
  <summary>What is staking/311404eca9d67fb05c5324135ffadbfaaed724be7dd31404eca9d67fb05c5324135ffadbfaaed724be7dd3?</summary>

`--kv-keys` is a flag that allows to register an interchain query that wants to read raw data from KV-storage on remote chain by some key.

At first we should compose the correct key for the query. Any delegation stores in KV-storage under this key ([cosmos-sdk code](https://github.com/cosmos/cosmos-sdk/blob/ad9e5620fb3445c716e9de45cfcdb56e8f1745bf/x/staking/types/keys.go#L176)):
```
0x31 + lengthPrefixed(delegator_address_bytes) + lengthPrefixed(validator_address_bytes)
```
We know delegator address and validator address in bech32, so we should decode them to get bytes, add prefixes and compose a final key:
1. Decode bech32 encoded address `neutron1qnk2n4nlkpw9xfqntladh74w6ujtulwn6dwq8z` to get hex representation:
```bash
foo@bar % neutrond debug addr neutron1qnk2n4nlkpw9xfqntladh74w6ujtulwn6dwq8z
Address: [4 236 169 214 127 176 92 83 36 19 95 250 219 250 174 215 36 190 125 211]
Address (hex): 04ECA9D67FB05C5324135FFADBFAAED724BE7DD3
Bech32 Acc: neutron1qnk2n4nlkpw9xfqntladh74w6ujtulwn6dwq8z
Bech32 Val: neutronvaloper1qnk2n4nlkpw9xfqntladh74w6ujtulwnqshepx
```
2. Decode bech32 encoded address `neutronvaloper1qnk2n4nlkpw9xfqntladh74w6ujtulwnqshepx` to get hex representation:
```bash
foo@bar % neutrond debug addr neutron1qnk2n4nlkpw9xfqntladh74w6ujtulwn6dwq8z
Address: [4 236 169 214 127 176 92 83 36 19 95 250 219 250 174 215 36 190 125 211]
Address (hex): 04ECA9D67FB05C5324135FFADBFAAED724BE7DD3
Bech32 Acc: neutron1qnk2n4nlkpw9xfqntladh74w6ujtulwn6dwq8z
Bech32 Val: neutronvaloper1qnk2n4nlkpw9xfqntladh74w6ujtulwnqshepx
```
Results are the same because it's a self-delegation of a validator.

3. Now we need to add length (in bytes) prefixes to these addresses (we can do it easily in python console).:
```python
>>> hex(len("04ECA9D67FB05C5324135FFADBFAAED724BE7DD3")//2) + "04ECA9D67FB05C5324135FFADBFAAED724BE7DD3"
'0x1404ECA9D67FB05C5324135FFADBFAAED724BE7DD3'
```
4. Now we can compose a full KV key for to get a delegation (we don't `0x` prefix in hex values):
```
0x31 + 0x1404ECA9D67FB05C5324135FFADBFAAED724BE7DD3 + 0x1404ECA9D67FB05C5324135FFADBFAAED724BE7DD3 = 311404eca9d67fb05c5324135ffadbfaaed724be7dd31404eca9d67fb05c5324135ffadbfaaed724be7dd3
```
5. And finally we need a module store key to tell the relayer where the KV values are stored. In case of staking module of Cosmos-SDK, it's just `staking`. We just add it to our finale key as `staking/` with slash:
```
staking/311404eca9d67fb05c5324135ffadbfaaed724be7dd31404eca9d67fb05c5324135ffadbfaaed724be7dd3
```
</details>


2. Register an interchain query to search transactions on remote chain by some event (in this case it'll try to find all transactions for bank transfer):
```
neutrond tx interchainqueries register-interchain-query test-2 connection-0 1 tx '{"message.module": "bank"}' --from demowallet1 --gas 10000000 --gas-adjustment 1.4 --gas-prices 0.5stake --broadcast-mode block --chain-id test-1 --keyring-backend test --home ./data/test-1 --node tcp://127.0.0.1:16657
```

3. Execute a bank transfer transaction on remote chain for the interchain query above:
```
neutrond tx bank send $(neutrond keys show demowallet2 -a --keyring-backend test --home ./data/test-2) neutron1mjk79fjjgpplak5wq838w0yd982gzkyf8fxu8u 1000stake --from demowallet2 --gas 10000000 --gas-adjustment 1.4 --gas-prices 0.5stake --broadcast-mode block --chain-id test-2 --keyring-backend test --home ./data/test-2 --node tcp://127.0.0.1:26657
```

4. After relayer process query events, you can see submitted results:
* Updated info about registered queries:
```shell
neutrond query interchainqueries registered-queries --node tcp://127.0.0.1:16657
```

* Result for KV storage query (in our case DelegatorDelegations and this query id is 1):
```shell
neutrond query interchainqueries query-result 1 --node tcp://127.0.0.1:16657
```

* Result for transactions search query (in our case MsgBankSend and this query id is 2, and we set limit and offset to 1 and 100 respectively):
```shell
neutrond query interchainqueries query-transactions-search-result 2 1 100 --node tcp://127.0.0.1:16657
```


### terminal 3

1. `git clone git@github.com:neutron-org/cosmos-query-relayer.git`
2. `cd cosmos-query-relayer`
3. `cp configs/dev.example.2-neutron-chains.yml configs/dev.yml`
4. `make dev`
