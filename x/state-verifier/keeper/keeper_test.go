package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	wasmKeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/proto/tendermint/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/v10/modules/core/23-commitment/types"
	host "github.com/cosmos/ibc-go/v10/modules/core/24-host"
	ibchost "github.com/cosmos/ibc-go/v10/modules/core/exported"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v6/app/params"
	"github.com/neutron-org/neutron/v6/testutil"
	iqtypes "github.com/neutron-org/neutron/v6/x/interchainqueries/types"
)

var reflectContractPath = "../../../wasmbinding/testdata/reflect.wasm"

type KeeperTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func (suite *KeeperTestSuite) TestVerifyValue() {
	tests := []struct {
		name     string
		malleate func(sender string, ctx sdk.Context) ([]*iqtypes.StorageValue, int64, error)
	}{
		{
			name: "valid KV storage proof",
			malleate: func(_ string, ctx sdk.Context) ([]*iqtypes.StorageValue, int64, error) {
				clientKey := host.FullClientStateKey(suite.Path.EndpointA.ClientID)

				resp, err := suite.ChainA.App.Query(ctx, &types.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", ibchost.StoreKey),
					Height: suite.ChainA.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})
				suite.Require().NoError(err)

				return []*iqtypes.StorageValue{{
					Key:           resp.Key,
					Proof:         resp.ProofOps,
					Value:         resp.Value,
					StoragePrefix: ibchost.StoreKey,
				}}, resp.Height, nil
			},
		},
		{
			name: "empty KV storage proof",
			malleate: func(_ string, ctx sdk.Context) ([]*iqtypes.StorageValue, int64, error) {
				clientKey := host.FullClientStateKey(suite.Path.EndpointA.ClientID)

				resp, err := suite.ChainA.App.Query(ctx, &types.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", ibchost.StoreKey),
					Height: suite.ChainA.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})
				suite.Require().NoError(err)

				return []*iqtypes.StorageValue{{
					Key:           resp.Key,
					Proof:         nil,
					Value:         resp.Value,
					StoragePrefix: ibchost.StoreKey,
				}}, resp.Height, ibctypes.ErrInvalidMerkleProof
			},
		},
		{
			name: "invalid KV storage proof",
			malleate: func(_ string, ctx sdk.Context) ([]*iqtypes.StorageValue, int64, error) {
				clientKey := host.FullClientStateKey(suite.Path.EndpointA.ClientID)

				resp, err := suite.ChainA.App.Query(ctx, &types.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", ibchost.StoreKey),
					Height: suite.ChainA.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})
				suite.Require().NoError(err)

				return []*iqtypes.StorageValue{{
					Key:           resp.Key,
					Proof:         &crypto.ProofOps{Ops: []crypto.ProofOp{{Type: "dasfsdf", Key: []byte("sffgsdf"), Data: []byte("sfdsdfs")}}},
					Value:         resp.Value,
					StoragePrefix: ibchost.StoreKey,
				}}, resp.Height, ibctypes.ErrInvalidMerkleProof
			},
		},
		{
			name: "invalid height for proof",
			malleate: func(_ string, ctx sdk.Context) ([]*iqtypes.StorageValue, int64, error) {
				clientKey := host.FullClientStateKey(suite.Path.EndpointA.ClientID)

				resp, err := suite.ChainA.App.Query(ctx, &types.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", ibchost.StoreKey),
					Height: suite.ChainA.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})
				suite.Require().NoError(err)

				return []*iqtypes.StorageValue{{
					Key:           resp.Key,
					Proof:         resp.ProofOps,
					Value:         resp.Value,
					StoragePrefix: ibchost.StoreKey,
				}}, resp.Height - 2, fmt.Errorf("Please ensure proof was submitted with correct proofHeight and to the correct chain.") //nolint:revive
			},
		},
		{
			name: "invalid storage prefix",
			malleate: func(_ string, ctx sdk.Context) ([]*iqtypes.StorageValue, int64, error) {
				clientKey := host.FullClientStateKey(suite.Path.EndpointA.ClientID)

				resp, err := suite.ChainA.App.Query(ctx, &types.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", ibchost.StoreKey),
					Height: suite.ChainA.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})
				suite.Require().NoError(err)

				return []*iqtypes.StorageValue{{
					Key:           resp.Key,
					Proof:         resp.ProofOps,
					Value:         resp.Value,
					StoragePrefix: "kekekek",
				}}, resp.Height, fmt.Errorf("Please ensure the path and value are both correct.") //nolint:revive
			},
		},
	}

	for i, tc := range tests {
		tt := tc
		suite.Run(fmt.Sprintf("Case %s, %d/%d tests", tt.name, i+1, len(tests)), func() {
			suite.SetupTest()

			var (
				ctx           = suite.ChainA.GetContext()
				contractOwner = wasmKeeper.RandomAccountAddress(suite.T())
			)

			// Store code and instantiate reflect contract.
			codeID := suite.StoreTestCode(ctx, contractOwner, reflectContractPath)
			contractAddress := suite.InstantiateTestContract(ctx, contractOwner, codeID)
			suite.Require().NotEmpty(contractAddress)

			err := testutil.SetupICAPath(suite.Path, contractAddress.String())
			suite.Require().NoError(err)

			// Top up contract address with native coins for deposit
			senderAddress := suite.ChainA.SenderAccounts[0].SenderAccount.GetAddress()
			suite.TopUpWallet(ctx, senderAddress, contractAddress)

			stValues, height, expectedError := tt.malleate(contractAddress.String(), ctx)

			if expectedError != nil {
				suite.Require().ErrorContains(suite.GetNeutronZoneApp(suite.ChainA).StateVerifierKeeper.Verify(ctx, height, stValues), expectedError.Error())
			} else {
				suite.Require().NoError(suite.GetNeutronZoneApp(suite.ChainA).StateVerifierKeeper.Verify(ctx, height, stValues))
			}
		})
	}
}

func (suite *KeeperTestSuite) TopUpWallet(ctx sdk.Context, sender, contractAddress sdk.AccAddress) {
	coinsAmnt := sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(int64(1_000_000))))
	bankKeeper := suite.GetNeutronZoneApp(suite.ChainA).BankKeeper
	err := bankKeeper.SendCoins(ctx, sender, contractAddress, coinsAmnt)
	suite.Require().NoError(err)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
