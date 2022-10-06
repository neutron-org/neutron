# x/gov
The Neutron's governance is a wrapper over original cosmos sdk `gov` module with a some key difference. For the modules are very similar, overview section describes only the differences. Learn more about the gov module (and therefore about the Neutron's one) by the link to the cosmos documentation site: https://docs.cosmos.network/master/modules/gov/

What makes this module different from the original one:
- Staking no more affects user's voting power
- Voting power is now calculates and stores in smart-contract
- Tally logic is modified

Read about these features below to make a better understanding.

## Staking no more affects user's voting power

The original gov module computes voting power on proposals using each user's delegations. Since the Neutron's plan is to not use standard staking & validators, it's necessary to remove rudimentary usage of staking module. Instead of this, however, we are introduced an alternative
## Voting power is now calculated and stored in smart-contract

We use cosm-wasm contract which implements several methods:
```rust
pub fn query_voting_power(deps: Deps, user_addr: Addr) -> StdResult<VotingPowerResponse> {...}
```

```rust
pub fn query_voting_powers(deps: Deps) -> StdResult<Vec<VotingPowerResponse>>  {...}  
```
where ```VotingPowerResponse``` is

```rust
pub struct VotingPowerResponse {
    /// Address of the user
    pub user: String,
    /// The user's current voting power, i.e. the amount of NTRN tokens locked in voting contract
    pub voting_power: Uint128,
}
```
currently neutron-core uses only `query_voting_powers`, but `query_voting_power` seems to be useful in future

## Tally logic is modified
Tally interface hasn't changed, but instead of calculating voting results by staked tokens, it uses an above contract's query
```golang
// GetTokensInDao queries the voting contract for an array of users who have tokens locked in the
// contract and their respective amount, as well as computing the total amount of locked tokens.
func GetTokensInDao(ctx sdk.Context, k wasmtypes.ViewKeeper, contractAddr sdk.AccAddress) (map[string]sdk.Int, sdk.Int, error) {...}
```
