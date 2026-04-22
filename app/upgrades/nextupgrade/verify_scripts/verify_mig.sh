#!/usr/bin/env bash
set -uo pipefail

PRE_FILE="${PRE_FILE:-pre_mig_state.json}"
POST_FILE="${POST_FILE:-post_mig_state.json}"

PASS_COUNT=0
FAIL_COUNT=0

GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
DIM='\033[2m'
RESET='\033[0m'

ok()   { printf "${GREEN}✓${RESET} %-60s = %s\n" "$1" "$2"; PASS_COUNT=$((PASS_COUNT + 1)); }
fail() { printf "${RED}✗${RESET} %-60s   expected: %s  actual: %s\n" "$1" "$2" "$3" >&2; FAIL_COUNT=$((FAIL_COUNT + 1)); }
info() { printf "  ${DIM}·${RESET} %-58s  %s\n" "$1" "$2"; }
note() { printf "  ${CYAN}%s${RESET}\n" "$1"; }

assert_eq() {
  local name="$1" expected="$2" actual="$3"
  if [[ "$actual" == "$expected" ]]; then ok "$name" "$actual"
  else fail "$name" "$expected" "$actual"
  fi
}

# For structural pre-vs-post comparisons: shows what matched (or what differed).
assert_unchanged() {
  local name="$1" pre_val="$2" post_val="$3"
  if [[ "$pre_val" == "$post_val" ]]; then
    ok "$name" "unchanged from pre-migration"
  else
    printf "${RED}✗${RESET} %-60s\n" "$name" >&2
    printf "  pre:  %s\n" "$pre_val" >&2
    printf "  post: %s\n" "$post_val" >&2
    FAIL_COUNT=$((FAIL_COUNT + 1))
  fi
}

rjq() { jq -r "$2" "$1"; }
pre()  { rjq "$PRE_FILE"  "$1"; }
post() { rjq "$POST_FILE" "$1"; }

check_file() {
  local file="$1"
  if [[ ! -f "$file" ]]; then printf "${RED}ERROR${RESET}: Missing file: %s\n" "$file" >&2; exit 1; fi
  if ! jq empty "$file" 2>/dev/null; then printf "${RED}ERROR${RESET}: Invalid JSON: %s\n" "$file" >&2; exit 1; fi
}

check_file "$PRE_FILE"
check_file "$POST_FILE"

gov_addr="$(post '.addresses.gov_module')"
multisig_addr="$(post '.addresses.ibc_rate_limits_multisig')"
main_dao_addr="$(post '.addresses.main_dao_contract')"
default_denom="untrn"

printf '\n=== Migration verification ===\n'
printf 'Pre:  %s\n' "$PRE_FILE"
printf 'Post: %s\n' "$POST_FILE"

NEW_VALIDATOR_SET=(
  "neutronvaloper1pggzzg4wzsyxpcg9g57h5hkwumf3507yvcf4u6"  # sg-1
  "neutronvaloper1rlyy2ltkc9t9s8gp2tmqxk6guggf6h9g6xj26y"  # 01node
  "neutronvaloper1xvl6sq77k6eq8kg9pyjyt8dzxzpyv9ukuuvsay"  # polkachu.com
  "neutronvaloper12ysnwtjx5d87m0vffjerfu5vh2hm8qz6p4fppq"  # stake&relax
  "neutronvaloper1v2azgtqlaj5xxktswpclsgn8zrr4960vhu50r2"  # pro delegators
  "neutronvaloper104du2yqarz2uhnkakmytfwc58e79dftkmf9xz8"  # hadron
  "neutronvaloper1jkva4td5hfmmpdtjuxdk2yxg2q2pvslr287hg9"  # smart stake
  "neutronvaloper15mfl7wxqww7zrvq2zw3d76mwsdxxff8cju5h5c"  # crosnest
  "neutronvaloper1kfqxeuqx2rxp0q6yh299737676r274398twxz8"  # quokka stake
  "neutronvaloper1cvwrxye2g79ggstv403rn7tu952hs88udwclch"  # newt node
  "neutronvaloper1c0rct7nkj4evl3j3s5sqzky4str95yr4fsg2mk"  # golden ratio staking
  "neutronvaloper1e6rw7a7ngjmn8qjqfymu4en0px7hlr27fpf55k"  # solva (cryptocrew)
  "neutronvaloper1md0k6m8y58w8u98x82kjah7r5zcajw7c5v5ypa"  # posthuman (stakedrop)
  "neutronvaloper1ap2gshzfwglun4y2gpz6meugggat42s7vndhsw"  # cosmostation
)
UNBONDING_PER_VALIDATOR="500000000"
NEW_VALIDATOR_COUNT=14
UNBONDING_TOTAL=$((UNBONDING_PER_VALIDATOR * NEW_VALIDATOR_COUNT))

