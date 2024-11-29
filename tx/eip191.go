package tx

import (
	"context"
	"fmt"
	"strconv"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/aminojson"
)

const EIP191MessagePrefix = "\x19Ethereum Signed Message:\n"
const EIP191DomainSeparatorPrefix = "\x19\x01\""

// SignModeEIP191Handler defines the SIGN_MODE_DIRECT SignModeHandler
type SignModeEIP191Handler struct {
	*aminojson.SignModeHandler
}

// NewSignModeEIP191Handler returns a new SignModeEIP191Handler.
func NewSignModeEIP191Handler(options aminojson.SignModeHandlerOptions) *SignModeEIP191Handler {
	return &SignModeEIP191Handler{
		SignModeHandler: aminojson.NewSignModeHandler(options),
	}
}

var _ signing.SignModeHandler = SignModeEIP191Handler{}

// Mode implements signing.SignModeHandler.Mode.
func (SignModeEIP191Handler) Mode() signingv1beta1.SignMode {
	return signingv1beta1.SignMode_SIGN_MODE_EIP_191 //nolint
}

// GetSignBytes implements SignModeHandler.GetSignBytes
func (h SignModeEIP191Handler) GetSignBytes(
	ctx context.Context, data signing.SignerData, txData signing.TxData,
) ([]byte, error) {
	aminoJSONBz, err := h.SignModeHandler.GetSignBytes(ctx, data, txData)
	if err != nil {
		return nil, err
	}

	fmt.Printf("TODO: aminoJsonBz: %s", string(aminoJSONBz))

	//srvBz := append(append(
	//	[]byte(EIP191MessagePrefix),
	//	[]byte(strconv.Itoa(len(aminoJSONBz)))...,
	//), aminoJSONBz...)

	//"\x19\x01" ‖ domainSeparator ‖ hashStruct(message)

	srvBz := append(append(
		[]byte(EIP191MessagePrefix),
		[]byte(strconv.Itoa(len(aminoJSONBz)))...,
	), aminoJSONBz...)

	return srvBz, nil
}
