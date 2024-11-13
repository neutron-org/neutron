package main

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/skip-mev/slinky/cmd/constants/marketmaps"
	"github.com/skip-mev/slinky/providers/apis/coinmarketcap"
	mmtypes "github.com/skip-mev/slinky/x/marketmap/types"
	"github.com/skip-mev/slinky/x/marketmap/types/tickermetadata"
)

var (
	convertToCMC     = flag.Bool("convert-to-cmc", false, "convert to coinmarketcap markets")
	marketFile       = flag.String("market-config-path", "", "market file to convert to coinmarketcap markets")
	autoEnable       = flag.Bool("auto-enable", false, "auto enable markets")
	isMMDeployment   = flag.Bool("is-mm-deployment", false, "is market map deployment")
	useCore          = flag.Bool("use-core", false, "use core markets")
	useRaydium       = flag.Bool("use-raydium", false, "use raydium markets")
	useUniswapV3Base = flag.Bool("use-uniswapv3-base", false, "use uniswapv3 base markets")
	useCoinGecko     = flag.Bool("use-coingecko", false, "use coingecko markets")
	useCoinMarketCap = flag.Bool("use-coinmarketcap", false, "use coinmarketcap markets")
	useOsmosis       = flag.Bool("use-osmosis", false, "use osmosis markets")
	usePolymarket    = flag.Bool("use-polymarket", false, "use polymarket markets")
	tempFile         = flag.String("temp-file", "markets.json", "temporary file to store the market map")
)

func main() {
	// Based on the flags, we determine what market.json to configure. By default, we use Core markets.
	// If the user specifies a different market.json, we use that instead.
	flag.Parse()

	if *isMMDeployment {
		if *marketFile == "" {
			fmt.Fprintf(flag.CommandLine.Output(), "market map config path (market-cfg-path) cannot be empty\n")
			panic("market map config path (market-cfg-path) cannot be empty")
		}

		marketMap, err := mmtypes.ReadMarketMapFromFile(*marketFile)
		if err != nil {
			fmt.Fprintf(flag.CommandLine.Output(), "failed to read market map from file: %s\n", err)
			panic(err)
		}

		if *convertToCMC {
			marketMap = filterToOnlyCMCMarkets(marketMap)
		}

		if *autoEnable {
			marketMap = enableAllMarkets(marketMap)
		}

		// Write the market map back to the original file.
		if err := mmtypes.WriteMarketMapToFile(marketMap, *marketFile); err != nil {
			fmt.Fprintf(flag.CommandLine.Output(), "failed to write market map to file: %s\n", err)
			panic(err)
		}

		return
	}

	marketMap := mmtypes.MarketMap{
		Markets: make(map[string]mmtypes.Market),
	}

	if *useCore {
		fmt.Fprintf(flag.CommandLine.Output(), "Using core markets\n")
		marketMap = mergeMarketMaps(marketMap, marketmaps.CoreMarketMap)
	}

	if *useRaydium {
		fmt.Fprintf(flag.CommandLine.Output(), "Using raydium markets\n")
		marketMap = mergeMarketMaps(marketMap, marketmaps.RaydiumMarketMap)
	}

	if *useUniswapV3Base {
		fmt.Fprintf(flag.CommandLine.Output(), "Using uniswapv3 base markets\n")
		marketMap = mergeMarketMaps(marketMap, marketmaps.UniswapV3BaseMarketMap)
	}

	if *useCoinGecko {
		fmt.Fprintf(flag.CommandLine.Output(), "Using coingecko markets\n")
		marketMap = mergeMarketMaps(marketMap, marketmaps.CoinGeckoMarketMap)
	}

	if *useCoinMarketCap {
		fmt.Fprintf(flag.CommandLine.Output(), "Using coinmarketcap markets\n")
		marketMap = mergeMarketMaps(marketMap, marketmaps.CoinMarketCapMarketMap)
	}

	if *useOsmosis {
		fmt.Fprintf(flag.CommandLine.Output(), "Using osmosis markets\n")
		marketMap = mergeMarketMaps(marketMap, marketmaps.OsmosisMarketMap)
	}

	if *usePolymarket {
		fmt.Fprintf(flag.CommandLine.Output(), "Using polymarket markets\n")
		marketMap = mergeMarketMaps(marketMap, marketmaps.PolymarketMarketMap)
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

// filterToOnlyCMCMarkets is a helper function that filters out all markets that are not from CoinMarketCap. It
// mutates the marketmap to only include CoinMarketCap markets. Notably the CMC ID will be pricing in the base
// asset.
func filterToOnlyCMCMarkets(marketmap mmtypes.MarketMap) mmtypes.MarketMap {
	res := mmtypes.MarketMap{
		Markets: make(map[string]mmtypes.Market),
	}

	// Filter out all markets that are not from CoinMarketCap.
	for _, market := range marketmap.Markets {
		var meta tickermetadata.DyDx
		if err := json.Unmarshal([]byte(market.Ticker.Metadata_JSON), &meta); err != nil {
			continue
		}

		var id string
		for _, aggregateID := range meta.AggregateIDs {
			if aggregateID.Venue == "coinmarketcap" {
				id = aggregateID.ID
				break
			}
		}

		if len(id) == 0 {
			continue
		}

		resTicker := market.Ticker
		resTicker.MinProviderCount = 1

		providers := []mmtypes.ProviderConfig{
			{
				Name:           coinmarketcap.Name,
				OffChainTicker: id,
			},
		}

		res.Markets[resTicker.CurrencyPair.String()] = mmtypes.Market{
			Ticker:          resTicker,
			ProviderConfigs: providers,
		}
	}

	return res
}

// enableAllMarkets is a helper function that enables all markets in the market map.
func enableAllMarkets(marketmap mmtypes.MarketMap) mmtypes.MarketMap {
	for name, market := range marketmap.Markets {
		market.Ticker.Enabled = true
		marketmap.Markets[name] = market
	}
	return marketmap
}