# ─────────────────────────────────────────────────────────────────────────────
printf '\n--- 1. gov / mint / distribution params ---\n'
note "Freshly initialized by the upgrade — gov, mint and distribution did not exist pre-migration."

gov="$(post '.module_params.gov.params')"
assert_eq "gov.min_deposit[0].denom"            "$default_denom"           "$(jq -r '.min_deposit[0].denom'             <<<"$gov")"
assert_eq "gov.min_deposit[0].amount"           "1000000000000"            "$(jq -r '.min_deposit[0].amount'            <<<"$gov")"
assert_eq "gov.expedited_min_deposit[0].amount" "3000000000000"            "$(jq -r '.expedited_min_deposit[0].amount'  <<<"$gov")"
assert_eq "gov.max_deposit_period"              "168h0m0s"                 "$(jq -r '.max_deposit_period'               <<<"$gov")"
assert_eq "gov.voting_period"                   "336h0m0s"                 "$(jq -r '.voting_period'                    <<<"$gov")"
assert_eq "gov.expedited_voting_period"         "72h0m0s"                  "$(jq -r '.expedited_voting_period'          <<<"$gov")"
assert_eq "gov.quorum"                          "0.450000000000000000"     "$(jq -r '.quorum'                           <<<"$gov")"
assert_eq "gov.threshold"                       "0.500000000000000000"     "$(jq -r '.threshold'                        <<<"$gov")"
assert_eq "gov.expedited_threshold"             "0.670000000000000000"     "$(jq -r '.expedited_threshold'              <<<"$gov")"
assert_eq "gov.veto_threshold"                  "0.330000000000000000"     "$(jq -r '.veto_threshold'                   <<<"$gov")"
assert_eq "gov.min_initial_deposit_ratio"       "1.000000000000000000"     "$(jq -r '.min_initial_deposit_ratio'        <<<"$gov")"
assert_eq "gov.proposal_cancel_ratio"           "0.000000000000000000"     "$(jq -r '.proposal_cancel_ratio'            <<<"$gov")"
assert_eq "gov.burn_vote_veto"                  "true"                     "$(jq -r '.burn_vote_veto'                   <<<"$gov")"
assert_eq "gov.min_deposit_ratio"               "0.100000000000000000"     "$(jq -r '.min_deposit_ratio'                <<<"$gov")"

mint="$(post '.module_params.mint.params')"
assert_eq "mint.mint_denom"            "$default_denom"          "$(jq -r '.mint_denom'            <<<"$mint")"
assert_eq "mint.inflation_max"         "0.300000000000000000"    "$(jq -r '.inflation_max'         <<<"$mint")"
assert_eq "mint.inflation_min"         "0.010000000000000000"    "$(jq -r '.inflation_min'         <<<"$mint")"
assert_eq "mint.inflation_rate_change" "0.200000000000000000"    "$(jq -r '.inflation_rate_change' <<<"$mint")"
assert_eq "mint.goal_bonded"           "0.670000000000000000"    "$(jq -r '.goal_bonded'           <<<"$mint")"

distr="$(post '.module_params.distribution.params')"
assert_eq "distribution.community_tax" "0.000000000000000000" "$(jq -r '.community_tax' <<<"$distr")"

# ─────────────────────────────────────────────────────────────────────────────
printf '\n--- 2. IBC rate limiter roles ---\n'
note "Gov module granted all roles; multisig and main DAO fully revoked."
info "gov module address" "$gov_addr"
info "multisig address"   "$multisig_addr"
info "main DAO address"   "$main_dao_addr"

