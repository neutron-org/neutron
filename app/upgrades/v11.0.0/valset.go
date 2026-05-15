package v11_0_0

/*
{
  "operator_address": "neutronvaloper1pypmhaet69f57ymv526e78em6n9u9dshkx0npn",
  "moniker": "Stakecito"
}
{
  "operator_address": "neutronvaloper1pggzzg4wzsyxpcg9g57h5hkwumf3507yvcf4u6",
  "moniker": "SG-1"
}
{
  "operator_address": "neutronvaloper1rlyy2ltkc9t9s8gp2tmqxk6guggf6h9g6xj26y",
  "moniker": "01node"
}
{
  "operator_address": "neutronvaloper1xvl6sq77k6eq8kg9pyjyt8dzxzpyv9ukuuvsay",
  "moniker": "polkachu.com"
}
{
  "operator_address": "neutronvaloper12ysnwtjx5d87m0vffjerfu5vh2hm8qz6p4fppq",
  "moniker": "Stake&Relax 🦥"
}
{
  "operator_address": "neutronvaloper1v2azgtqlaj5xxktswpclsgn8zrr4960vhu50r2",
  "moniker": "PRO Delegators"
}
{
  "operator_address": "neutronvaloper104du2yqarz2uhnkakmytfwc58e79dftkmf9xz8",
  "moniker": "Hadron"
}
{
  "operator_address": "neutronvaloper1jkva4td5hfmmpdtjuxdk2yxg2q2pvslr287hg9",
  "moniker": "Smart Stake"
}
{
  "operator_address": "neutronvaloper15mfl7wxqww7zrvq2zw3d76mwsdxxff8cju5h5c",
  "moniker": "Crosnest"
}
{
  "operator_address": "neutronvaloper1kfqxeuqx2rxp0q6yh299737676r274398twxz8",
  "moniker": "🐹 Quokka Stake"
}
{
  "operator_address": "neutronvaloper1ksuxjvl70y9a60prfa4drqztsx8g7yuhvf36q6",
  "moniker": "Allnodes"
}
{
  "operator_address": "neutronvaloper1cvwrxye2g79ggstv403rn7tu952hs88udwclch",
  "moniker": "Newt Node"
}
{
  "operator_address": "neutronvaloper1c0rct7nkj4evl3j3s5sqzky4str95yr4fsg2mk",
  "moniker": "Golden Ratio Staking"
}
{
  "operator_address": "neutronvaloper1ekvdq09eczgjuv2yh86pesjkh6lt805dpw2qrw",
  "moniker": "Informal Systems"
}
{
  "operator_address": "neutronvaloper1e6rw7a7ngjmn8qjqfymu4en0px7hlr27fpf55k",
  "moniker": "Solva (CryptoCrew)"
}
{
  "operator_address": "neutronvaloper1md0k6m8y58w8u98x82kjah7r5zcajw7c5v5ypa",
  "moniker": "POSTHUMAN 🧬 StakeDrop"
}
{
  "operator_address": "neutronvaloper1uaz9tnrxlvad37t8d7vcevwd7z7xlp7lmu5pjv",
  "moniker": "P2P.org 💙"
}
{
  "operator_address": "neutronvaloper1ap2gshzfwglun4y2gpz6meugggat42s7vndhsw",
  "moniker": "Cosmostation"
}
*/

// NewValidatorSet is the target set of validators the DAO funds will be redelegated to.
// we keep all set except allnodes, informal systems, p2p.org and stakecito.
var NewValidatorSet = []string{
	"neutronvaloper1pggzzg4wzsyxpcg9g57h5hkwumf3507yvcf4u6", // sg-1
	"neutronvaloper1rlyy2ltkc9t9s8gp2tmqxk6guggf6h9g6xj26y", // 01node
	"neutronvaloper1xvl6sq77k6eq8kg9pyjyt8dzxzpyv9ukuuvsay", // polkachu.com
	"neutronvaloper12ysnwtjx5d87m0vffjerfu5vh2hm8qz6p4fppq", // stake&relax
	"neutronvaloper1v2azgtqlaj5xxktswpclsgn8zrr4960vhu50r2", // pro delegators
	"neutronvaloper1jkva4td5hfmmpdtjuxdk2yxg2q2pvslr287hg9", // smart stake
	"neutronvaloper15mfl7wxqww7zrvq2zw3d76mwsdxxff8cju5h5c", // crosnest
	"neutronvaloper1kfqxeuqx2rxp0q6yh299737676r274398twxz8", // 🐹 Quokka Stake
	"neutronvaloper1cvwrxye2g79ggstv403rn7tu952hs88udwclch", // newt node
	"neutronvaloper1c0rct7nkj4evl3j3s5sqzky4str95yr4fsg2mk", // golden ratio staking
	"neutronvaloper1e6rw7a7ngjmn8qjqfymu4en0px7hlr27fpf55k", // solva (cryptocrew)
	"neutronvaloper1md0k6m8y58w8u98x82kjah7r5zcajw7c5v5ypa", // posthuman (stakedrop)
	"neutronvaloper1ap2gshzfwglun4y2gpz6meugggat42s7vndhsw", // cosmostation
}
