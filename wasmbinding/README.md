This package allows for custom queries and custom messages sends from contract.

### What is supported 

- Queries:
  - InterchainQueryResult - Get the result of a registered interchain query by query_id
  - InterchainAccountAddress - Get the interchain account address by owner_id and connection_id
  - RegisteredInterchainQueries - all set of registered interchain queries.
  - RegisteredInterchainQuery - registered interchain query with specified query_id
- Messages:
  - RegisterInterchainAccount - register an interchain account
  - SubmitTx - submit a transaction for execution on a remote chain
  - RegisterInterchainQuery - register an interchain query
  - UpdateInterchainQuery - update an interchain query
  - RemoveInterchainQuery - remove an interchain query
