package utils

import (
	"context"
	"fmt"

	"cosmossdk.io/math"

	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	types2 "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ccvconsumerkeeper "github.com/cosmos/interchain-security/v5/x/ccv/consumer/keeper"
	"github.com/cosmos/interchain-security/v5/x/ccv/consumer/types"
	"github.com/skip-mev/slinky/abci/ve"
	"github.com/skip-mev/slinky/pkg/math/voteweighted"
)

// Implement `ve.ValidatorStore` for `ConsumerValidatorStore` in order to pass in through to ValidateVoteExtensionsFn
var _ ve.ValidatorStore = (*ConsumerValidatorStore)(nil)

type ConsumerValidatorStore struct {
	k *ccvconsumerkeeper.Keeper
}

func NewConsumerValidatorStore(keeper *ccvconsumerkeeper.Keeper) ConsumerValidatorStore {
	return ConsumerValidatorStore{
		k: keeper,
	}
}

func (c ConsumerValidatorStore) GetPubKeyByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (cmtprotocrypto.PublicKey, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	val, found := c.k.GetCCValidator(sdkCtx, consAddr)
	if !found {
		// TODO
		return cmtprotocrypto.PublicKey{}, fmt.Errorf("not found CCValidator for address = TODO: ")
	}

	consPubKey, err := val.ConsPubKey()
	if err != nil {
		// TODO
		return cmtprotocrypto.PublicKey{}, fmt.Errorf("TODO")
	}
	tmPubKey, err := cryptocodec.ToCmtProtoPublicKey(consPubKey)
	if err != nil {
		// TODO
		return cmtprotocrypto.PublicKey{}, err
	}

	return tmPubKey, nil
}

// Implement `voteweighted.ValidatorStore` for `ConsumerValidatorStoreForAggregation` in order to pass in through to ValidateVoteExtensionsFn
var _ voteweighted.ValidatorStore = (*ConsumerValidatorStoreForAggregation)(nil)

type ConsumerValidatorStoreForAggregation struct {
	k *ccvconsumerkeeper.Keeper
}

func NewConsumerValidatorStoreForAggregation(keeper *ccvconsumerkeeper.Keeper) ConsumerValidatorStoreForAggregation {
	return ConsumerValidatorStoreForAggregation{
		k: keeper,
	}
}

func (c ConsumerValidatorStoreForAggregation) ValidatorByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (stakingtypes.ValidatorI, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	val, found := c.k.GetCCValidator(sdkCtx, consAddr)
	if !found {
		// TODO
		return nil, fmt.Errorf("not found CCValidator for address = TODO: ")
	}

	return ValidatorWithOnlyBondedTokens{v: &val}, nil
}

func (c ConsumerValidatorStoreForAggregation) TotalBondedTokens(ctx context.Context) (math.Int, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	validators := c.k.GetAllCCValidator(sdkCtx)
	totalPower := int64(0)
	for _, item := range validators {
		totalPower += item.Power
	}

	return math.NewInt(totalPower), nil
}

type ValidatorWithOnlyBondedTokens struct {
	v *types.CrossChainValidator
}

func (v ValidatorWithOnlyBondedTokens) IsJailed() bool {
	// TODO implement me
	panic("implement me")
}

func (v ValidatorWithOnlyBondedTokens) GetMoniker() string {
	// TODO implement me
	panic("implement me")
}

func (v ValidatorWithOnlyBondedTokens) GetStatus() stakingtypes.BondStatus {
	// TODO implement me
	panic("implement me")
}

func (v ValidatorWithOnlyBondedTokens) IsBonded() bool {
	// TODO implement me
	panic("implement me")
}

func (v ValidatorWithOnlyBondedTokens) IsUnbonded() bool {
	// TODO implement me
	panic("implement me")
}

func (v ValidatorWithOnlyBondedTokens) IsUnbonding() bool {
	// TODO implement me
	panic("implement me")
}

func (v ValidatorWithOnlyBondedTokens) GetOperator() string {
	// TODO implement me
	panic("implement me")
}

func (v ValidatorWithOnlyBondedTokens) ConsPubKey() (types2.PubKey, error) {
	// TODO implement me
	panic("implement me")
}

func (v ValidatorWithOnlyBondedTokens) TmConsPublicKey() (cmtprotocrypto.PublicKey, error) {
	// TODO implement me
	panic("implement me")
}

func (v ValidatorWithOnlyBondedTokens) GetConsAddr() ([]byte, error) {
	// TODO implement me
	panic("implement me")
}

func (v ValidatorWithOnlyBondedTokens) GetTokens() math.Int {
	// TODO implement me
	panic("implement me")
}

func (v ValidatorWithOnlyBondedTokens) GetBondedTokens() math.Int {
	return math.NewInt(v.v.Power)
}

func (v ValidatorWithOnlyBondedTokens) GetConsensusPower(_ math.Int) int64 {
	// TODO implement me
	panic("implement me")
}

func (v ValidatorWithOnlyBondedTokens) GetCommission() math.LegacyDec {
	// TODO implement me
	panic("implement me")
}

func (v ValidatorWithOnlyBondedTokens) GetMinSelfDelegation() math.Int {
	// TODO implement me
	panic("implement me")
}

func (v ValidatorWithOnlyBondedTokens) GetDelegatorShares() math.LegacyDec {
	// TODO implement me
	panic("implement me")
}

func (v ValidatorWithOnlyBondedTokens) TokensFromShares(_ math.LegacyDec) math.LegacyDec {
	// TODO implement me
	panic("implement me")
}

func (v ValidatorWithOnlyBondedTokens) TokensFromSharesTruncated(_ math.LegacyDec) math.LegacyDec {
	// TODO implement me
	panic("implement me")
}

func (v ValidatorWithOnlyBondedTokens) TokensFromSharesRoundUp(_ math.LegacyDec) math.LegacyDec {
	// TODO implement me
	panic("implement me")
}

func (v ValidatorWithOnlyBondedTokens) SharesFromTokens(_ math.Int) (math.LegacyDec, error) {
	// TODO implement me
	panic("implement me")
}

func (v ValidatorWithOnlyBondedTokens) SharesFromTokensTruncated(_ math.Int) (math.LegacyDec, error) {
	// TODO implement me
	panic("implement me")
}
