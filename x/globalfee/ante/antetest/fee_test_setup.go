package antetest

import (
	"github.com/neutron-org/neutron/v3/testutil"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	neutronapp "github.com/neutron-org/neutron/v3/app"
	gaiaparams "github.com/neutron-org/neutron/v3/app/params"
	"github.com/neutron-org/neutron/v3/x/globalfee"
	gaiafeeante "github.com/neutron-org/neutron/v3/x/globalfee/ante"
	globfeetypes "github.com/neutron-org/neutron/v3/x/globalfee/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	app       *neutronapp.App
	ctx       sdk.Context
	clientCtx client.Context
	txBuilder client.TxBuilder
}

var testBondDenom = "uatom"

func (s *IntegrationTestSuite) SetupTest() {
	neutronapp.GetDefaultConfig()
	app := testutil.Setup(s.T())
	ctx := app.GetBaseApp().NewContext(false)

	encodingConfig := gaiaparams.MakeEncodingConfig()
	encodingConfig.Amino.RegisterConcrete(&testdata.TestMsg{}, "testdata.TestMsg", nil)
	testdata.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	s.ctx = ctx
	s.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
}

func (s *IntegrationTestSuite) SetupTestGlobalFeeStoreAndMinGasPrice(minGasPrice []sdk.DecCoin, globalFeeParams *globfeetypes.Params) (gaiafeeante.FeeDecorator, sdk.AnteHandler) {
	subspace := s.app.GetSubspace(globalfee.ModuleName)
	subspace.SetParamSet(s.ctx, globalFeeParams)
	s.ctx = s.ctx.WithMinGasPrices(minGasPrice).WithIsCheckTx(true)

	// setup staking bond denom to "uatom"
	// since it's "stake" per default
	//params := s.app.StakingKeeper.GetParams(s.ctx)
	//params.BondDenom = testBondDenom
	//err := s.app.StakingKeeper.SetParams(s.ctx, params)
	//s.Require().NoError(err)

	// build fee decorator
	feeDecorator := gaiafeeante.NewFeeDecorator(subspace)

	// chain fee decorator to antehandler
	antehandler := sdk.ChainAnteDecorators(feeDecorator)

	return feeDecorator, antehandler
}

func (s *IntegrationTestSuite) CreateTestTx(privs []cryptotypes.PrivKey, accNums []uint64, accSeqs []uint64, chainID string) (xauthsigning.Tx, error) {
	var sigsV2 []signing.SignatureV2
	signMode, err := xauthsigning.APISignModeToInternal(s.clientCtx.TxConfig.SignModeHandler().DefaultMode())
	if err != nil {
		return nil, err
	}
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  signMode,
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}

	if err := s.txBuilder.SetSignatures(sigsV2...); err != nil {
		return nil, err
	}

	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := xauthsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		sigV2, err := tx.SignWithPrivKey(
			s.ctx,
			signMode,
			signerData,
			s.txBuilder,
			priv,
			s.clientCtx.TxConfig,
			accSeqs[i],
		)
		if err != nil {
			return nil, err
		}

		sigsV2 = append(sigsV2, sigV2)
	}

	if err := s.txBuilder.SetSignatures(sigsV2...); err != nil {
		return nil, err
	}

	return s.txBuilder.GetTx(), nil
}
