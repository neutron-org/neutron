package apptesting

import (
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/sims"

	//nolint:staticcheck
	"github.com/cosmos/cosmos-sdk/testutil/network"

	"github.com/neutron-org/neutron/v6/app"
)

// NewAppConstructor returns a new Osmosis app given encoding type configs.
func NewAppConstructor(chainId string) network.AppConstructor {
	return func(val network.ValidatorI) servertypes.Application {
		valCtx := val.GetCtx()
		appConfig := val.GetAppConfig()

		return app.New(
			valCtx.Logger, dbm.NewMemDB(), nil, true, make(map[int64]bool), valCtx.Config.RootDir, 0,
			app.MakeEncodingConfig(),
			sims.EmptyAppOptions{},
			nil,
			baseapp.SetMinGasPrices(appConfig.MinGasPrices),
			baseapp.SetChainID(chainId),
		)
	}
}
