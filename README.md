![Neutron](https://github.com/neutron-org/neutron-docs/blob/1db1e92098c915ae8ad4defc0bd30ef549175201/static/img/neutron_wide_logo.png)
# Neutron


**NOTE: current version is a WIP.**

## Requirements
* Go 1.20
* Ignite Cli
* Hermes IBC Relayer

### How to install Ignite CLI

```shell
curl https://get.ignite.com/cli! | bash
```

### How to install Hermes IBC Relayer
Install binary
```
cargo install --version 1.4.0 ibc-relayer-cli --bin hermes --locked
```

or install from sources
```shell
git clone https://github.com/informalsystems/hermes.git
cd hermes
git checkout v1.4.0
cargo build --release --bin hermes
```

## Build and Install Neutron Zone

```shell
make install
```

## Run local testnet node instances connected via IBC

### Bootstrap two chains and create an IBC connection

In order to bootstrap two chains it is required to clone gaiad v9.0.3 next to `neutron` repository and install it.

```shell
git clone https://github.com/cosmos/gaia.git
cd gaia
git checkout v9.0.3
make install
```

after that you can turn back into `neutron` directory and run:

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
@docker rmi neutron-node
```

## Documentation

You can check the documentation here: https://neutron-org.github.io/neutron-docs/

> Note: we are going to open & deploy the docs soon.

## Examples

You can check out the example contracts here: https://github.com/neutron-org/neutron-contracts

## Tests

Integration tests are implemented here: https://github.com/neutron-org/neutron-integration-tests
