// copypasted from: https://github.com/skip-mev/slinky/blob/main/scripts/genesis.go

package main

import (
	"flag"
	"fmt"

	"github.com/skip-mev/slinky/cmd/constants"
	mmtypes "github.com/skip-mev/slinky/x/marketmap/types"
)

var (
	useCore          = flag.Bool("use-core", false, "use core markets")
	useRaydium       = flag.Bool("use-raydium", false, "use raydium markets")
	useUniswapV3Base = flag.Bool("use-uniswapv3-base", false, "use uniswapv3 base markets")
	useCoinGecko     = flag.Bool("use-coingecko", false, "use coingecko markets")
	tempFile         = flag.String("temp-file", "markets.json", "temporary file to store the market map")
)

func main() {
	// Based on the flags, we determine what market.json to configure. By default, we use Core markets.
	// If the user specifies a different market.json, we use that instead.
	flag.Parse()

	marketMap := mmtypes.MarketMap{
		Markets: make(map[string]mmtypes.Market),
	}

	if *useCore {
		fmt.Fprintf(flag.CommandLine.Output(), "Using core markets\n")
		marketMap = mergeMarketMaps(marketMap, constants.CoreMarketMap)
	}

	if *useRaydium {
		fmt.Fprintf(flag.CommandLine.Output(), "Using raydium markets\n")
		marketMap = mergeMarketMaps(marketMap, constants.RaydiumMarketMap)
	}

	if *useUniswapV3Base {
		fmt.Fprintf(flag.CommandLine.Output(), "Using uniswapv3 base markets\n")
		marketMap = mergeMarketMaps(marketMap, constants.UniswapV3BaseMarketMap)
	}

	if *useCoinGecko {
		fmt.Fprintf(flag.CommandLine.Output(), "Using coingecko markets\n")
		marketMap = mergeMarketMaps(marketMap, constants.CoinGeckoMarketMap)
	}

	if err := marketMap.ValidateBasic(); err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "failed to validate market map: %s\n", err)
		panic(err)
	}

	// Write the market map to the temporary file.
	if *tempFile == "" {
		fmt.Fprintf(flag.CommandLine.Output(), "temp file cannot be empty\n")
		panic("temp file cannot be empty")
	}

	if err := mmtypes.WriteMarketMapToFile(marketMap, *tempFile); err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "failed to write market map to file: %s\n", err)
		panic(err)
	}
}

// mergeMarketMaps merges the two market maps together. If a market already exists in one of the maps, we
// merge based on the provider set.
func mergeMarketMaps(this, other mmtypes.MarketMap) mmtypes.MarketMap {
	for name, otherMarket := range other.Markets {
		// If the market does not exist in this map, we add it.
		thisMarket, ok := this.Markets[name]
		if !ok {
			this.Markets[name] = otherMarket
			continue
		}

		seen := make(map[string]struct{})
		for _, provider := range thisMarket.ProviderConfigs {
			key := providerConfigToKey(provider)
			seen[key] = struct{}{}
		}

		for _, provider := range otherMarket.ProviderConfigs {
			key := providerConfigToKey(provider)
			if _, ok := seen[key]; !ok {
				thisMarket.ProviderConfigs = append(thisMarket.ProviderConfigs, provider)
			}
		}

		this.Markets[name] = thisMarket
	}

	return this
}

func providerConfigToKey(cfg mmtypes.ProviderConfig) string {
	return cfg.Name + cfg.OffChainTicker
}
