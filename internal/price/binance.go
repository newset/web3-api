package price

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type BinanceProvider struct {
	client    *http.Client
	symbolMap map[string]string
}

func NewBinanceProvider(client *http.Client) *BinanceProvider {
	return &BinanceProvider{
		client: client,
		symbolMap: map[string]string{
			"eth":           "ETHUSDT",
			"eth_sepolia":   "ETHUSDT",
			"bnb":           "BNBUSDT",
			"bnb_testnet":   "BNBUSDT",
			"tron":          "TRXUSDT",
			"tron_shasta":   "TRXUSDT",
			"solana":        "SOLUSDT",
			"solana_devnet": "SOLUSDT",
		},
	}
}

func (p *BinanceProvider) Name() string { return "binance" }

func (p *BinanceProvider) GetPrices(ctx context.Context, chains []string) (map[string]PriceData, error) {
	symbols := make([]string, 0, len(chains))
	chainBySymbol := make(map[string]string)
	for _, chain := range chains {
		if sym, ok := p.symbolMap[chain]; ok {
			symbols = append(symbols, sym)
			chainBySymbol[sym] = chain
		}
	}
	if len(symbols) == 0 {
		return nil, fmt.Errorf("no valid chains")
	}

	quoted := make([]string, len(symbols))
	for i, s := range symbols {
		quoted[i] = `"` + s + `"`
	}
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/24hr?symbols=[%s]", strings.Join(quoted, ","))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status=%d body=%s", resp.StatusCode, string(body))
	}

	var tickers []binance24hrTicker
	if err := json.Unmarshal(body, &tickers); err != nil {
		return nil, err
	}

	result := make(map[string]PriceData)
	for _, t := range tickers {
		if chain, ok := chainBySymbol[t.Symbol]; ok {
			result[chain] = PriceData{
				Price: t.LastPrice,
				Open:  t.OpenPrice,
			}
		}
	}
	return result, nil
}

type binance24hrTicker struct {
	Symbol    string `json:"symbol"`
	LastPrice string `json:"lastPrice"`
	OpenPrice string `json:"openPrice"`
}
