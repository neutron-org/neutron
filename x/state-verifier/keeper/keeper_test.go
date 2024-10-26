package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	wasmKeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	ibchost "github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v5/app/params"
	"github.com/neutron-org/neutron/v5/testutil"
	iqtypes "github.com/neutron-org/neutron/v5/x/interchainqueries/types"
)

var reflectContractPath = "../../../wasmbinding/testdata/reflect.wasm"

type KeeperTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func (suite *KeeperTestSuite) TestVerifyValue() {
	tests := []struct {
		name     string
		malleate func(sender string, ctx sdk.Context)
	}{
		{
			name: "valid KV storage proof",
			malleate: func(sender string, ctx sdk.Context) {
				clientKey := host.FullClientStateKey(suite.Path.EndpointA.ClientID)

				resp, err := suite.ChainA.App.Query(ctx, &types.RequestQuery{
					Path:   fmt.Sprintf("store/%s/key", ibchost.StoreKey),
					Height: suite.ChainA.LastHeader.Header.Height - 1,
					Data:   clientKey,
					Prove:  true,
				})
				suite.Require().NoError(err)

				suite.Require().NoError(suite.GetNeutronZoneApp(suite.ChainA).StateVerifierKeeper.Verify(ctx, resp.Height, []*iqtypes.StorageValue{{
					Key:           resp.Key,
					Proof:         resp.ProofOps,
					Value:         resp.Value,
					StoragePrefix: ibchost.StoreKey,
				}}))
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

			tt.malleate(contractAddress.String(), ctx)
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
