package internal

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcfeekeeper "github.com/cosmos/ibc-go/v4/modules/apps/29-fee/keeper"
	ibcfeetypes "github.com/cosmos/ibc-go/v4/modules/apps/29-fee/types"
)

func PayPacketFee(ctx sdk.Context, ibcfeeKeeper ibcfeekeeper.Keeper, payer, channelID, portID string) error {
	goCtx := sdk.WrapSDKContext(ctx)

	payFeeMsg := ibcfeetypes.MsgPayPacketFee{
		Fee: ibcfeetypes.Fee{
			RecvFee:    sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(10))),
			AckFee:     sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(10))),
			TimeoutFee: sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(10))),
		},
		SourcePortId:    portID,
		SourceChannelId: channelID,
		Signer:          payer,
		Relayers:        nil,
	}

	if _, err := ibcfeeKeeper.PayPacketFee(goCtx, &payFeeMsg); err != nil {
		return err
	}

	return nil
}
