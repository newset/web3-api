package chain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"web3-api/internal/config"
	"web3-api/internal/model"
)

type solanaProvider struct {
	client  *http.Client
	rpcURL  string
	apiKey  string
	dasURL  string // Helius DAS endpoint
}

func NewSolanaProvider(cfg config.ChainConfig) ChainProvider {
	return &solanaProvider{
		client: &http.Client{Timeout: 30 * time.Second},
		rpcURL: cfg.BaseURL,
		apiKey: cfg.APIKey,
		dasURL: cfg.BaseURL, // Helius uses same base for DAS
	}
}

func (p *solanaProvider) GetTokens(ctx context.Context, address string, cursor string) (*model.TokenPage, error) {
	url := fmt.Sprintf("%s/?api-key=%s", p.rpcURL, p.apiKey)

	reqBody := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "getTokenAccountsByOwner",
		Params: []interface{}{
			address,
			map[string]string{"programId": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"},
			map[string]string{"encoding": "jsonParsed"},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var resp jsonrpcResponse
	if err := p.doJSONRPCRequest(ctx, url, body, &resp); err != nil {
		return nil, err
	}

	var parsed solanaTokenAccountsResp
	if err := json.Unmarshal(resp.Result, &parsed); err != nil {
		return nil, fmt.Errorf("unmarshal token accounts: %w", err)
	}

	items := make([]model.Token, 0, len(parsed.Value))
	for _, v := range parsed.Value {
		info := v.Account.Data.Parsed.Info
		dec := int(info.TokenAmount.Decimals)
		items = append(items, model.Token{
			ContractAddress: info.Mint,
			Symbol:          info.TokenAmount.Symbol,
			Decimals:        dec,
			Balance:         info.TokenAmount.Amount,
		})
	}

	return &model.TokenPage{Items: items}, nil
}

func (p *solanaProvider) GetTransactions(ctx context.Context, address, contractAddress string, cursor string) (*model.TransactionPage, error) {
	url := fmt.Sprintf("%s/?api-key=%s", p.rpcURL, p.apiKey)

	reqBody := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "getSignaturesForAddress",
		Params: []interface{}{
			address,
			map[string]interface{}{"limit": 20},
		},
	}

	if cursor != "" {
		reqBody.Params = append(reqBody.Params, map[string]string{"before": cursor})
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var resp jsonrpcResponse
	if err := p.doJSONRPCRequest(ctx, url, body, &resp); err != nil {
		return nil, err
	}

	var sigs []solanaSignatureInfo
	if err := json.Unmarshal(resp.Result, &sigs); err != nil {
		return nil, fmt.Errorf("unmarshal signatures: %w", err)
	}

	items := make([]model.Transaction, 0, len(sigs))
	for _, sig := range sigs {
		ts := time.Unix(sig.BlockTime, 0)
		status := "success"
		if sig.Err != nil {
			status = "failed"
		}
		items = append(items, model.Transaction{
			Hash:      sig.Signature,
			Timestamp: ts,
			Status:    status,
		})
	}

	var nextCursor string
	if len(sigs) > 0 {
		nextCursor = sigs[len(sigs)-1].Signature
	}

	return &model.TransactionPage{
		Items:      items,
		NextCursor: nextCursor,
	}, nil
}

func (p *solanaProvider) GetNativeTransactions(ctx context.Context, address string, cursor string) (*model.TransactionPage, error) {
	url := fmt.Sprintf("%s/?api-key=%s", p.rpcURL, p.apiKey)

	reqBody := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "getSignaturesForAddress",
		Params: []interface{}{
			address,
			map[string]interface{}{"limit": 20},
		},
	}

	if cursor != "" {
		reqBody.Params = append(reqBody.Params, map[string]string{"before": cursor})
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var resp jsonrpcResponse
	if err := p.doJSONRPCRequest(ctx, url, body, &resp); err != nil {
		return nil, err
	}

	var sigs []solanaSignatureInfo
	if err := json.Unmarshal(resp.Result, &sigs); err != nil {
		return nil, fmt.Errorf("unmarshal signatures: %w", err)
	}

	items := make([]model.Transaction, 0, len(sigs))
	for _, sig := range sigs {
		ts := time.Unix(sig.BlockTime, 0)
		status := "success"
		if sig.Err != nil {
			status = "failed"
		}
		items = append(items, model.Transaction{
			Hash:      sig.Signature,
			Timestamp: ts,
			Status:    status,
		})
	}

	var nextCursor string
	if len(sigs) > 0 {
		nextCursor = sigs[len(sigs)-1].Signature
	}

	return &model.TransactionPage{
		Items:      items,
		NextCursor: nextCursor,
	}, nil
}

func (p *solanaProvider) GetNFTs(ctx context.Context, address string, cursor string) (*model.NFTPage, error) {
	url := fmt.Sprintf("%s/?api-key=%s", p.rpcURL, p.apiKey)

	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getAssetsByOwner",
		"params": map[string]interface{}{
			"ownerAddress": address,
			"sortBy": map[string]string{
				"sortBy":        "created",
				"sortDirection": "desc",
			},
			"limit": 20,
			"page":  1,
			"displayOptions": map[string]bool{
				"showFungible": false,
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var resp jsonrpcResponse
	if err := p.doJSONRPCRequest(ctx, url, body, &resp); err != nil {
		return nil, err
	}

	var parsed solanaAssetsResp
	if len(resp.Result) > 0 && string(resp.Result) != "null" {
		if err := json.Unmarshal(resp.Result, &parsed); err != nil {
			return nil, fmt.Errorf("unmarshal assets: %w", err)
		}
	}

	items := make([]model.NFT, 0, len(parsed.Items))
	for _, a := range parsed.Items {
		nft := model.NFT{
			ContractAddress: a.ID,
			TokenID:         a.Content.Metadata.TokenStandard,
			Name:            a.Content.Metadata.Name,
		}
		if len(a.Content.Files) > 0 {
			nft.Image = model.NormalizeMediaURL(a.Content.Files[0].URI)
		}
		if len(a.Grouping) > 0 {
			nft.CollectionName = a.Grouping[0].Value
		}
		items = append(items, nft)
	}

	return &model.NFTPage{Items: items}, nil
}

func (p *solanaProvider) doJSONRPCRequest(ctx context.Context, url string, body []byte, result *jsonrpcResponse) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("helius rpc error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	if err := json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("unmarshal rpc response: %w", err)
	}
	if result.Error != nil {
		return fmt.Errorf("rpc error %d: %s", result.Error.Code, result.Error.Message)
	}
	return nil
}

// JSON-RPC types

type jsonrpcRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type jsonrpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *jsonrpcError   `json:"error,omitempty"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Solana response types

type solanaTokenAccountsResp struct {
	Value []struct {
		Account struct {
			Data struct {
				Parsed struct {
					Info struct {
						Mint        string `json:"mint"`
						TokenAmount struct {
							Amount   string `json:"amount"`
							Decimals int    `json:"decimals"`
							Symbol   string `json:"symbol"`
						} `json:"tokenAmount"`
					} `json:"info"`
				} `json:"parsed"`
			} `json:"data"`
		} `json:"account"`
	} `json:"value"`
}

type solanaSignatureInfo struct {
	Signature string         `json:"signature"`
	BlockTime int64          `json:"blockTime"`
	Err       map[string]interface{} `json:"err"`
}

type solanaAssetsResp struct {
	Total int `json:"total"`
	Items []struct {
		ID      string `json:"id"`
		Content struct {
			Metadata struct {
				Name           string `json:"name"`
				TokenStandard  string `json:"token_standard"`
			} `json:"metadata"`
			Files []struct {
				URI string `json:"uri"`
			} `json:"files"`
		} `json:"content"`
		Grouping []struct {
			Value string `json:"value"`
		} `json:"grouping"`
	} `json:"items"`
}
