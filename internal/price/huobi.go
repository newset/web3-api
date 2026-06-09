package price

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HuobiProvider struct {
	client    *http.Client
	symbolMap map[string]string // chain → Huobi symbol (e.g. ethusdt)
}

func NewHuobiProvider(client *http.Client) *HuobiProvider {
	return &HuobiProvider{
		client: client,
		symbolMap: map[string]string{
			"eth":           "ethusdt",
			"eth_sepolia":   "ethusdt",
			"bnb":           "bnbusdt",
			"bnb_testnet":   "bnbusdt",
			"tron":          "trxusdt",
			"tron_shasta":   "trxusdt",
			"solana":        "solusdt",
			"solana_devnet": "solusdt",
		},
	}
}

func (p *HuobiProvider) Name() string { return "huobi" }

func (p *HuobiProvider) GetPrices(ctx context.Context, chains []string) (map[string]PriceData, error) {
	// 收集需要的 symbol，去重
	needSymbols := make(map[string]string) // symbol → chain (取第一个)
	for _, chain := range chains {
		if sym, ok := p.symbolMap[chain]; ok {
			if _, exists := needSymbols[sym]; !exists {
				needSymbols[sym] = chain
			}
		}
	}
	if len(needSymbols) == 0 {
		return nil, fmt.Errorf("no valid chains")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.huobi.pro/market/tickers", nil)
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
		return nil, fmt.Errorf("huobi status=%d body=%s", resp.StatusCode, string(body))
	}

	var data huobiTickersResp
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	if data.Status != "ok" {
		return nil, fmt.Errorf("huobi response not ok")
	}

	// 从全部 ticker 中筛选需要的
	tickerBySymbol := make(map[string]huobiTicker)
	for _, t := range data.Data {
		if _, ok := needSymbols[t.Symbol]; ok {
			tickerBySymbol[t.Symbol] = t
		}
	}

	result := make(map[string]PriceData)
	for _, chain := range chains {
		sym, ok := p.symbolMap[chain]
		if !ok {
			continue
		}
		if t, ok := tickerBySymbol[sym]; ok {
			result[chain] = PriceData{
				Price: fmt.Sprintf("%.2f", t.Close),
				Open:  fmt.Sprintf("%.2f", t.Open),
			}
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no prices matched")
	}
	return result, nil
}

type huobiTickersResp struct {
	Status string        `json:"status"`
	Data   []huobiTicker `json:"data"`
}

type huobiTicker struct {
	Symbol string  `json:"symbol"`
	Close  float64 `json:"close"`
	Open   float64 `json:"open"`
}
