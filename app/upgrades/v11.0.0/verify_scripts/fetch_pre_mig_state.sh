#!/usr/bin/env bash
set -euo pipefail

BINARY="${BINARY:-neutrond}"
NODE="${NODE:-http://localhost:26657}"
OUT="${OUT:-pre_mig_state.json}"

DEFAULT_DENOM="untrn"
DNTRN_DENOM="${DNTRN_DENOM:-factory/neutron1frc0p5czd9uaaymdkug2njz7dc7j65jxukp9apmt9260a8egujkspms2t2/udntrn}"

MAIN_DAO_CONTRACT="${MAIN_DAO_CONTRACT:-neutron1suhgf5svhu4usrurvxzlgn54ksxmn8gljarjtxqnapv8kjnp4nrstdxvff}"
PUPPETEER_CONTRACT="${PUPPETEER_CONTRACT:-neutron17jsl4t4hhaw37tnhenskrfntm7mv44wzjr3f990hx4p9r5m0gzdqquhtd3}"
STAKING_REWARDS_CONTRACT="${STAKING_REWARDS_CONTRACT:-neutron1gqq3c735pj6ese3yru5xr6ud0fvxgltxesygvyyzpsrt74v6yg4sgkrgwq}"
IBCRL_CONTRACT="${IBCRL_CONTRACT:-neutron15aqgplxcavqhurr0g5wwtdw6025dknkqwkfh0n46gp2qjl6236cs2yd3nl}"
IBCRL_MULTISIG="${IBCRL_MULTISIG:-neutron1el2rymcsg5wxth2fz2g5l08nue3xhyj3ny5wea3yxwr9f7es8d6smmwrck}"
REVENUE_MODULE_NAME="${REVENUE_MODULE_NAME:-revenue-treasury}"
VALENCE_WITHDRAW_READY_ACCOUNT="${VALENCE_WITHDRAW_READY_ACCOUNT:-neutron1406thv6pxhzsk6l5femp6af3t53hxas7cwe92dph32d9lk7seuwq2mzhqh}"
VALENCE_PROVIDE_READY_ACCOUNT="${VALENCE_PROVIDE_READY_ACCOUNT:-neutron1kzhld870xq4yrkzhh837wcqwg6t9q74cscnwjhdv6wgsl0wv0n6qeual3s}"
USDC_DENOM="${USDC_DENOM:-ibc/B559A80D62249C8AA07A380E2A2BEA6E5CA9A6F079C912C3A9E9B494105E4F81}"
NTRN_USDC_PAIR_SHARE_DENOM="${NTRN_USDC_PAIR_SHARE_DENOM:-factory/neutron18c8qejysp4hgcfuxdpj4wf29mevzwllz5yh8uayjxamwtrs0n9fshq9vtv/astroport/share}"

LEGACY_ACCOUNTS=(
  "adminmodule"
  "harpoon"
  "revenue"
  "feeburner"
  "revenue-fee-redistribute"
  "revenue-staking-rewards"
)

q() {
  "$BINARY" "$@" --node "$NODE" -o json
}

query_or_empty_object() {
  if out="$(q "$@" 2>/dev/null)"; then
    printf '%s' "$out"
  else
    printf '{}'
  fi
}


module_address() {
  local module_name="$1"
  q query auth module-account "$module_name" 2>/dev/null | jq -r '.account.value.address // .account.base_account.address // empty' 2>/dev/null || true
}

try_wasm_query() {
  local contract="$1"
  local msg="$2"
  local out
  if out="$(q query wasm contract-state smart "$contract" "$msg" 2>/dev/null)"; then
    printf '%s' "$out"
    return 0
  fi
  return 1
}

roles_for_owner() {
  local owner="$1"
  local out=""
  local q1 q2
  q1="{\"get_roles\":{\"owner\":\"$owner\"}}"
  q2="{\"get_roles\":{\"signer\":\"$owner\"}}"
  if out="$(try_wasm_query "$IBCRL_CONTRACT" "$q1")"; then
    jq -c '.data // []' <<<"$out"
    return 0
  fi
  if out="$(try_wasm_query "$IBCRL_CONTRACT" "$q2")"; then
    jq -c '.data // []' <<<"$out"
    return 0
  fi
  printf '[]'
}

