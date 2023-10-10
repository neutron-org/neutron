package types_test

import (
	"testing"
	time "time"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/testutil/apptesting"
	dextypes "github.com/neutron-org/neutron/x/dex/types"
	. "github.com/neutron-org/neutron/x/incentives/types"
)

// TestMsgCreatePool tests if valid/invalid create pool messages are properly validated/invalidated
func TestMsgCreatePool(t *testing.T) {
	// generate a private/public key pair and get the respective address
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address())

	// make a proper createPool message
	createMsg := func(after func(msg MsgCreateGauge) MsgCreateGauge) MsgCreateGauge {
		distributeTo := QueryCondition{
			PairID: &dextypes.PairID{
				Token0: "TokenA",
				Token1: "TokenB",
			},
			StartTick: -10,
			EndTick:   10,
		}

		properMsg := *NewMsgCreateGauge(
			false,
			addr1,
			distributeTo,
			sdk.Coins{},
			time.Now(),
			2,
			0,
		)

		return after(properMsg)
	}

	// validate createPool message was created as intended
	msg := createMsg(func(msg MsgCreateGauge) MsgCreateGauge {
		return msg
	})
	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "create_gauge")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1.String())

	tests := []struct {
		name       string
		msg        MsgCreateGauge
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgCreateGauge) MsgCreateGauge {
				return msg
			}),
			expectPass: true,
		},
		{
			name: "empty owner",
			msg: createMsg(func(msg MsgCreateGauge) MsgCreateGauge {
				msg.Owner = ""
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid distribution start time",
			msg: createMsg(func(msg MsgCreateGauge) MsgCreateGauge {
				msg.StartTime = time.Time{}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid num epochs paid over",
			msg: createMsg(func(msg MsgCreateGauge) MsgCreateGauge {
				msg.NumEpochsPaidOver = 0
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid num epochs paid over for perpetual gauge",
			msg: createMsg(func(msg MsgCreateGauge) MsgCreateGauge {
				msg.NumEpochsPaidOver = 2
				msg.IsPerpetual = true
				return msg
			}),
			expectPass: false,
		},
		{
			name: "valid num epochs paid over for perpetual gauge",
			msg: createMsg(func(msg MsgCreateGauge) MsgCreateGauge {
				msg.NumEpochsPaidOver = 1
				msg.IsPerpetual = true
				return msg
			}),
			expectPass: true,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
		}
	}
}

// TestMsgAddToGauge tests if valid/invalid add to gauge messages are properly validated/invalidated
func TestMsgAddToGauge(t *testing.T) {
	// generate a private/public key pair and get the respective address
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address())

	// make a proper addToGauge message
	createMsg := func(after func(msg MsgAddToGauge) MsgAddToGauge) MsgAddToGauge {
		properMsg := *NewMsgAddToGauge(
			addr1,
			1,
			sdk.Coins{sdk.NewInt64Coin("stake", 10)},
		)

		return after(properMsg)
	}

	// validate addToGauge message was created as intended
	msg := createMsg(func(msg MsgAddToGauge) MsgAddToGauge {
		return msg
	})
	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "add_to_gauge")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1.String())

	tests := []struct {
		name       string
		msg        MsgAddToGauge
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgAddToGauge) MsgAddToGauge {
				return msg
			}),
			expectPass: true,
		},
		{
			name: "empty owner",
			msg: createMsg(func(msg MsgAddToGauge) MsgAddToGauge {
				msg.Owner = ""
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty rewards",
			msg: createMsg(func(msg MsgAddToGauge) MsgAddToGauge {
				msg.Rewards = sdk.Coins{}
				return msg
			}),
			expectPass: false,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
		}
	}
}

func TestMsgSetupStake(t *testing.T) {
	addr1, invalidAddr := apptesting.GenerateTestAddrs()

	tests := []struct {
		name       string
		msg        MsgStake
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: MsgStake{
				Owner: addr1,
				Coins: sdk.NewCoins(sdk.NewCoin("test", math.NewInt(100))),
			},
			expectPass: true,
		},
		{
			name: "invalid owner",
			msg: MsgStake{
				Owner: invalidAddr,
				Coins: sdk.NewCoins(sdk.NewCoin("test", math.NewInt(100))),
			},
		},
		{
			name: "zero token amount",
			msg: MsgStake{
				Owner: addr1,
				Coins: sdk.NewCoins(sdk.NewCoin("test", math.NewInt(0))),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.expectPass {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), RouterKey)
				require.Equal(t, test.msg.Type(), "stake_tokens")
				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), addr1)
			} else {
				require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
			}
		})
	}
}

func TestMsgUnstake(t *testing.T) {
	addr1, invalidAddr := apptesting.GenerateTestAddrs()

	tests := []struct {
		name       string
		msg        MsgUnstake
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: MsgUnstake{
				Owner: addr1,
				Unstakes: []*MsgUnstake_UnstakeDescriptor{
					{
						ID:    1,
						Coins: sdk.NewCoins(sdk.NewCoin("test", math.NewInt(100))),
					},
				},
			},
			expectPass: true,
		},
		{
			name: "invalid owner",
			msg: MsgUnstake{
				Owner: invalidAddr,
				Unstakes: []*MsgUnstake_UnstakeDescriptor{
					{
						ID:    1,
						Coins: sdk.NewCoins(sdk.NewCoin("test", math.NewInt(100))),
					},
				},
			},
		},
		{
			name: "invalid stake ID",
			msg: MsgUnstake{
				Owner: addr1,
				Unstakes: []*MsgUnstake_UnstakeDescriptor{
					{
						ID:    0,
						Coins: sdk.NewCoins(sdk.NewCoin("test", math.NewInt(100))),
					},
				},
			},
		},
		{
			name: "zero coins (same as nil)",
			msg: MsgUnstake{
				Owner: addr1,
				Unstakes: []*MsgUnstake_UnstakeDescriptor{
					{
						ID:    1,
						Coins: sdk.NewCoins(sdk.NewCoin("test1", math.NewInt(0))),
					},
				},
			},
			expectPass: true,
		},
		{
			name: "nil coins (unstake by ID)",
			msg: MsgUnstake{
				Owner: addr1,
				Unstakes: []*MsgUnstake_UnstakeDescriptor{
					{
						ID:    1,
						Coins: sdk.NewCoins(),
					},
				},
			},
			expectPass: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.expectPass {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), RouterKey)
				require.Equal(t, test.msg.Type(), "begin_unstaking")
				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), addr1)
			} else {
				require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
			}
		})
	}
}
