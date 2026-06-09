package price

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type CoinGeckoProvider struct {
	client *http.Client
	idMap  map[string]string // chain → CoinGecko id
}

func NewCoinGeckoProvider(client *http.Client) *CoinGeckoProvider {
	return &CoinGeckoProvider{
		client: client,
		idMap: map[string]string{
			"eth":           "ethereum",
			"eth_sepolia":   "ethereum",
			"bnb":           "binancecoin",
			"bnb_testnet":   "binancecoin",
			"tron":          "tron",
			"tron_shasta":   "tron",
			"solana":        "solana",
			"solana_devnet": "solana",
		},
	}
}

func (p *CoinGeckoProvider) Name() string { return "coingecko" }

func (p *CoinGeckoProvider) GetPrices(ctx context.Context, chains []string) (map[string]PriceData, error) {
	ids := make([]string, 0, len(chains))
	chainByID := make(map[string]string)
	for _, chain := range chains {
		if id, ok := p.idMap[chain]; ok {
			if _, exists := chainByID[id]; !exists {
				ids = append(ids, id)
			}
			chainByID[id] = chain
		}
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("no valid chains")
	}

	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd&include_24hr_change=true",
		strings.Join(ids, ","))

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
		return nil, fmt.Errorf("coingecko status=%d body=%s", resp.StatusCode, string(body))
	}

	// {"ethereum":{"usd":1670.0,"usd_24h_change":-1.5}}
	var data map[string]map[string]float64
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	result := make(map[string]PriceData)
	for _, chain := range chains {
		id, ok := p.idMap[chain]
		if !ok {
			continue
		}
		coin, ok := data[id]
		if !ok {
			continue
		}
		usd := coin["usd"]
		change := coin["usd_24h_change"]
		// CoinGecko 直接给百分比变化，算回 open 价: open = price / (1 + change/100)
		open := usd / (1 + change/100)
		result[chain] = PriceData{
			Price: fmt.Sprintf("%.2f", usd),
			Open:  fmt.Sprintf("%.2f", open),
		}
	}

	return result, nil
}
