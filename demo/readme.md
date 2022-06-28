# Demo scripts

These scripts can be used to test interchain transactions with deployed wasm contracts

## Requirements

1. have wasm contracts be build in `./../lido-interchain-staking-contracts/artifacts`
2. have gaia build and run like `make build && make init && make start-rly`

## How to test

1. run `./demo/init_contracts.sh`
2. make sure you don't see any errors
3. run `./demo/register-interchain-account.sh`
4. make sure you don't see any errors
5. here you can check OpenAck answer from hub contract by `cat ./data/test-*.log | grep Sudo`
6. you should see some `SudoOpenAck received response`
7. now we can execute transactions
8. run `./demo/send_delegate.sh`
9. run `./demo/send_undelegate.sh`
10. check `cat ./data/test-*.log | grep Sudo`
11. you can see something like `{"level":"info","module":"x/interchaintxs","err":null,"response":"/cosmos.staking.v1beta1.MsgUndelegate 1658215452","time":"2022-06-28T07:24:15Z","message":"SudoResponse received response"}`
