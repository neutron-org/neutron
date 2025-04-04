package antetest

import (
	"github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/neutron-org/neutron/v6/app/config"

	"github.com/neutron-org/neutron/v6/testutil"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	neutronapp "github.com/neutron-org/neutron/v6/app"
	gaiaparams "github.com/neutron-org/neutron/v6/app/params"
	gaiafeeante "github.com/neutron-org/neutron/v6/x/globalfee/ante"
	globfeetypes "github.com/neutron-org/neutron/v6/x/globalfee/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	app       *neutronapp.App
	ctx       sdk.Context
	clientCtx client.Context
	txBuilder client.TxBuilder
}

func (s *IntegrationTestSuite) SetupTest() {
	config.GetDefaultConfig()
	s.app = testutil.Setup(s.T()).(*neutronapp.App)
	ctx := s.app.GetBaseApp().NewUncachedContext(false, types.Header{})

	encodingConfig := gaiaparams.MakeEncodingConfig()
	encodingConfig.Amino.RegisterConcrete(&testdata.TestMsg{}, "testdata.TestMsg", nil)
	testdata.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	s.ctx = ctx
	s.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
}

func (s *IntegrationTestSuite) SetupTestGlobalFeeStoreAndMinGasPrice(minGasPrice []sdk.DecCoin, globalFeeParams *globfeetypes.Params) (gaiafeeante.FeeDecorator, sdk.AnteHandler) {
	err := s.app.GlobalFeeKeeper.SetParams(s.ctx, *globalFeeParams)
	s.Require().NoError(err)
	s.ctx = s.ctx.WithMinGasPrices(minGasPrice).WithIsCheckTx(true)

	// build fee decorator
	feeDecorator := gaiafeeante.NewFeeDecorator(s.app.GlobalFeeKeeper)

	// chain fee decorator to antehandler
	antehandler := sdk.ChainAnteDecorators(feeDecorator)

	return feeDecorator, antehandler
}

func (s *IntegrationTestSuite) CreateTestTx(privs []cryptotypes.PrivKey, accNums, accSeqs []uint64, chainID string) (xauthsigning.Tx, error) {
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
			PubKey:        priv.PubKey(),
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
