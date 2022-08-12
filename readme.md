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

## Running with docker

Build the image:
```shell
make build-docker-image
```

After the image is built, you can start/stop with:
```shell
make start-docker-container
make stop-docker-container
```

## Running with docker + relayer

```shell
ssh-add ./.ssh/id_rsa
make start-cosmopark
make stop-cosmopark
```

Make sure you delete node image if you use the whole thing in dev purposes
```shell
@docker rmi neutron_node
```

## Documentation

You can check the documentation here: https://github.com/neutron-org/neutron-docs

> Note: we are going to open & deploy the docs soon.

## Examples

You can check out the example contracts here: https://github.com/neutron-org/neutron-contracts

## Tests

Integration tests are implemented here: https://github.com/neutron-org/neutron-integration-tests