package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"web3-api/internal/config"
	"web3-api/internal/model"
)

type tronProvider struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

func NewTronProvider(cfg config.ChainConfig) ChainProvider {
	return &tronProvider{
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
	}
}

func (p *tronProvider) GetTokens(ctx context.Context, address string, cursor string) (*model.TokenPage, error) {
	url := fmt.Sprintf("%s/v1/accounts/%s/trc20", p.baseURL, address)
	if cursor != "" {
		url += "&fingerprint=" + cursor
	}

	var resp tronTokenResp
	if err := p.doRequest(ctx, url, &resp); err != nil {
		return nil, err
	}

	items := make([]model.Token, 0, len(resp.Data))
	for _, t := range resp.Data {
		items = append(items, model.Token{
			ContractAddress: t.ContractAddress,
			Name:            t.Name,
			Symbol:          t.Symbol,
			Decimals:        t.Decimals,
			Balance:         t.Balance,
		})
	}

	return &model.TokenPage{
		Items:      items,
		NextCursor: resp.Meta.Fingerprint,
	}, nil
}

func (p *tronProvider) GetTransactions(ctx context.Context, address, contractAddress string, cursor string) (*model.TransactionPage, error) {
	url := fmt.Sprintf("%s/v1/accounts/%s/trc20/transfers?contract_address=%s", p.baseURL, address, contractAddress)
	if cursor != "" {
		url += "&fingerprint=" + cursor
	}

	var resp tronTransferResp
	if err := p.doRequest(ctx, url, &resp); err != nil {
		return nil, err
	}

	items := make([]model.Transaction, 0, len(resp.Data))
	for _, tx := range resp.Data {
		ts := time.UnixMilli(tx.BlockTimestamp)
		status := "success"
		if !tx.Ret {
			status = "failed"
		}
		items = append(items, model.Transaction{
			Hash:        tx.TransactionID,
			From:        tx.From,
			To:          tx.To,
			Value:       tx.Value,
			BlockNumber: 0, // TronGrid doesn't always return block number
			Timestamp:   ts,
			Status:      status,
		})
	}

	return &model.TransactionPage{
		Items:      items,
		NextCursor: resp.Meta.Fingerprint,
	}, nil
}

func (p *tronProvider) GetNativeTransactions(ctx context.Context, address string, cursor string) (*model.TransactionPage, error) {
	url := fmt.Sprintf("%s/v1/accounts/%s/transactions", p.baseURL, address)
	if cursor != "" {
		url += "&fingerprint=" + cursor
	}

	var resp tronNativeTxResp
	if err := p.doRequest(ctx, url, &resp); err != nil {
		return nil, err
	}

	items := make([]model.Transaction, 0, len(resp.Data))
	for _, tx := range resp.Data {
		ts := time.UnixMilli(tx.BlockTimestamp)
		status := "success"
		if !tx.Ret {
			status = "failed"
		}
		items = append(items, model.Transaction{
			Hash:        tx.TxID,
			From:        tx.OwnerAddress,
			To:          tx.ToAddress,
			Value:       fmt.Sprintf("%d", tx.Amount),
			BlockNumber: 0,
			Timestamp:   ts,
			Status:      status,
		})
	}

	return &model.TransactionPage{
		Items:      items,
		NextCursor: resp.Meta.Fingerprint,
	}, nil
}

func (p *tronProvider) GetNFTs(ctx context.Context, address string, cursor string) (*model.NFTPage, error) {
	url := fmt.Sprintf("%s/v1/accounts/%s/nft", p.baseURL, address)
	if cursor != "" {
		url += "&fingerprint=" + cursor
	}

	var resp tronNFTResp
	if err := p.doRequest(ctx, url, &resp); err != nil {
		return nil, err
	}

	items := make([]model.NFT, 0, len(resp.Data))
	for _, n := range resp.Data {
		var metadata map[string]interface{}
		if n.Metadata != "" {
			_ = json.Unmarshal([]byte(n.Metadata), &metadata)
		}
		var imageURL string
		if img, ok := metadata["image"].(string); ok {
			imageURL = model.NormalizeMediaURL(img)
		}
		items = append(items, model.NFT{
			ContractAddress: n.ContractAddress,
			TokenID:         n.TokenID,
			Name:            n.Name,
			Image:           imageURL,
			Metadata:        metadata,
		})
	}

	return &model.NFTPage{
		Items:      items,
		NextCursor: resp.Meta.Fingerprint,
	}, nil
}

func (p *tronProvider) doRequest(ctx context.Context, url string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if p.apiKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", p.apiKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("trongrid api error: status=%d body=%s", resp.StatusCode, string(body))
	}

	return json.Unmarshal(body, result)
}

// TronGrid response types

type tronTokenResp struct {
	Data []struct {
		ContractAddress string `json:"contract_address"`
		Name            string `json:"name"`
		Symbol          string `json:"symbol"`
		Decimals        int    `json:"decimals"`
		Balance         string `json:"balance"`
	} `json:"data"`
	Meta struct {
		Fingerprint string `json:"fingerprint"`
	} `json:"meta"`
}

type tronTransferResp struct {
	Data []struct {
		TransactionID string `json:"transaction_id"`
		From          string `json:"from"`
		To            string `json:"to"`
		Value         string `json:"value"`
		BlockTimestamp int64 `json:"block_timestamp"`
		Ret           bool   `json:"ret"`
	} `json:"data"`
	Meta struct {
		Fingerprint string `json:"fingerprint"`
	} `json:"meta"`
}

type tronNFTResp struct {
	Data []struct {
		ContractAddress string `json:"contract_address"`
		TokenID         string `json:"token_id"`
		Name            string `json:"name"`
		Metadata        string `json:"metadata"`
	} `json:"data"`
	Meta struct {
		Fingerprint string `json:"fingerprint"`
	} `json:"meta"`
}

type tronNativeTxResp struct {
	Data []struct {
		TxID           string `json:"txID"`
		BlockTimestamp int64  `json:"block_timestamp"`
		OwnerAddress   string `json:"owner_address"`
		ToAddress      string `json:"to_address"`
		Amount         int64  `json:"amount"`
		Ret            bool   `json:"ret"`
	} `json:"data"`
	Meta struct {
		Fingerprint string `json:"fingerprint"`
	} `json:"meta"`
}
