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

## Run local testnet node instance

```shell
./start.sh
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