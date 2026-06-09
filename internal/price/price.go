package price

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type PriceData struct {
	Price string
	Open  string
}

type PriceInfo struct {
	Symbol    string `json:"symbol" example:"ETH"`
	Price     string `json:"price" example:"1670.00"`
	Chain     string `json:"chain" example:"eth"`
	Change24h string `json:"change_24h" example:"-1.25"`
}

type Provider interface {
	Name() string
	GetPrices(ctx context.Context, chains []string) (map[string]PriceData, error)
}

type Service struct {
	providers []Provider
	allChains []string

	mu       sync.RWMutex
	cache    map[string]PriceInfo
	interval time.Duration
}

func NewService() *Service {
	client := &http.Client{Timeout: 10 * time.Second}
	s := &Service{
		providers: []Provider{
			NewCoinGeckoProvider(client),
			NewHuobiProvider(client),
			NewBinanceProvider(client),
		},
		allChains: []string{"eth", "bnb", "tron", "solana"},
		cache:     make(map[string]PriceInfo),
		interval:  30 * time.Second,
	}
	// 首次同步加载 + 后台定时刷新
	s.refresh()
	go s.loop()
	return s
}

func (s *Service) AllChains() []string {
	return s.allChains
}

func (s *Service) GetPrice(ctx context.Context, chain string) (*PriceInfo, error) {
	results, err := s.GetPrices(ctx, []string{chain})
	if err != nil {
		return nil, err
	}
	return &results[0], nil
}

func (s *Service) GetPrices(ctx context.Context, chains []string) ([]PriceInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]PriceInfo, 0, len(chains))
	for _, chain := range chains {
		if info, ok := s.cache[chain]; ok {
			result = append(result, info)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no cached prices available, try again later")
	}
	return result, nil
}

func (s *Service) loop() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for range ticker.C {
		s.refresh()
	}
}

func (s *Service) refresh() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var lastErr error
	for _, p := range s.providers {
		prices, err := p.GetPrices(ctx, s.allChains)
		if err != nil {
			lastErr = fmt.Errorf("%s: %w", p.Name(), err)
			continue
		}
		s.mu.Lock()
		for _, chain := range s.allChains {
			if d, ok := prices[chain]; ok {
				s.cache[chain] = PriceInfo{
					Symbol:    chainSymbol(chain),
					Price:     d.Price,
					Chain:     chain,
					Change24h: calcChange(d.Open, d.Price),
				}
			}
		}
		s.mu.Unlock()
		return
	}
	// all providers failed, keep stale cache
	_ = lastErr
}

func chainSymbol(chain string) string {
	symbols := map[string]string{
		"eth": "ETH", "eth_sepolia": "ETH",
		"bnb": "BNB", "bnb_testnet": "BNB",
		"tron": "TRX", "tron_shasta": "TRX",
		"solana": "SOL", "solana_devnet": "SOL",
	}
	if s, ok := symbols[chain]; ok {
		return s
	}
	return strings.ToUpper(chain)
}

func calcChange(openStr, closeStr string) string {
	open, err1 := parseFloat(openStr)
	close, err2 := parseFloat(closeStr)
	if err1 != nil || err2 != nil || open == 0 {
		return "0"
	}
	pct := (close - open) / open * 100
	return fmt.Sprintf("%.2f", pct)
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