healthcheck() {
  local status_out network
  status_out="$("$BINARY" status --node "$NODE" -o json 2>&1)" || true
  network="$(jq -r '.node_info.network // empty' <<<"$status_out" 2>/dev/null || true)"
  if [[ -z "$network" ]]; then
    printf 'Error: healthcheck failed — could not reach node at %s\n' "$NODE" >&2
    exit 1
  fi
  printf '[healthcheck] node at %s is healthy (network: %s). Fetching state...\n' "$NODE" "$network" >&2
}

healthcheck

module_accounts_json='[]'
for acc in "${LEGACY_ACCOUNTS[@]}" "$REVENUE_MODULE_NAME"; do
  addr="$(module_address "$acc")"
  if [[ -n "$addr" ]]; then
    bal_json="$(query_or_empty_object query bank balances "$addr")"
    entry="$(jq -cn --arg name "$acc" --arg address "$addr" --argjson balances "$(jq -c '.balances // []' <<<"$bal_json")" \
      '{name:$name,address:$address,balances:$balances}')"
  else
    entry="$(jq -cn --arg name "$acc" '{name:$name,address:null,balances:[] }')"
  fi
  module_accounts_json="$(jq -c --argjson e "$entry" '. + [$e]' <<<"$module_accounts_json")"
done

revenue_addr="$(module_address "$REVENUE_MODULE_NAME")"

feemarket_params="$(query_or_empty_object query feemarket params)"
cron_params="$(query_or_empty_object query cron params)"
marketmap_params="$(query_or_empty_object query marketmap params)"
staking_params="$(query_or_empty_object query staking params)"
cron_schedules="$(query_or_empty_object query cron list-schedule)"

main_dao_bal="$(query_or_empty_object query bank balances "$MAIN_DAO_CONTRACT")"
staking_rewards_bal="$(query_or_empty_object query bank balances "$STAKING_REWARDS_CONTRACT")"
if [[ -n "$revenue_addr" ]]; then
  revenue_bal="$(query_or_empty_object query bank balances "$revenue_addr")"
else
  revenue_bal='{"balances":[]}'
fi

puppeteer_admin="$(query_or_empty_object query wasm contract "$PUPPETEER_CONTRACT" | jq -r '.contract_info.admin // empty')"
puppeteer_delegations="$(query_or_empty_object query staking delegations "$PUPPETEER_CONTRACT")"

multisig_roles_json="$(roles_for_owner "$IBCRL_MULTISIG")"
dao_roles_json="$(roles_for_owner "$MAIN_DAO_CONTRACT")"

total_power="$(q q wasm cs smart "$MAIN_DAO_CONTRACT" '{"total_power_at_height":{}}' | jq -r '.data.power')"
voting_registry="$(try_wasm_query "$MAIN_DAO_CONTRACT" '{"voting_module":{}}' | jq -r '.data')"
active_voting_vaults_count="$(try_wasm_query "$voting_registry" '{"voting_vaults":{}}' | jq -r '[.data[] | select(.state == "Active")] | length')"

valence_wra_bal="$(query_or_empty_object query bank balances "$VALENCE_WITHDRAW_READY_ACCOUNT")"
valence_pra_bal="$(query_or_empty_object query bank balances "$VALENCE_PROVIDE_READY_ACCOUNT")"

