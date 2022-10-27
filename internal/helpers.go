package internal

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcfeekeeper "github.com/cosmos/ibc-go/v4/modules/apps/29-fee/keeper"
	ibcfeetypes "github.com/cosmos/ibc-go/v4/modules/apps/29-fee/types"
)

func PayPacketFee(ctx sdk.Context, ibcfeeKeeper ibcfeekeeper.Keeper, payer, channelID, portID string, fee ibcfeetypes.Fee) error {
	goCtx := sdk.WrapSDKContext(ctx)

	payFeeMsg := ibcfeetypes.MsgPayPacketFee{
		Fee:             fee,
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
