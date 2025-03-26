package v3_test

import (
	"testing"

	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"

	"github.com/neutron-org/neutron/v6/testutil"
	v3 "github.com/neutron-org/neutron/v6/x/interchainqueries/migrations/v3"
	"github.com/neutron-org/neutron/v6/x/interchainqueries/types"
)

type V3ICQMigrationTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(V3ICQMigrationTestSuite))
}

// ParamsV2 defines the parameters for the module v2.
type ParamsV2 struct {
	// Defines amount of blocks required before query becomes available for
	// removal by anybody
	QuerySubmitTimeout uint64 `protobuf:"varint,1,opt,name=query_submit_timeout,json=querySubmitTimeout,proto3" json:"query_submit_timeout,omitempty"`
	// Amount of coins deposited for the query.
	QueryDeposit github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,2,rep,name=query_deposit,json=queryDeposit,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"query_deposit"`
	// Amount of tx hashes to be removed during a single EndBlock. Can vary to
	// balance between network cleaning speed and EndBlock duration. A zero value
	// means no limit.
	TxQueryRemovalLimit uint64 `protobuf:"varint,3,opt,name=tx_query_removal_limit,json=txQueryRemovalLimit,proto3" json:"tx_query_removal_limit,omitempty"`
}

func (p *ParamsV2) Reset()        { *p = ParamsV2{} }
func (p *ParamsV2) ProtoMessage() {}

// String implements the Stringer interface.
func (p ParamsV2) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func (suite *V3ICQMigrationTestSuite) TestParamsMigration() {
	var (
		app      = suite.GetNeutronZoneApp(suite.ChainA)
		storeKey = app.GetKey(types.StoreKey)
		ctx      = suite.ChainA.GetContext()
		cdc      = app.AppCodec()
	)

	// preinitialize v2 params
	p := ParamsV2{
		QuerySubmitTimeout:  types.DefaultQuerySubmitTimeout,
		QueryDeposit:        types.DefaultQueryDeposit,
		TxQueryRemovalLimit: types.DefaultTxQueryRemovalLimit,
	}
	store := ctx.KVStore(storeKey)
	bz, err := cdc.Marshal(&p)
	suite.Require().NoError(err)
	store.Set(types.ParamsKey, bz)

	paramsOld := app.InterchainQueriesKeeper.GetParams(ctx)
	suite.Require().Equal(paramsOld.TxQueryRemovalLimit, p.TxQueryRemovalLimit)
	suite.Require().Equal(paramsOld.QuerySubmitTimeout, p.QuerySubmitTimeout)
	suite.Require().Equal(paramsOld.QueryDeposit, p.QueryDeposit)
	suite.Require().Equal(paramsOld.MaxTransactionsFilters, uint64(0))
	suite.Require().Equal(paramsOld.MaxKvQueryKeysCount, uint64(0))

	err = v3.MigrateParams(ctx, cdc, storeKey)
	suite.Require().NoError(err)

	paramsNew := app.InterchainQueriesKeeper.GetParams(ctx)
	params := types.Params{
		QuerySubmitTimeout:     types.DefaultQuerySubmitTimeout,
		QueryDeposit:           types.DefaultQueryDeposit,
		TxQueryRemovalLimit:    types.DefaultTxQueryRemovalLimit,
		MaxKvQueryKeysCount:    types.DefaultMaxKvQueryKeysCount,
		MaxTransactionsFilters: types.DefaultMaxTransactionsFilters,
	}
	suite.Require().Equal(params, paramsNew)
}