all_roles_sorted='["AddRateLimit","EditPathQuota","GrantRole","ManageDenomRestrictions","RemoveMessage","RemoveRateLimit","ResetPathQuota","RevokeRole","SetTimelockDelay"]'

assert_eq "ibcrl.gov.roles" \
  "$all_roles_sorted" \
  "$(jq -cS --arg a "$gov_addr"      '.ibc_rate_limits.roles_by_actor[$a] // [] | sort' "$POST_FILE")"
assert_eq "ibcrl.multisig.roles (revoked)" \
  "[]" \
  "$(jq -cS --arg a "$multisig_addr" '.ibc_rate_limits.roles_by_actor[$a] // [] | sort' "$POST_FILE")"
assert_eq "ibcrl.main_dao.roles (revoked)" \
  "[]" \
  "$(jq -cS --arg a "$main_dao_addr" '.ibc_rate_limits.roles_by_actor[$a] // [] | sort' "$POST_FILE")"

# ─────────────────────────────────────────────────────────────────────────────
printf '\n--- 3. feemarket params ---\n'
note "send_tip_to_proposer should flip false→true; all other params should carry over unchanged."

assert_eq "feemarket.send_tip_to_proposer (pre, before upgrade)" \
  "false" "$(pre '.module_params.feemarket.send_tip_to_proposer')"
assert_eq "feemarket.send_tip_to_proposer (post, flipped by upgrade)" \
  "true"  "$(post '.module_params.feemarket.send_tip_to_proposer')"
assert_unchanged "feemarket.all_other_params" \
  "$(jq -cS 'del(.send_tip_to_proposer)' <<<"$(pre  '.module_params.feemarket')")" \
  "$(jq -cS 'del(.send_tip_to_proposer)' <<<"$(post '.module_params.feemarket')")"

# ─────────────────────────────────────────────────────────────────────────────
printf '\n--- 4. Burns: revenue treasury and staking rewards ---\n'
note "Upgrade burns all untrn from revenue-treasury and staking-rewards contract."

rev_pre="$(jq -r --arg d "$default_denom" \
  '.balances.revenue_treasury.balances | map(select(.denom==$d)) | first | .amount // "0"' "$PRE_FILE")"
rev_post="$(jq -r --arg d "$default_denom" \
  '.balances.revenue_treasury.balances | map(select(.denom==$d)) | first | .amount // "0"' "$POST_FILE")"
info "revenue_treasury.$default_denom pre-migration balance" "$rev_pre"
assert_eq "revenue_treasury.$default_denom (burned to zero)" "0" "$rev_post"

sr_pre="$(jq -r --arg d "$default_denom" \
  '.balances.staking_rewards_contract.balances | map(select(.denom==$d)) | first | .amount // "0"' "$PRE_FILE")"
sr_post="$(jq -r --arg d "$default_denom" \
  '.balances.staking_rewards_contract.balances | map(select(.denom==$d)) | first | .amount // "0"' "$POST_FILE")"
info "staking_rewards_contract.$default_denom pre-migration balance" "$sr_pre"
assert_eq "staking_rewards_contract.$default_denom (burned to zero)" "0" "$sr_post"

# ─────────────────────────────────────────────────────────────────────────────
printf '\n--- 5. Burns: DAO dNTRN and Astroport LP shares ---\n'
note "dNTRN burned directly; LP shares withdrawn, converted to NTRN via Drop swap, then burned."

dntrn_pre="$(pre  '.balances.main_dao_contract.dntrn')"
dntrn_post="$(post '.balances.main_dao_contract.dntrn')"
info "main_dao.dntrn pre-migration balance" "$dntrn_pre"
assert_eq "main_dao.dntrn (burned to zero)" "0" "$dntrn_post"

astro_post="$(post '.balances.main_dao_contract.astroport_share')"
assert_eq "main_dao.astroport_share (withdrawn and burned)" "0" "$astro_post"

