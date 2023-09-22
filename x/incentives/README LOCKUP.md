# Stakeup

## Abstract

Stakeup module provides an interface for users to stake tokens (also known as bonding) into the module to get incentives.

After tokens have been added to a specific pool and turned into LP shares through the GAMM module, users can then stake these LP shares with a specific duration in order to begin earing rewards.

To unstake these LP shares, users must trigger the unstake timer and wait for the unstake period that was set initially to be completed. After the unstake period is over, users can turn LP shares back into their respective share of tokens.

This module provides interfaces for other modules to iterate the stakes efficiently and grpc query to check the status of staked coins.

## Contents

1. **[Concept](#concepts)**
2. **[State](#state)**
3. **[Messages](#messages)**
4. **[Events](#events)**
5. **[Keeper](#keeper)**
6. **[Hooks](#hooks)**
7. **[Queries](#queries)**
8. **[Transactions](#transactions)**
9. **[Params](#params)**
10. **[Endbstakeer](#endbstakeer)**

## Concepts

The purpose of `stakeup` module is to provide the functionality to stake
tokens for specific period of time for LP token stakers to get
incentives.

To unstake these LP shares, users must trigger the unstake timer and wait for the unstake period that was set initially to be completed. After the unstake period is over, users can turn LP shares back into their respective share of tokens.

This module provides interfaces for other modules to iterate the stakes efficiently and grpc query to check the status of staked coins.

There are currently three incentivize stakeup periods; `1 day` (24h), `1 week` (168h), and `2 weeks` (336h). When staking tokens in the 2 week period, the liquidity provider is effectively earning rewards for a combination of the 1 day, 1 week, and 2 week bonding periods.

The 2 week period refers to how long it takes to unbond the LP shares. The liquidity provider can keep their LP shares bonded to the 2 week stakeup period indefinitely. Unbonding is only required when the liquidity provider desires access to the underlying assets.

If the liquidity provider begins the unbonding process for their 2 week bonded LP shares, they will earn rewards for all three bonding periods during the first day of unbonding.

After the first day passes, they will only receive rewards for the 1 day and 1 week stakeup periods. After seven days pass, they will only receive the 1 day rewards until the 2 weeks is complete and their LP shares are unstaked. The below chart is a visual example of what was just explained.

<br/>
<p style="text-align:center;">
<img src="/img/bonding.png" height="300"/>
</p>

</br>
</br>

## State

### Staked coins management

Staked coins are all stored in module account for `stakeup` module which
is called `StakePool`. When user stake coins within `stakeup` module, it's
moved from user account to `StakePool` and a record (`PeriodStake` struct)
is created.

Once the period is over, user can withdraw it at anytime from
`StakePool`. User can withdraw by PeriodStake ID or withdraw all
`UnstakeableCoins` at a time.

### Period Stake

A `PeriodStake` is a single unit of stake by period. It's a record of
staked coin at a specific time. It stores owner, duration, unstake time
and the amount of coins staked.

``` {.go}
type PeriodStake struct {
  ID         uint64
  Owner      sdk.AccAddress
  Duration   time.Duration
  UnstakeTime time.Time
  Coins      sdk.Coins
}
```

All stakes are stored on the KVStore as value at
`{KeyPrefixPeriodStake}{ID}` key.

### Period stake reference queues

To provide time efficient queries, several reference queues are managed
by denom, unstake time, and duration. There are two big queues to store
the stake references. (`a_prefix_key`)

1. Stake references that hasn't started with unstaking yet has prefix of
    `KeyPrefixNotUnstaking`.
2. Stake references that has started unstaking already has prefix of
    `KeyPrefixUnstaking`.
3. Stake references that has withdrawn, it's removed from the store.

Regardless the stake has started unstaking or not, it stores below
references. (`b_prefix_key`)

1. `{KeyPrefixStakeDuration}{Duration}`
2. `{KeyPrefixAccountStakeDuration}{Owner}{Duration}`
3. `{KeyPrefixDenomStakeDuration}{Denom}{Duration}`
4. `{KeyPrefixAccountDenomStakeDuration}{Owner}{Denom}{Duration}`

If the stake is unstaking, it also stores the below references.

1. `{KeyPrefixStakeTimestamp}{StakeEndTime}`
2. `{KeyPrefixAccountStakeTimestamp}{Owner}{StakeEndTime}`
3. `{KeyPrefixDenomStakeTimestamp}{Denom}{StakeEndTime}`
4. `{KeyPrefixAccountDenomStakeTimestamp}{Owner}{Denom}{StakeEndTime}`

For end time keys, they are converted to sortable string by using
`sdk.FormatTimeBytes` function.

**Note:** Additionally, for stakes that hasn't started unstaking yet, it
stores accumulation store for efficient rewards distribution mechanism.

For reference management, `addStakeRefByKey` function is used a lot. Here
key is the prefix key to be used for iteration. It is combination of two
prefix keys.(`{a_prefix_key}{b_prefix_key}`)

``` {.go}
// addStakeRefByKey make a stakeID iterable with the prefix `key`
func (k Keeper) addStakeRefByKey(ctx sdk.Context, key []byte, stakeID uint64) error {
 store := ctx.KVStore(k.storeKey)
 stakeIDBz := sdk.Uint64ToBigEndian(stakeID)
 endKey := combineKeys(key, stakeIDBz)
 if store.Has(endKey) {
  return fmt.Errorf("stake with same ID exist: %d", stakeID)
 }
 store.Set(endKey, stakeIDBz)
 return nil
}
```

## Messages

### Stake Tokens

`MsgStake` can be submitted by any token holder via a
`MsgStake` transaction.

``` {.go}
type MsgStake struct {
 Owner    sdk.AccAddress
 Duration time.Duration
 Coins    sdk.Coins
}
```

**State modifications:**

- Validate `Owner` has enough tokens
- Generate new `PeriodStake` record
- Save the record inside the keeper's time basis unstake queue
- Transfer the tokens from the `Owner` to stakeup `ModuleAccount`.

### Begin Unstake of all stakes

Once time is over, users can withdraw unstaked coins from stakeup
`ModuleAccount`.

``` {.go}
type MsgBeginUnstakingAll struct {
 Owner string
}
```

**State modifications:**

- Fetch all unstakeable `PeriodStake`s that has not started unstaking
    yet
- Set `PeriodStake`'s unstake time
- Remove stake references from `NotUnstaking` queue
- Add stake references to `Unstaking` queue

### Begin unstake for a stake

Once time is over, users can withdraw unstaked coins from stakeup
`ModuleAccount`.

``` {.go}
type MsgUnstake struct {
 Owner string
 ID    uint64
}
```

**State modifications:**

- Check `PeriodStake` with `ID` specified by `MsgUnstake` is not
    started unstaking yet
- Set `PeriodStake`'s unstake time
- Remove stake references from `NotUnstaking` queue
- Add stake references to `Unstaking` queue

Note: If another module needs past `PeriodStake` item, it can log the
details themselves using the hooks.

## Events

The stakeup module emits the following events:

### Handlers

#### MsgStake

|  Type          | Attribute Key     | Attribute Value  |
|  --------------| ------------------| -----------------|
|  stake\_tokens  | period\_stake\_id  | {periodStakeID}   |
|  stake\_tokens  | owner             | {owner}          |
|  stake\_tokens  | amount            | {amount}         |
|  stake\_tokens  | duration          | {duration}       |
|  stake\_tokens  | unstake\_time      | {unstakeTime}     |
|  message       | action            | stake\_tokens     |
|  message       | sender            | {owner}          |
|  transfer      | recipient         | {moduleAccount}  |
|  transfer      | sender            | {owner}          |
|  transfer      | amount            | {amount}         |

#### MsgUnstake

|  Type           | Attribute Key     | Attribute Value   |
|  ---------------| ------------------| ------------------|
|  begin\_unstake  | period\_stake\_id  | {periodStakeID}    |
|  begin\_unstake  | owner             | {owner}           |
|  begin\_unstake  | amount            | {amount}          |
|  begin\_unstake  | duration          | {duration}        |
|  begin\_unstake  | unstake\_time      | {unstakeTime}      |
|  message        | action            | begin\_unstaking  |
|  message        | sender            | {owner}           |

#### MsgBeginUnstakingAll

|  Type                | Attribute Key     | Attribute Value        |
|  --------------------| ------------------| -----------------------|
|  begin\_unstake\_all  | owner             | {owner}                |
|  begin\_unstake\_all  | unstaked\_coins   | {unstakedCoins}        |
|  begin\_unstake       | period\_stake\_id  | {periodStakeID}         |
|  begin\_unstake       | owner             | {owner}                |
|  begin\_unstake       | amount            | {amount}               |
|  begin\_unstake       | duration          | {duration}             |
|  begin\_unstake       | unstake\_time      | {unstakeTime}           |
|  message             | action            | begin\_unstaking\_all  |
|  message             | sender            | {owner}                |

### Endbstakeer

#### Automatic withdraw when unstake time mature

|  Type            | Attribute Key     | Attribute Value  |
|  ----------------| ------------------| -----------------|
|  message         | action            | unstake\_tokens   |
|  message         | sender            | {owner}          |
|  transfer\[\]    | recipient         | {owner}          |
|  transfer\[\]    | sender            | {moduleAccount}  |
|  transfer\[\]    | amount            | {unstakeAmount}   |
|  unstake\[\]      | period\_stake\_id  | {owner}          |
|  unstake\[\]      | owner             | {stakeID}         |
|  unstake\[\]      | duration          | {stakeDuration}   |
|  unstake\[\]      | unstake\_time      | {unstakeTime}     |
|  unstake\_tokens  | owner             | {owner}          |
|  unstake\_tokens  | unstaked\_coins   | {totalAmount}    |

## Keepers

### Stakeup Keeper

Stakeup keeper provides utility functions to store stake queues and query
stakes.

```go
// Keeper is the interface for stakeup module keeper
type Keeper interface {
    // GetModuleBalance Returns full balance of the module
    GetModuleBalance(sdk.Context) sdk.Coins
    // GetModuleStakedCoins Returns staked balance of the module
    GetModuleStakedCoins(sdk.Context) sdk.Coins
    // GetAccountUnstakeableCoins Returns whole unstakeable coins which are not withdrawn yet
    GetAccountUnstakeableCoins(sdk.Context, addr sdk.AccAddress) sdk.Coins
    // GetAccountUnstakingCoins Returns whole unstaking coins
    GetAccountUnstakingCoins(sdk.Context, addr sdk.AccAddress) sdk.Coins
    // GetAccountStakedCoins Returns a staked coins that can't be withdrawn
    GetAccountStakedCoins(sdk.Context, addr sdk.AccAddress) sdk.Coins
    // GetAccountStakedPastTime Returns the total stakes of an account whose unstake time is beyond timestamp
    GetAccountStakedPastTime(sdk.Context, addr sdk.AccAddress, timestamp time.Time) []types.PeriodStake
    // GetAccountUnstakedBeforeTime Returns the total unstakes of an account whose unstake time is before timestamp
    GetAccountUnstakedBeforeTime(sdk.Context, addr sdk.AccAddress, timestamp time.Time) []types.PeriodStake
    // GetAccountStakedPastTimeDenom is equal to GetAccountStakedPastTime but denom specific
    GetAccountStakedPastTimeDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, timestamp time.Time) []types.PeriodStake

    // GetAccountStakedLongerDuration Returns account staked with duration longer than specified
    GetAccountStakedLongerDuration(sdk.Context, addr sdk.AccAddress, duration time.Duration) []types.PeriodStake
    // GetAccountStakedLongerDurationDenom Returns account staked with duration longer than specified with specific denom
    GetAccountStakedLongerDurationDenom(sdk.Context, addr sdk.AccAddress, denom string, duration time.Duration) []types.PeriodStake
    // GetStakesPastTimeDenom Returns the stakes whose unstake time is beyond timestamp
    GetStakesPastTimeDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, timestamp time.Time) []types.PeriodStake
    // GetStakesLongerThanDurationDenom Returns the stakes whose unstake duration is longer than duration
    GetStakesLongerThanDurationDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, duration time.Duration) []types.PeriodStake
    // GetStakeByID Returns stake from stakeID
    GetStakeByID(sdk.Context, stakeID uint64) (*types.PeriodStake, error)
    // GetPeriodStakes Returns the period stakes on pool
    GetPeriodStakes(sdk.Context) ([]types.PeriodStake, error)
    // UnstakeAllUnstakeableCoins Unstake all unstakeable coins
    UnstakeAllUnstakeableCoins(sdk.Context, account sdk.AccAddress) (sdk.Coins, error)
    // StakeTokens stake tokens from an account for specified duration
    StakeTokens(sdk.Context, owner sdk.AccAddress, coins sdk.Coins, duration time.Duration) (types.PeriodStake, error)
    // AddTokensToStake stakes more tokens into a stakeup
    AddTokensToStake(ctx sdk.Context, owner sdk.AccAddress, stakeID uint64, coins sdk.Coins) (*types.PeriodStake, error)
    // Stake is a utility to stake coins into module account
    Stake(sdk.Context, stake types.PeriodStake) error
    // Unstake is a utility to unstake coins from module account
    Unstake(sdk.Context, stake types.PeriodStake) error
```

## Hooks

In this section we describe the "hooks" that `stakeup` module provide for
other modules.

### Tokens Staked

On stake/unstake events, stakeup module execute hooks for other modules to
make following actions.

``` go
  OnTokenStaked(ctx sdk.Context, address sdk.AccAddress, stakeID uint64, amount sdk.Coins, stakeDuration time.Duration, unstakeTime time.Time)
  OnTokenUnstaked(ctx sdk.Context, address sdk.AccAddress, stakeID uint64, amount sdk.Coins, stakeDuration time.Duration, unstakeTime time.Time)
```

## Parameters

The stakeup module contains the following parameters:

| Key                    | Type            | Example |
| ---------------------- | --------------- | ------- |

Note: Currently no parameters are set for `stakeup` module, we will need
to move stakeable durations from incentives module to stakeup module.

## Endbstakeer

### Withdraw tokens after unstake time mature

Once time is over, endbstakeer withdraw coins from matured stakes and
coins are sent from stakeup `ModuleAccount`.

**State modifications:**

- Fetch all unstakeable `PeriodStake`s that `Owner` has not withdrawn
    yet
- Remove `PeriodStake` records from the state
- Transfer the tokens from stakeup `ModuleAccount` to the
    `MsgUnstakeTokens.Owner`.

### Remove synthetic stakes after removal time mature

For synthetic stakes, no coin movement is made, but stakeup record and
reference queues are removed.


## Transactions

### stake-tokens

Bond tokens in a LP for a set duration

```sh
dualityd tx stakeup stake-tokens [tokens] --duration --from --chain-id
```

::: details Example

To stakeup `15.527546134174465309gamm/pool/3` tokens for a `one day` bonding period from `WALLET_NAME` on the osmosis mainnet:

```bash
dualityd tx stakeup stake-tokens 15527546134174465309gamm/pool/3 --duration="24h" --from WALLET_NAME --chain-id osmosis-1
```

To stakeup `25.527546134174465309gamm/pool/13` tokens for a `one week` bonding period from `WALLET_NAME` on the osmosis testnet:

```bash
dualityd tx stakeup stake-tokens 25527546134174465309gamm/pool/13 --duration="168h" --from WALLET_NAME --chain-id osmo-test-4
```

To stakeup `35.527546134174465309 gamm/pool/197` tokens for a `two week` bonding period from `WALLET_NAME` on the osmosis mainnet:

```bash
dualityd tx stakeup stake-tokens 35527546134174465309gamm/pool/197 --duration="336h" --from WALLET_NAME --chain-id osmosis-1
```
:::


### begin-unstake-by-id

Begin the unbonding process for tokens given their unique stake ID

```sh
dualityd tx stakeup begin-unstake-by-id [id] --from --chain-id
```

::: details Example

To begin the unbonding time for all bonded tokens under id `75` from `WALLET_NAME` on the osmosis mainnet:

```bash
dualityd tx stakeup begin-unstake-by-id 75 --from WALLET_NAME --chain-id osmosis-1
```
:::
::: warning Note
The ID corresponds to the unique ID given to your stakeup transaction (explained more in stake-by-id section)
:::

### begin-unstake-tokens

Begin unbonding process for all bonded tokens in a wallet

```sh
dualityd tx stakeup begin-unstake-tokens --from --chain-id
```

::: details Example

To begin unbonding time for ALL pools and ALL bonded tokens in `WALLET_NAME` on the osmosis mainnet:


```bash
dualityd tx stakeup begin-unstake-tokens --from=WALLET_NAME --chain-id=osmosis-1 --yes
```
:::

## Queries

In this section we describe the queries required on grpc server.

``` protobuf
// Query defines the gRPC QueryServer service.
service Query {
    // Return full balance of the module
 rpc ModuleBalance(ModuleBalanceRequest) returns (ModuleBalanceResponse);
 // Return staked balance of the module
 rpc ModuleStakedAmount(ModuleStakedAmountRequest) returns (ModuleStakedAmountResponse);

 // Returns unstakeable coins which are not withdrawn yet
 rpc AccountUnstakeableCoins(AccountUnstakeableCoinsRequest) returns (AccountUnstakeableCoinsResponse);
 // Returns unstaking coins
   rpc AccountUnstakingCoins(AccountUnstakingCoinsRequest) returns (AccountUnstakingCoinsResponse) {}
 // Return a staked coins that can't be withdrawn
 rpc AccountStakedCoins(AccountStakedCoinsRequest) returns (AccountStakedCoinsResponse);

 // Returns staked records of an account with unstake time beyond timestamp
 rpc AccountStakedPastTime(AccountStakedPastTimeRequest) returns (AccountStakedPastTimeResponse);
 // Returns staked records of an account with unstake time beyond timestamp excluding tokens started unstaking
 rpc AccountStakedPastTimeNotUnstakingOnly(AccountStakedPastTimeNotUnstakingOnlyRequest) returns (AccountStakedPastTimeNotUnstakingOnlyResponse) {}
 // Returns unstaked records with unstake time before timestamp
 rpc AccountUnstakedBeforeTime(AccountUnstakedBeforeTimeRequest) returns (AccountUnstakedBeforeTimeResponse);

 // Returns stake records by address, timestamp, denom
 rpc AccountStakedPastTimeDenom(AccountStakedPastTimeDenomRequest) returns (AccountStakedPastTimeDenomResponse);
 // Returns stake record by id
 rpc StakedByID(StakedRequest) returns (StakedResponse);

 // Returns account staked records with longer duration
 rpc AccountStakedLongerDuration(AccountStakedLongerDurationRequest) returns (AccountStakedLongerDurationResponse);
 // Returns account staked records with longer duration excluding tokens started unstaking
   rpc AccountStakedLongerDurationNotUnstakingOnly(AccountStakedLongerDurationNotUnstakingOnlyRequest) returns (AccountStakedLongerDurationNotUnstakingOnlyResponse) {}
 // Returns account's staked records for a denom with longer duration
 rpc AccountStakedLongerDurationDenom(AccountStakedLongerDurationDenomRequest) returns (AccountStakedLongerDurationDenomResponse);

 // Returns account staked records with a specific duration
 rpc AccountStakedDuration(AccountStakedDurationRequest) returns (AccountStakedDurationResponse);
}
```

### account-staked-beforetime

Query an account's unstaked records after a specified time (UNIX) has passed

In other words, if an account unstaked all their bonded tokens the moment the query was executed, only the stakes that would have completed their bond time requirement by the time the `TIMESTAMP` is reached will be returned.

::: details Example

In this example, the current UNIX time is `1639776682`, 2 days from now is approx `1639971082`, and 15 days from now is approx `1641094282`.

An account's `ADDRESS` is staked in both the `1 day` and `1 week` gamm/pool/3. To query the `ADDRESS` with a timestamp 2 days from now `1639971082`:

```bash
dualityd query stakeup account-staked-beforetime ADDRESS 1639971082
```

In this example will output the `1 day` stake but not the `1 week` stake:

```bash
stakes:
- ID: "571839"
  coins:
  - amount: "15527546134174465309"
    denom: gamm/pool/3
  duration: 24h
  end_time: "2021-12-18T23:32:58.900715388Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

If querying the same `ADDRESS` with a timestamp 15 days from now `1641094282`:

```bash
dualityd query stakeup account-staked-beforetime ADDRESS 1641094282
```

In this example will output both the `1 day` and `1 week` stake:

```bash
stakes:
- ID: "572027"
  coins:
  - amount: "16120691802759484268"
    denom: gamm/pool/3
  duration: 604800.000006193s
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
- ID: "571839"
  coins:
  - amount: "15527546134174465309"
    denom: gamm/pool/3
  duration: 24h
  end_time: "2021-12-18T23:32:58.900715388Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```
:::


### account-staked-coins

Query an account's staked (bonded) LP tokens

```sh
dualityd query stakeup account-staked-coins [address]
```

:::: details Example

```bash
dualityd query stakeup account-staked-coins osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

An example output:

```bash
coins:
- amount: "413553955105681228583"
  denom: gamm/pool/1
- amount: "32155370994266157441309"
  denom: gamm/pool/10
- amount: "220957857520769912023"
  denom: gamm/pool/3
- amount: "31648237936933949577"
  denom: gamm/pool/42
- amount: "14162624050980051053569"
  denom: gamm/pool/5
- amount: "1023186951315714985896914"
  denom: gamm/pool/9
```
::: warning Note
All GAMM amounts listed are 10^18. Move the decimal place to the left 18 places to get the GAMM amount listed in the GUI.

You may also specify a --height flag to see bonded LP tokens at a specified height (note: if running a pruned node, this may result in an error)
:::
::::

### account-staked-longer-duration

Query an account's staked records that are greater than or equal to a specified stake duration

```sh
dualityd query stakeup account-staked-longer-duration [address] [duration]
```

::: details Example

Here is an example of querying an `ADDRESS` for all `1 day` or greater bonding periods:

```bash
dualityd query stakeup account-staked-longer-duration osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259 24h
```

An example output:

```bash
stakes:
- ID: "572027"
  coins:
  - amount: "16120691802759484268"
    denom: gamm/pool/3
  duration: 604800.000006193s
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
- ID: "571839"
  coins:
  - amount: "15527546134174465309"
    denom: gamm/pool/3
  duration: 24h
  end_time: "2021-12-18T23:32:58.900715388Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```
:::


### account-staked-longer-duration-denom

Query an account's staked records for a denom that is staked equal to or greater than the specified duration AND match a specified denom

```sh
dualityd query stakeup account-staked-longer-duration-denom [address] [duration] [denom]
```

::: details Example

Here is an example of an `ADDRESS` that is staked in both the `1 day` and `1 week` for both the gamm/pool/3 and gamm/pool/1, then queries the `ADDRESS` for all bonding periods equal to or greater than `1 day` for just the gamm/pool/3:

```bash
dualityd query stakeup account-staked-longer-duration-denom osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259 24h gamm/pool/3
```

An example output:

```bash
stakes:
- ID: "571839"
  coins:
  - amount: "15527546134174465309"
    denom: gamm/pool/3
  duration: 24h
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
- ID: "572027"
  coins:
  - amount: "16120691802759484268"
    denom: gamm/pool/3
  duration: 604800.000006193s
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

As shown, the gamm/pool/3 is returned but not the gamm/pool/1 due to the denom filter.
:::


### account-staked-longer-duration-not-unstaking

Query an account's staked records for a denom that is staked equal to or greater than the specified duration AND is not in the process of being unstaked

```sh
dualityd query stakeup account-staked-longer-duration-not-unstaking [address] [duration]
```

::: details Example

Here is an example of an `ADDRESS` that is staked in both the `1 day` and `1 week` gamm/pool/3, begins unstaking process for the `1 day` bond, and queries the `ADDRESS` for all bonding periods equal to or greater than `1 day` that are not unbonding:

```bash
dualityd query stakeup account-staked-longer-duration-not-unstaking osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259 24h
```

An example output:

```bash
stakes:
- ID: "571839"
  coins:
  - amount: "16120691802759484268"
    denom: gamm/pool/3
  duration: 604800.000006193s
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

The `1 day` bond does not show since it is in the process of unbonding.
:::


### account-staked-pasttime

Query the staked records of an account with the unstake time beyond timestamp (UNIX)

```bash
dualityd query stakeup account-staked-pasttime [address] [timestamp]
```

::: details Example

Here is an example of an account that is staked in both the `1 day` and `1 week` gamm/pool/3. In this example, the UNIX time is currently `1639776682` and queries an `ADDRESS` for UNIX time two days later from the current time (which in this example would be `1639971082`)

```bash
dualityd query stakeup account-staked-pasttime osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259 1639971082
```

The example output:

```bash
stakes:
- ID: "572027"
  coins:
  - amount: "16120691802759484268"
    denom: gamm/pool/3
  duration: 604800.000006193s
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

Note that the `1 day` stake ID did not display because, if the unbonding time began counting down from the time the command was executed, the bonding period would be complete before the two day window given by the UNIX timestamp input.
:::


### account-staked-pasttime-denom

Query the staked records of an account with the unstake time beyond timestamp (unix) and filter by a specific denom

```bash
dualityd query stakeup account-staked-pasttime-denom osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259 [timestamp] [denom]
```

::: details Example

Here is an example of an account that is staked in both the `1 day` and `1 week` gamm/pool/3 and `1 day` and `1 week` gamm/pool/1. In this example, the UNIX time is currently `1639776682` and queries an `ADDRESS` for UNIX time two days later from the current time (which in this example would be `1639971082`) and filters for gamm/pool/3

```bash
dualityd query stakeup account-staked-pasttime-denom osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259 1639971082 gamm/pool/3
```

The example output:

```bash
stakes:
- ID: "572027"
  coins:
  - amount: "16120691802759484268"
    denom: gamm/pool/3
  duration: 604800.000006193s
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

Note that the `1 day` stake ID did not display because, if the unbonding time began counting down from the time the command was executed, the bonding period would be complete before the two day window given by the UNIX timestamp input. Additionally, neither of the `1 day` or `1 week` stake IDs for gamm/pool/1 showed due to the denom filter.
:::


### account-staked-pasttime-not-unstaking

Query the staked records of an account with the unstake time beyond timestamp (unix) AND is not in the process of unstaking

```sh
dualityd query stakeup account-staked-pasttime [address] [timestamp]
```

::: details Example

Here is an example of an account that is staked in both the `1 day` and `1 week` gamm/pool/3. In this example, the UNIX time is currently `1639776682` and queries an `ADDRESS` for UNIX time two days later from the current time (which in this example would be `1639971082`) AND is not unstaking:

```bash
dualityd query stakeup account-staked-pasttime osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259 1639971082
```

The example output:

```bash
stakes:
- ID: "572027"
  coins:
  - amount: "16120691802759484268"
    denom: gamm/pool/3
  duration: 604800.000006193s
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

Note that the `1 day` stake ID did not display because, if the unbonding time began counting down from the time the command was executed, the bonding period would be complete before the two day window given by the UNIX timestamp input. Additionally, if ID 572027 were to begin the unstaking process, the query would have returned blank.
:::


### account-unstakeable-coins

Query an address's LP shares that have completed the unstaking period and are ready to be withdrawn

```bash
dualityd query stakeup account-unstakeable-coins ADDRESS
```



### account-unstaking-coins

Query an address's LP shares that are currently unstaking

```sh
dualityd query stakeup account-unstaking-coins [address]
```

::: details Example

```bash
dualityd query stakeup account-unstaking-coins osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

Example output:

```bash
coins:
- amount: "15527546134174465309"
  denom: gamm/pool/3
```
:::


### stake-by-id

Query a stake record by its ID

```sh
dualityd query stakeup stake-by-id [id]
```

::: details Example

Every time a user bonds tokens to an LP, a unique stake ID is created for that transaction.

Here is an example viewing the stake record for ID 9:

```bash
dualityd query stakeup stake-by-id 9
```

And its output:

```bash
stake:
  ID: "9"
  coins:
  - amount: "2449472670508255020346507"
    denom: gamm/pool/2
  duration: 336h
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo16r39ghhwqjcwxa8q3yswlz8jhzldygy66vlm82
```

In summary, this shows wallet `osmo16r39ghhwqjcwxa8q3yswlz8jhzldygy66vlm82` bonded `2449472.670 gamm/pool/2` LP shares for a `2 week` staking period.
:::


### module-balance

Query the balance of all LP shares (bonded and unbonded)

```sh
dualityd query stakeup module-balance
```

::: details Example

```bash
dualityd query stakeup module-balance
```

An example output:

```bash
coins:
- amount: "118851922644152734549498647"
  denom: gamm/pool/1
- amount: "2165392672114512349039263626"
  denom: gamm/pool/10
- amount: "9346769826591025900804"
  denom: gamm/pool/13
- amount: "229347389639275840044722315"
  denom: gamm/pool/15
- amount: "81217698776012800247869"
  denom: gamm/pool/183
- amount: "284253336860259874753775"
  denom: gamm/pool/197
- amount: "664300804648059580124426710"
  denom: gamm/pool/2
- amount: "5087102794776326441530430"
  denom: gamm/pool/22
- amount: "178900843925960029029567880"
  denom: gamm/pool/3
- amount: "64845148811263846652326124"
  denom: gamm/pool/4
- amount: "177831279847453210600513"
  denom: gamm/pool/42
- amount: "18685913727862493301261661338"
  denom: gamm/pool/5
- amount: "23579028640963777558149250419"
  denom: gamm/pool/6
- amount: "1273329284855460149381904976"
  denom: gamm/pool/7
- amount: "625252103927082207683116933"
  denom: gamm/pool/8
- amount: "1148475247281090606949382402"
  denom: gamm/pool/9
```
:::


### module-staked-amount

Query the balance of all bonded LP shares

```sh
dualityd query stakeup module-staked-amount
```

::: details Example

```bash
dualityd query stakeup module-staked-amount
```

An example output:

```bash

  "coins":
    {
      "denom": "gamm/pool/1",
      "amount": "247321084020868094262821308"
    },
    {
      "denom": "gamm/pool/10",
      "amount": "2866946821820635047398966697"
    },
    {
      "denom": "gamm/pool/13",
      "amount": "9366580741745176812984"
    },
    {
      "denom": "gamm/pool/15",
      "amount": "193294911294343602187680438"
    },
    {
      "denom": "gamm/pool/183",
      "amount": "196722012808526595790871"
    },
    {
      "denom": "gamm/pool/197",
      "amount": "1157025085661167198918241"
    },
    {
      "denom": "gamm/pool/2",
      "amount": "633051132033131378888258047"
    },
    {
      "denom": "gamm/pool/22",
      "amount": "3622601406125950733194696"
    },
...

```

NOTE: This command seems to only work on gRPC and on CLI returns an EOF error.
:::



### output-all-stakes

Output all stakes into a json file

```sh
dualityd query stakeup output-all-stakes [max stake ID]
```

:::: details Example

This example command outputs stakes 1-1000 and saves to a json file:

```bash
dualityd query stakeup output-all-stakes 1000
```
::: warning Note
If a stakeup has been completed, the stakeup status will show as "0" (or successful) and no further information will be available. To get further information on a completed stake, run the stake-by-id query.
:::
::::


### total-staked-of-denom

Query staked amount for a specific denom in the duration provided

```sh
dualityd query stakeup total-staked-of-denom [denom] --min-duration
```

:::: details Example

This example command outputs the amount of `gamm/pool/2` LP shares that staked in the `2 week` bonding period:

```bash
dualityd query stakeup total-staked-of-denom gamm/pool/2 --min-duration "336h"
```

Which, at the time of this writing outputs `14106985399822075248947045` which is equivalent to `14106985.3998 gamm/pool/2`

NOTE: As of this writing, there is a bug that defaults the min duration to days instead of seconds. Ensure you specify the time in seconds to get the correct response.
:::

## Commands

```sh
# 1 day 100stake stake-tokens command
dualityd tx stakeup stake-tokens 200stake --duration="86400s" --from=validator --chain-id=testing --keyring-backend=test --yes

# 5s 100stake stake-tokens command
dualityd tx stakeup stake-tokens 100stake --duration="5s" --from=validator --chain-id=testing --keyring-backend=test --yes

# begin unstake tokens, NOTE: add more gas when unstaking more than two stakes in a same command
dualityd tx stakeup begin-unstake-tokens --from=validator --gas=500000 --chain-id=testing --keyring-backend=test --yes

# unstake tokens, NOTE: add more gas when unstaking more than two stakes in a same command
dualityd tx stakeup unstake-tokens --from=validator --gas=500000 --chain-id=testing --keyring-backend=test --yes

# unstake specific period stake
dualityd tx stakeup unstake-by-id 1 --from=validator --chain-id=testing --keyring-backend=test --yes

# account balance
dualityd query bank balances $(dualityd keys show -a validator --keyring-backend=test)

# query module balance
dualityd query stakeup module-balance

# query staked amount
dualityd query stakeup module-staked-amount

# query stake by id
dualityd query stakeup stake-by-id 1

# query account unstakeable coins
dualityd query stakeup account-unstakeable-coins $(dualityd keys show -a validator --keyring-backend=test)

# query account stakes by denom past time
dualityd query stakeup account-staked-pasttime-denom $(dualityd keys show -a validator --keyring-backend=test) 1611879610 stake

# query account stakes past time
dualityd query stakeup account-staked-pasttime $(dualityd keys show -a validator --keyring-backend=test) 1611879610

# query account stakes by denom with longer duration
dualityd query stakeup account-staked-longer-duration-denom $(dualityd keys show -a validator --keyring-backend=test) 5.1s stake

# query account stakes with longer duration
dualityd query stakeup account-staked-longer-duration $(dualityd keys show -a validator --keyring-backend=test) 5.1s

# query account staked coins
dualityd query stakeup account-staked-coins $(dualityd keys show -a validator --keyring-backend=test)

# query account stakes before time
dualityd query stakeup account-staked-beforetime $(dualityd keys show -a validator --keyring-backend=test) 1611879610
```
