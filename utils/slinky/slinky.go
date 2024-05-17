package slinky

import (
	"encoding/json"
	"fmt"

	marketmaptypes "github.com/skip-mev/slinky/x/marketmap/types"
)

// ReadMarketsFromFile reads a market map configuration from a file at the given path.
func ReadMarketsFromFile(marketsBz []byte) ([]marketmaptypes.Market, error) {
	// Initialize the struct to hold the configuration
	var markets []marketmaptypes.Market

	// Unmarshal the JSON data into the config struct
	if err := json.Unmarshal(marketsBz, &markets); err != nil {
		return nil, fmt.Errorf("error unmarshalling config JSON: %w", err)
	}

	return markets, nil
}

func ToMarketMap(markets []marketmaptypes.Market) marketmaptypes.MarketMap {
	mm := marketmaptypes.MarketMap{
		Markets: make(map[string]marketmaptypes.Market, len(markets)),
	}

	for _, m := range markets {
		mm.Markets[m.Ticker.String()] = m
	}

	return mm
}