# ─────────────────────────────────────────────────────────────────────────────
printf '\n--- 6. Cron params ---\n'
note "security_address should change to gov module; all other cron params should be unchanged."

pre_security="$(pre '.module_params.cron.params.security_address')"
info "cron.security_address pre-migration" "$pre_security"
assert_eq "cron.security_address (set to gov module)" "$gov_addr" \
  "$(post '.module_params.cron.params.security_address')"
assert_unchanged "cron.all_other_params" \
  "$(jq -cS 'del(.security_address)' <<<"$(pre  '.module_params.cron.params')")" \
  "$(jq -cS 'del(.security_address)' <<<"$(post '.module_params.cron.params')")"

# ─────────────────────────────────────────────────────────────────────────────
printf '\n--- 7. MarketMap params ---\n'
note "admin and market_authorities should both be set to the gov module address."

pre_admin="$(pre '.module_params.marketmap.admin // .module_params.marketmap.params.admin')"
info "marketmap.admin pre-migration" "$pre_admin"
assert_eq "marketmap.admin (set to gov module)" "$gov_addr" \
  "$(post '.module_params.marketmap.admin // .module_params.marketmap.params.admin')"
assert_eq "marketmap.market_authorities count" "1" \
  "$(jq -r '(.module_params.marketmap.market_authorities // .module_params.marketmap.params.market_authorities) | length' "$POST_FILE")"
assert_eq "marketmap.market_authorities[0] (set to gov module)" "$gov_addr" \
  "$(jq -r '(.module_params.marketmap.market_authorities // .module_params.marketmap.params.market_authorities)[0]' "$POST_FILE")"

# ─────────────────────────────────────────────────────────────────────────────
printf '\n--- 8. Module accounts drained ---\n'
note "Upgrade transfers all balances from legacy module accounts to main DAO."

while IFS= read -r entry; do
  name="$(jq -r '.name' <<<"$entry")"
  addr="$(jq -r '.address // empty' <<<"$entry")"
  if [[ -z "$addr" ]]; then
    info "$name" "no on-chain address pre-migration, skipping"
    continue
  fi
  pre_bal_count="$(jq -r --arg n "$name" '(.module_accounts[] | select(.name==$n) | .balances | length) // 0' "$PRE_FILE")"
  post_bal_count="$(jq -r --arg n "$name" '(.module_accounts[] | select(.name==$n) | .balances | length) // 0' "$POST_FILE")"
  info "$name pre-migration balance entries" "$pre_bal_count"
  assert_eq "module_account.$name (drained, 0 balance entries)" "0" "$post_bal_count"
done < <(jq -c '.module_accounts[]' "$PRE_FILE")

# ─────────────────────────────────────────────────────────────────────────────
printf '\n--- 9. Staking: max_validators and new validator set ---\n'
note "max_validators updated 18→14; all other staking params unchanged; delegations only to 14 new validators."

pre_max="$(pre '.module_params.staking.params.max_validators | tostring')"
info "staking.max_validators pre-migration" "$pre_max"
assert_eq "staking.max_validators (updated to 14)" "14" \
  "$(post '.module_params.staking.params.max_validators | tostring')"
assert_unchanged "staking.all_other_params" \
  "$(jq -cS 'del(.max_validators)' <<<"$(pre  '.module_params.staking.params')")" \
  "$(jq -cS 'del(.max_validators)' <<<"$(post '.module_params.staking.params')")"

printf '\n'
note "Each validator in the new set must have a delegation from the puppeteer:"
for val in "${NEW_VALIDATOR_SET[@]}"; do
  amt="$(jq -r --arg v "$val" '.puppeteer.delegations[$v] // "missing"' "$POST_FILE")"
  assert_eq "delegation to $val" "true" \
    "$(jq -r --arg v "$val" '.puppeteer.delegations | has($v) | tostring' "$POST_FILE")"
  info "  amount" "$amt"
done

printf '\n'
note "No delegations outside the new validator set:"
while IFS= read -r val; do
  found="false"
  for expected in "${NEW_VALIDATOR_SET[@]}"; do
    [[ "$val" == "$expected" ]] && found="true" && break
  done
  assert_eq "delegation validator $val in new set" "true" "$found"
