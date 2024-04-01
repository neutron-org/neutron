package utils

import (
	"context"
	"fmt"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ccvconsumerkeeper "github.com/cosmos/interchain-security/v5/x/ccv/consumer/keeper"
)

//var _ ConsumerValidatorStore = (*ve.ValidatorStore)(nil)

type ConsumerValidatorStore struct {
	k *ccvconsumerkeeper.Keeper
}

func (c ConsumerValidatorStore) GetPubKeyByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (cmtprotocrypto.PublicKey, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	val, found := c.k.GetCCValidator(sdkCtx, consAddr)
	if !found {
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

func NewConsumerValidatorStore(keeper *ccvconsumerkeeper.Keeper) ConsumerValidatorStore {
	return ConsumerValidatorStore{
		k: keeper,
	}
}
