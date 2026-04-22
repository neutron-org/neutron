#!/usr/bin/env bash
set -euo pipefail

BINARY="${BINARY:-neutrond}"
NODE="${NODE:-http://localhost:26657}"
OUT="${OUT:-post_mig_state.json}"
PRE_FILE="${PRE_FILE:-pre_mig_state.json}"

DEFAULT_DENOM="untrn"
DNTRN_DENOM="${DNTRN_DENOM:-factory/neutron1frc0p5czd9uaaymdkug2njz7dc7j65jxukp9apmt9260a8egujkspms2t2/udntrn}"
ASTROPORT_SHARE_DENOM="${ASTROPORT_SHARE_DENOM:-factory/neutron1pd9u7h4vf36vtj5lqlcp4376xf4wktdnhmzqtn8958wyh0nzwsmsavc2dz/astroport/share}"

MAIN_DAO_CONTRACT="${MAIN_DAO_CONTRACT:-neutron1suhgf5svhu4usrurvxzlgn54ksxmn8gljarjtxqnapv8kjnp4nrstdxvff}"
PUPPETEER_CONTRACT="${PUPPETEER_CONTRACT:-neutron17jsl4t4hhaw37tnhenskrfntm7mv44wzjr3f990hx4p9r5m0gzdqquhtd3}"
STAKING_REWARDS_CONTRACT="${STAKING_REWARDS_CONTRACT:-neutron1gqq3c735pj6ese3yru5xr6ud0fvxgltxesygvyyzpsrt74v6yg4sgkrgwq}"
IBCRL_CONTRACT="${IBCRL_CONTRACT:-neutron15aqgplxcavqhurr0g5wwtdw6025dknkqwkfh0n46gp2qjl6236cs2yd3nl}"
IBCRL_MULTISIG="${IBCRL_MULTISIG:-neutron1el2rymcsg5wxth2fz2g5l08nue3xhyj3ny5wea3yxwr9f7es8d6smmwrck}"

if [[ ! -f "$PRE_FILE" ]]; then
  printf 'Error: pre-migration state file not found: %s\n' "$PRE_FILE" >&2
  exit 1
fi

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

# TODO: healthcheck to node

# Read module account addresses from pre-migration state, query their post-migration balances.
module_accounts_json='[]'
while IFS= read -r entry; do
  name="$(jq -r '.name' <<<"$entry")"
  addr="$(jq -r '.address // empty' <<<"$entry")"
  if [[ -n "$addr" ]]; then
    bal_json="$(query_or_empty_object query bank balances "$addr")"
    new_entry="$(jq -cn --arg name "$name" --arg address "$addr" \
      --argjson balances "$(jq -c '.balances // []' <<<"$bal_json")" \
      '{name:$name,address:$address,balances:$balances}')"
  else
    new_entry="$(jq -cn --arg name "$name" '{name:$name,address:null,balances:[]}')"
  fi
  module_accounts_json="$(jq -c --argjson e "$new_entry" '. + [$e]' <<<"$module_accounts_json")"
done < <(jq -c '.module_accounts[]' "$PRE_FILE")

gov_addr="$(module_address gov)"
revenue_addr="$(jq -r '.addresses.revenue_treasury // empty' "$PRE_FILE")"

gov_params="$(query_or_empty_object query gov params)"
mint_params="$(query_or_empty_object query mint params)"
distribution_params="$(query_or_empty_object query distribution params)"
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
puppeteer_unbonding="$(query_or_empty_object query staking unbonding-delegations "$PUPPETEER_CONTRACT")"

gov_roles_json="$(roles_for_owner "$gov_addr")"
multisig_roles_json="$(roles_for_owner "$IBCRL_MULTISIG")"
dao_roles_json="$(roles_for_owner "$MAIN_DAO_CONTRACT")"

jq -n \
  --arg default_denom "$DEFAULT_DENOM" \
  --arg dntrn_denom "$DNTRN_DENOM" \
  --arg astro_share_denom "$ASTROPORT_SHARE_DENOM" \
  --arg main_dao "$MAIN_DAO_CONTRACT" \
  --arg staking_rewards "$STAKING_REWARDS_CONTRACT" \
  --arg puppeteer "$PUPPETEER_CONTRACT" \
  --arg ibcrl "$IBCRL_CONTRACT" \
  --arg ibcrl_multisig "$IBCRL_MULTISIG" \
  --arg gov_addr "$gov_addr" \
  --arg revenue_addr "${revenue_addr:-}" \
  --argjson module_accounts "$module_accounts_json" \
  --argjson gov_params "$gov_params" \
  --argjson mint_params "$mint_params" \
  --argjson distribution_params "$distribution_params" \
  --argjson feemarket_params "$feemarket_params" \
  --argjson cron_params "$cron_params" \
  --argjson marketmap_params "$marketmap_params" \
  --argjson staking_params "$staking_params" \
  --argjson cron_schedules "$cron_schedules" \
  --argjson main_dao_bal "$main_dao_bal" \
  --argjson staking_rewards_bal "$staking_rewards_bal" \
  --argjson revenue_bal "$revenue_bal" \
  --arg puppeteer_admin "${puppeteer_admin:-}" \
  --argjson puppeteer_delegations "$puppeteer_delegations" \
  --argjson puppeteer_unbonding "$puppeteer_unbonding" \
  --argjson gov_roles "$gov_roles_json" \
  --argjson multisig_roles "$multisig_roles_json" \
  --argjson dao_roles "$dao_roles_json" \
  '
{
  denoms: {
    dntrn: $dntrn_denom,
    astro_share: $astro_share_denom
  },
  addresses: {
    gov_module: (if $gov_addr == "" then null else $gov_addr end),
    revenue_treasury: (if $revenue_addr == "" then null else $revenue_addr end),
    main_dao_contract: $main_dao,
    staking_rewards_contract: $staking_rewards,
    puppeteer_contract: $puppeteer,
    ibc_rate_limits_contract: $ibcrl,
    ibc_rate_limits_multisig: $ibcrl_multisig
  },
  module_params: {
    gov: $gov_params,
    mint: $mint_params,
    distribution: $distribution_params,
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
      astroport_share: (($main_dao_bal.balances // []) | map(select(.denom == $astro_share_denom)) | first | .amount // "0")
    }
  },
  ibc_rate_limits: {
    contract: $ibcrl,
    roles_by_actor: {
      ($gov_addr): $gov_roles,
      ($ibcrl_multisig): $multisig_roles,
      ($main_dao): $dao_roles
    }
  },
  cron_schedules: ($cron_schedules.schedules // []),
  puppeteer: {
    admin: (if $puppeteer_admin == "" then null else $puppeteer_admin end),
    delegations: (($puppeteer_delegations.delegation_responses // []) | map({key: .delegation.validator_address, value: .balance.amount}) | from_entries),
    unbonding_delegations: (($puppeteer_unbonding.unbonding_responses // []) | map({key: .validator_address, value: (.entries | map(.balance | tonumber) | add // 0 | tostring)}) | from_entries)
  }
}
' >"$OUT"

printf 'Wrote %s\n' "$OUT"