done < <(jq -r '.puppeteer.delegations | keys[]' "$POST_FILE")

# ─────────────────────────────────────────────────────────────────────────────
printf '\n--- 10. Delegation sum: pre = post + 14×500000000 ---\n'
note "14 validators each have 500M untrn being unbonded; pre total should equal post total + those unbondings."

pre_sum="$(jq -r '[.puppeteer.delegations | to_entries[] | .value | tonumber] | add // 0' "$PRE_FILE")"
post_sum="$(jq -r '[.puppeteer.delegations | to_entries[] | .value | tonumber] | add // 0' "$POST_FILE")"
expected_post_sum=$(( pre_sum - UNBONDING_TOTAL ))

info "delegations sum pre-migration"       "$pre_sum"
info "delegations sum post-migration"      "$post_sum"
info "unbonding total (14 × 500000000)"    "$UNBONDING_TOTAL"
info "expected post sum (pre − unbonding)" "$expected_post_sum"
assert_eq "delegations.post_sum = pre_sum − unbonding_total" "$expected_post_sum" "$post_sum"

# ─────────────────────────────────────────────────────────────────────────────
printf '\n--- 11. Cron schedule: undelegations manager tick & burn ---\n'
note "New schedule registered by the upgrade: calls tick then burn on the undelegations manager contract."
SCHEDULE_NAME="undelegations manager contract tick & burn"

assert_eq "schedule \"$SCHEDULE_NAME\" exists" "true" \
  "$(jq -r --arg n "$SCHEDULE_NAME" '[.cron_schedules[] | select(.name==$n)] | length > 0 | tostring' "$POST_FILE")"
assert_eq "schedule message count" "2" \
  "$(jq -r --arg n "$SCHEDULE_NAME" '[.cron_schedules[] | select(.name==$n)] | first | .msgs | length' "$POST_FILE")"
assert_eq "schedule msgs[0].msg" '{"tick": {}}' \
  "$(jq -r --arg n "$SCHEDULE_NAME" '[.cron_schedules[] | select(.name==$n)] | first | .msgs[0].msg' "$POST_FILE")"
assert_eq "schedule msgs[1].msg" '{"burn": {}}' \
  "$(jq -r --arg n "$SCHEDULE_NAME" '[.cron_schedules[] | select(.name==$n)] | first | .msgs[1].msg' "$POST_FILE")"

# ─────────────────────────────────────────────────────────────────────────────
printf '\n--- 12. Puppeteer unbonding delegations ---\n'
note "14 unbonding delegations of 500000000 untrn each — one per new validator, created during redelegation."

assert_eq "puppeteer unbonding delegation count" "14" \
  "$(jq -r '.puppeteer.unbonding_delegations | length' "$POST_FILE")"

while IFS= read -r entry; do
  val="$(jq -r '.key'   <<<"$entry")"
  amt="$(jq -r '.value' <<<"$entry")"
  assert_eq "unbonding $val amount" "$UNBONDING_PER_VALIDATOR" "$amt"
done < <(jq -c '.puppeteer.unbonding_delegations | to_entries[]' "$POST_FILE")

# ─────────────────────────────────────────────────────────────────────────────
printf '\n--- 13. Puppeteer admin ---\n'
note "Puppeteer contract admin should be transferred to the gov module by the upgrade."

pre_puppeteer_admin="$(pre '.puppeteer.admin')"
info "puppeteer admin pre-migration" "$pre_puppeteer_admin"
assert_eq "puppeteer.admin (set to gov module)" "$gov_addr" "$(post '.puppeteer.admin')"

# ─────────────────────────────────────────────────────────────────────────────
printf '\n'
if [[ "$FAIL_COUNT" -gt 0 ]]; then
  printf "${RED}=== Summary: FAILED — ✓ %d  ✗ %d ===${RESET}\n" "$PASS_COUNT" "$FAIL_COUNT" >&2
  exit 1
fi
printf "${GREEN}=== Summary: PASSED — ✓ %d  ✗ %d ===${RESET}\n" "$PASS_COUNT" "$FAIL_COUNT"
