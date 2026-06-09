package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
	"web3-api/internal/config"
	"web3-api/internal/model"
)

type evmProvider struct {
	client  *http.Client
	baseURL string
	apiKey  string
	chain   string // "eth" or "bnb" — used as Moralis chain param when network is empty
	network string // Moralis chain override: "sepolia", "bsc testnet", etc.
}

func newEVMProvider(chain string) ProviderFactory {
	return func(cfg config.ChainConfig) ChainProvider {
		return &evmProvider{
			client:  &http.Client{Timeout: 30 * time.Second},
			baseURL: cfg.BaseURL,
			apiKey:  cfg.APIKey,
			chain:   chain,
			network: cfg.Network,
		}
	}
}

func (p *evmProvider) moralisChain() string {
	if p.network != "" {
		return p.network
	}
	return p.chain
}

func (p *evmProvider) GetTokens(ctx context.Context, address string, cursor string) (*model.TokenPage, error) {
	url := fmt.Sprintf("%s/%s/erc20?chain=%s", p.baseURL, address, p.moralisChain())
	if cursor != "" {
		url += "&cursor=" + cursor
	}

	var resp moralisTokenResp
	if err := p.doRequest(ctx, url, &resp); err != nil {
		return nil, err
	}

	items := make([]model.Token, 0, len(resp))
	for _, t := range resp {
		items = append(items, model.Token{
			ContractAddress: t.TokenAddress,
			Name:            t.Name,
			Symbol:          t.Symbol,
			Decimals:        t.Decimals,
			Balance:         t.Balance,
			Logo:            t.Thumbnail,
		})
	}

	return &model.TokenPage{
		Items: items,
	}, nil
}

func (p *evmProvider) GetTransactions(ctx context.Context, address, contractAddress string, cursor string) (*model.TransactionPage, error) {
	url := fmt.Sprintf("%s/%s/erc20/transfers?chain=%s&token_address=%s", p.baseURL, address, p.moralisChain(), contractAddress)
	if cursor != "" {
		url += "&cursor=" + cursor
	}

	var resp moralisTransferResp
	if err := p.doRequest(ctx, url, &resp); err != nil {
		return nil, err
	}

	items := make([]model.Transaction, 0, len(resp.Result))
	for _, tx := range resp.Result {
		blockNum, _ := strconv.ParseUint(tx.BlockNumber, 10, 64)
		ts, _ := time.Parse(time.RFC3339, tx.BlockTimestamp)
		items = append(items, model.Transaction{
			Hash:        tx.TransactionHash,
			From:        tx.FromAddress,
			To:          tx.ToAddress,
			Value:       tx.Value,
			BlockNumber: blockNum,
			Timestamp:   ts,
			Status:      "success",
		})
	}

	return &model.TransactionPage{
		Items: items,
	}, nil
}

func (p *evmProvider) GetNativeTransactions(ctx context.Context, address string, cursor string) (*model.TransactionPage, error) {
	url := fmt.Sprintf("%s/%s?chain=%s", p.baseURL, address, p.moralisChain())
	if cursor != "" {
		url += "&cursor=" + cursor
	}

	var resp moralisNativeTxResp
	if err := p.doRequest(ctx, url, &resp); err != nil {
		return nil, err
	}

	items := make([]model.Transaction, 0, len(resp.Result))
	for _, tx := range resp.Result {
		blockNum, _ := strconv.ParseUint(tx.BlockNumber, 10, 64)
		ts, _ := time.Parse(time.RFC3339, tx.BlockTimestamp)
		status := "success"
		if tx.ReceiptStatus == "0" {
			status = "failed"
		}
		items = append(items, model.Transaction{
			Hash:        tx.Hash,
			From:        tx.FromAddress,
			To:          tx.ToAddress,
			Value:       tx.Value,
			BlockNumber: blockNum,
			Timestamp:   ts,
			Status:      status,
		})
	}

	return &model.TransactionPage{
		Items: items,
	}, nil
}

func (p *evmProvider) GetNFTs(ctx context.Context, address string, cursor string) (*model.NFTPage, error) {
	url := fmt.Sprintf("%s/%s/nft?chain=%s", p.baseURL, address, p.moralisChain())
	if cursor != "" {
		url += "&cursor=" + cursor
	}

	var resp moralisNFTResp
	if err := p.doRequest(ctx, url, &resp); err != nil {
		return nil, err
	}

	items := make([]model.NFT, 0, len(resp.Result))
	for _, n := range resp.Result {
		var metadata map[string]interface{}
		if n.Metadata != "" {
			_ = json.Unmarshal([]byte(n.Metadata), &metadata)
		}
		var imageURL string
		if img, ok := metadata["image"].(string); ok {
			imageURL = model.NormalizeMediaURL(img)
		}
		items = append(items, model.NFT{
			ContractAddress: n.TokenAddress,
			TokenID:         n.TokenID,
			Name:            n.Name,
			Image:           imageURL,
			CollectionName:  n.Name,
			Metadata:        metadata,
		})
	}

	return &model.NFTPage{
		Items: items,
	}, nil
}

func (p *evmProvider) doRequest(ctx context.Context, url string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-API-Key", p.apiKey)
	req.Header.Set("Accept", "application/json")

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
		return fmt.Errorf("moralis api error: status=%d body=%s", resp.StatusCode, string(body))
	}

	return json.Unmarshal(body, result)
}

// Moralis response types

type moralisTokenResp []struct {
	TokenAddress string `json:"token_address"`
	Name         string `json:"name"`
	Symbol       string `json:"symbol"`
	Decimals     int    `json:"decimals"`
	Balance      string `json:"balance"`
	Thumbnail    string `json:"thumbnail"`
}

type moralisTransferResp struct {
	Result []struct {
		TransactionHash  string `json:"transaction_hash"`
		FromAddress      string `json:"from_address"`
		ToAddress        string `json:"to_address"`
		Value            string `json:"value"`
		BlockNumber      string `json:"block_number"`
		BlockTimestamp   string `json:"block_timestamp"`
		TransactionIndex int    `json:"transaction_index"`
		LogIndex         int    `json:"log_index"`
	} `json:"result"`
	Cursor string `json:"cursor"`
}

type moralisNFTResp struct {
	Result []struct {
		TokenAddress string `json:"token_address"`
		TokenID      string `json:"token_id"`
		Name         string `json:"name"`
		Metadata     string `json:"metadata"`
	} `json:"result"`
}

type moralisNativeTxResp struct {
	Result []struct {
		Hash          string `json:"hash"`
		FromAddress   string `json:"from_address"`
		ToAddress     string `json:"to_address"`
		Value         string `json:"value"`
		BlockNumber   string `json:"block_number"`
		BlockTimestamp string `json:"block_timestamp"`
		ReceiptStatus string `json:"receipt_status"`
	} `json:"result"`
	Cursor string `json:"cursor"`
}
