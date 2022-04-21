package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetTxMsgs casts the attached *types.Any messages to SDK messages.
func (m *MsgSubmitTx) GetTxMsgs() (sdkMsgs []sdk.Msg, err error) {
	for idx, msg := range m.Msgs {
		sdkMsg, ok := msg.GetCachedValue().(sdk.Msg)
		if !ok {
			return nil, fmt.Errorf("failed to cast message #%d to sdk.Msg", idx)
		}

		sdkMsgs = append(sdkMsgs, sdkMsg)
	}

	return
}
