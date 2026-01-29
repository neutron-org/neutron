package utils

import (
	"testing"

	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"
)

func TestMustExtractDenomFromPacketOnRecv(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		denom    string
		response string
	}{
		{
			desc:     "native denom",
			denom:    "untrn",
			response: "ibc/0C698C8970DB4C539455E5225665A804F6338753211319E44BAD39758B238695",
		},
		{
			desc:     "ibc denom",
			denom:    "transfer/channel-0/kekw",
			response: "kekw",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			res := MustExtractDenomFromPacketOnRecv(makeMockPacket(tc.denom))
			require.Equal(t, res, tc.response)
		})
	}
}

func makeMockPacket(denom string) channeltypes.Packet {
	packetData := transfertypes.FungibleTokenPacketData{
		Denom:    denom,
		Amount:   "1",
		Sender:   "neutronsender",
		Receiver: "neutronreceiver",
		Memo:     "",
	}

	packet := channeltypes.NewPacket(
		packetData.GetBytes(),
		1,
		"transfer",
		"channel-0",
		"transfer",
		"channel-1",
		clienttypes.NewHeight(1, 150),
		0,
	)
	return packet
}
