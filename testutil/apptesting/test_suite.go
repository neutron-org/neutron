package apptesting

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/log"
	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/app"
	"github.com/neutron-org/neutron/testutil"
	dexmoduletypes "github.com/neutron-org/neutron/x/dex/types"
	"github.com/stretchr/testify/suite"
)

type KeeperTestHelper struct {
	suite.Suite

	App   *app.App
	Ctx   sdk.Context
	GoCtx context.Context
	// Used for testing queries end to end.
	// You can wrap this in a module-specific QueryClient()
	// and then make calls as you would a normal GRPC client.
	QueryHelper *baseapp.QueryServiceTestHelper
}

// Setup sets up basic environment for suite (App, Ctx, and test accounts)
func (s *KeeperTestHelper) Setup() {
	s.App = testutil.Setup(s.T(), false)
	s.Ctx = s.App.BaseApp.NewContext(
		false,
		tmtypes.Header{Height: 1, ChainID: "neutron-1", Time: time.Now().UTC()},
	)
	s.GoCtx = sdk.WrapSDKContext(s.Ctx)
	s.QueryHelper = &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: s.App.GRPCQueryRouter(),
		Ctx:             s.Ctx,
	}

	s.SetEpochStartTime()
}

func (s *KeeperTestHelper) SetEpochStartTime() {
	epochsKeeper := s.App.EpochsKeeper

	for _, epoch := range epochsKeeper.AllEpochInfos(s.Ctx) {
		epoch.StartTime = s.Ctx.BlockTime()
		epochsKeeper.DeleteEpochInfo(s.Ctx, epoch.Identifier)
		err := epochsKeeper.AddEpochInfo(s.Ctx, epoch)
		if err != nil {
			panic(err)
		}
	}
}

// setupAddr takes a balance, prefix, and address number. Then returns the respective account address byte array.
// If prefix is left blank, it will be replaced with a random prefix.
func SetupAddr(index int) sdk.AccAddress {
	prefixBz := make([]byte, 8)
	_, _ = rand.Read(prefixBz)
	prefix := string(prefixBz)
	addr := sdk.AccAddress([]byte(fmt.Sprintf("addr%s%8d", prefix, index)))
	return addr
}

func (s *KeeperTestHelper) SetupAddr(index int) sdk.AccAddress {
	return SetupAddr(index)
}

func SetupAddrs(numAddrs int) []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, numAddrs)
	for i := 0; i < numAddrs; i++ {
		addrs[i] = SetupAddr(i)
	}
	return addrs
}

func (s *KeeperTestHelper) SetupAddrs(numAddrs int) []sdk.AccAddress {
	return SetupAddrs(numAddrs)
}

// These are for testing msg.ValidateBasic() functions
// which need to validate for valid/invalid addresses.
// Should not be used for anything else because these addresses
// are totally uninterpretable (100% random).
func GenerateTestAddrs() (string, string) {
	pk1 := ed25519.GenPrivKey().PubKey()
	validAddr := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("").String()
	return validAddr, invalidAddr
}

// CreateTestContext creates a test context.
func (s *KeeperTestHelper) CreateTestContext() sdk.Context {
	ctx, _ := s.CreateTestContextWithMultiStore()
	return ctx
}

// CreateTestContextWithMultiStore creates a test context and returns it together with multi store.
func (s *KeeperTestHelper) CreateTestContextWithMultiStore() (sdk.Context, sdk.CommitMultiStore) {
	db := dbm.NewMemDB()
	logger := log.NewNopLogger()

	ms := rootmulti.NewStore(db, logger)

	return sdk.NewContext(ms, tmtypes.Header{}, false, logger), ms
}

// CreateTestContext creates a test context.
func (s *KeeperTestHelper) Commit() {
	s.App.EndBlock(abci.RequestEndBlock{Height: s.Ctx.BlockHeight()})
	oldHeight := s.Ctx.BlockHeight()
	oldHeader := s.Ctx.BlockHeader()
	s.App.Commit()
	newHeader := tmtypes.Header{
		Height:  oldHeight + 1,
		ChainID: oldHeader.ChainID,
		Time:    oldHeader.Time.Add(time.Minute),
	}
	s.App.BeginBlock(abci.RequestBeginBlock{Header: newHeader})
	s.Ctx = s.App.GetBaseApp().NewContext(false, newHeader)
}

// FundAcc funds target address with specified amount.
func (s *KeeperTestHelper) FundAcc(acc sdk.AccAddress, amounts sdk.Coins) {
	err := s.App.BankKeeper.MintCoins(s.Ctx, dexmoduletypes.ModuleName, amounts)
	s.Require().NoError(err)

	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, dexmoduletypes.ModuleName, acc, amounts)
	s.Require().NoError(err)
}
