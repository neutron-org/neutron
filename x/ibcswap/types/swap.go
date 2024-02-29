package types

import (
	"encoding/json"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/iancoleman/orderedmap"

	dextypes "github.com/neutron-org/neutron/v3/x/dex/types"
)

// PacketMetadata wraps the SwapMetadata. The root key in the incoming ICS20 transfer packet's memo needs to be set to the same
// value as the json tag in order for the swap middleware to process the swap.
type PacketMetadata struct {
	Swap *SwapMetadata `json:"swap"`
}

// SwapMetadata defines the parameters necessary to perform a swap utilizing the memo field from an incoming ICS20
// transfer packet. The next field is a string so that you can nest any arbitrary metadata to be handled
// further in the middleware stack or on the counterparty.
type SwapMetadata struct {
	*dextypes.MsgPlaceLimitOrder
	// If a value is provided for NeutronRefundAddress and the swap fails the Transfer.Amount will be moved to this address for later recovery.
	// If no NeutronRefundAddress is provided and a swap fails we will fail the ibc transfer and tokens will be refunded on the source chain.
	NeutronRefundAddress string `json:"refund-address,omitempty"`

	// Using JSONObject so that objects for next property will not be mutated by golang's lexicographic key sort on map keys during Marshal.
	// Supports primitives for Unmarshal/Marshal so that an escaped JSON-marshaled string is also valid.
	Next *JSONObject `json:"next,omitempty"`
}

// Validate ensures that all the required fields are present in the SwapMetadata and contain valid values.
func (sm SwapMetadata) Validate() error {
	if err := sm.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(ErrInvalidSwapMetadata, err.Error())
	}
	if sm.TokenIn == "" {
		return sdkerrors.Wrap(ErrInvalidSwapMetadata, "limit order tokenIn cannot be an empty string")
	}
	if sm.TokenOut == "" {
		return sdkerrors.Wrap(ErrInvalidSwapMetadata, "limit order tokenOut cannot be an empty string")
	}
	if sm.NeutronRefundAddress != "" {
		_, err := sdk.AccAddressFromBech32(sm.NeutronRefundAddress)
		if err != nil {
			return sdkerrors.Wrapf(dextypes.ErrInvalidAddress, "%s is not a valid Neutron address", sm.NeutronRefundAddress)
		}
	}

	if !sm.OrderType.IsFoK() {
		return sdkerrors.Wrap(ErrInvalidSwapMetadata, "Limit Order type must be FILL_OR_KILL")
	}

	return nil
}

// ContainsPFM checks if the Swapetadata is wrapping packet-forward-middleware
func (sm SwapMetadata) ContainsPFM() bool {
	if sm.Next == nil {
		return false
	}
	forward, _ := sm.Next.orderedMap.Get("forward")

	return forward != nil
}

// JSONObject is a wrapper type to allow either a primitive type or a JSON object.
// In the case the value is a JSON object, OrderedMap type is used so that key order
// is retained across Unmarshal/Marshal.
type JSONObject struct {
	obj        bool
	primitive  []byte
	orderedMap orderedmap.OrderedMap
}

// NewJSONObject is a constructor used for tests.
// The usage of JSONObject in the middleware is only json Marshal/Unmarshal
func NewJSONObject(object bool, primitive []byte, orderedMap orderedmap.OrderedMap) *JSONObject {
	return &JSONObject{
		obj:        object,
		primitive:  primitive,
		orderedMap: orderedMap,
	}
}

// UnmarshalJSON overrides the default json.Unmarshal behavior
func (o *JSONObject) UnmarshalJSON(b []byte) error {
	if err := o.orderedMap.UnmarshalJSON(b); err != nil {
		// If ordered map unmarshal fails, this is a primitive value
		o.obj = false
		// Attempt to unmarshal as string, this removes extra JSON escaping
		var primitiveStr string
		if err := json.Unmarshal(b, &primitiveStr); err != nil {
			o.primitive = b
			return nil
		}
		o.primitive = []byte(primitiveStr)
		return nil
	}
	// This is a JSON object, now stored as an ordered map to retain key order.
	o.obj = true
	return nil
}

// MarshalJSON overrides the default json.Marshal behavior
func (o *JSONObject) MarshalJSON() ([]byte, error) {
	if o.obj {
		// non-primitive, return marshaled ordered map.
		return o.orderedMap.MarshalJSON()
	}
	// primitive, return raw bytes.
	return o.primitive, nil
}