jq -n \
  --arg default_denom "$DEFAULT_DENOM" \
  --arg dntrn_denom "$DNTRN_DENOM" \
  --arg usdc_denom "$USDC_DENOM" \
  --arg ntrn_usdc_pair_share_denom "$NTRN_USDC_PAIR_SHARE_DENOM" \
  --arg main_dao "$MAIN_DAO_CONTRACT" \
  --arg staking_rewards "$STAKING_REWARDS_CONTRACT" \
  --arg puppeteer "$PUPPETEER_CONTRACT" \
  --arg ibcrl "$IBCRL_CONTRACT" \
  --arg ibcrl_multisig "$IBCRL_MULTISIG" \
  --arg revenue_module "$REVENUE_MODULE_NAME" \
  --arg revenue_addr "${revenue_addr:-}" \
  --arg valence_wra "$VALENCE_WITHDRAW_READY_ACCOUNT" \
  --arg valence_pra "$VALENCE_PROVIDE_READY_ACCOUNT" \
  --argjson module_accounts "$module_accounts_json" \
  --argjson feemarket_params "$feemarket_params" \
  --argjson cron_params "$cron_params" \
  --argjson marketmap_params "$marketmap_params" \
  --argjson staking_params "$staking_params" \
  --argjson cron_schedules "$cron_schedules" \
  --argjson main_dao_bal "$main_dao_bal" \
  --argjson staking_rewards_bal "$staking_rewards_bal" \
  --argjson revenue_bal "$revenue_bal" \
  --argjson valence_wra_bal "$valence_wra_bal" \
  --argjson valence_pra_bal "$valence_pra_bal" \
  --argjson multisig_roles "$multisig_roles_json" \
  --argjson dao_roles "$dao_roles_json" \
  --arg puppeteer_admin "${puppeteer_admin:-}" \
  --argjson puppeteer_delegations "$puppeteer_delegations" \
  --arg total_power "$total_power" \
  --arg active_voting_vaults_count "$active_voting_vaults_count" \
  '
{
  denoms: {
    dntrn: $dntrn_denom,
    usdc: $usdc_denom,
    ntrn_usdc_pair_share: $ntrn_usdc_pair_share_denom
  },
  addresses: {
    revenue_treasury: (if $revenue_addr == "" then null else $revenue_addr end),
    main_dao_contract: $main_dao,
    staking_rewards_contract: $staking_rewards,
    puppeteer_contract: $puppeteer,
    ibc_rate_limits_contract: $ibcrl,
    ibc_rate_limits_multisig: $ibcrl_multisig,
    valence_withdraw_ready_account: $valence_wra,
    valence_provide_ready_account: $valence_pra
  },
  module_params: {
    feemarket: $feemarket_params,
    cron: $cron_params,
    marketmap: $marketmap_params,
    staking: $staking_params
  },
  module_accounts: $module_accounts,
  balances: {
    revenue_treasury: {
      address: (if $revenue_addr == "" then null else $revenue_addr end),
      balances: ($revenue_bal.balances // [])
    },
    staking_rewards_contract: {
      address: $staking_rewards,
      balances: ($staking_rewards_bal.balances // [])
    },
    main_dao_contract: {
      address: $main_dao,
      ($default_denom): (($main_dao_bal.balances // []) | map(select(.denom == $default_denom)) | first | .amount // "0"),
      dntrn: (($main_dao_bal.balances // []) | map(select(.denom == $dntrn_denom)) | first | .amount // "0"),
      ntrn_usdc_pair_share: (($main_dao_bal.balances // []) | map(select(.denom == $ntrn_usdc_pair_share_denom)) | first | .amount // "0"),
      usdc: (($main_dao_bal.balances // []) | map(select(.denom == $usdc_denom)) | first | .amount // "0")
    },
    valence_withdraw_ready_account: {
      address: $valence_wra,
      ntrn_usdc_pair_share: (($valence_wra_bal.balances // []) | map(select(.denom == $ntrn_usdc_pair_share_denom)) | first | .amount // "0")
    },
    valence_provide_ready_account: {
      address: $valence_pra,
      usdc: (($valence_pra_bal.balances // []) | map(select(.denom == $usdc_denom)) | first | .amount // "0")
    }
  },
  dao_setup: {
    total_power: $total_power,
    active_voting_vaults_count: $active_voting_vaults_count
  },
  ibc_rate_limits: {
    contract: $ibcrl,
    roles_by_actor: {
      ($ibcrl_multisig): $multisig_roles,
      ($main_dao): $dao_roles
    }
  },
  cron_schedules: ($cron_schedules.schedules // []),
  puppeteer: {
    admin: (if $puppeteer_admin == "" then null else $puppeteer_admin end),
    delegations: (($puppeteer_delegations.delegation_responses // []) | map({key: .delegation.validator_address, value: .balance.amount}) | from_entries)
  }
}
' >"$OUT"

printf 'Wrote %s\n' "$OUT"
