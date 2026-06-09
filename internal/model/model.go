package model

import (
	"strings"
	"time"
)

// NormalizeMediaURL converts ipfs:// and ar:// URLs to publicly accessible HTTP gateway URLs.
func NormalizeMediaURL(raw string) string {
	if raw == "" {
		return raw
	}
	switch {
	case strings.HasPrefix(raw, "ipfs://"):
		return "https://ipfs.io/ipfs/" + strings.TrimPrefix(raw, "ipfs://")
	case strings.HasPrefix(raw, "ar://"):
		return "https://arweave.net/" + strings.TrimPrefix(raw, "ar://")
	default:
		return raw
	}
}

// Response is the standard API response wrapper.
type Response struct {
	Code    int         `json:"code" example:"0"`
	Message string      `json:"message" example:"ok"`
	Data    interface{} `json:"data,omitempty"`
}

// Token represents an ERC20/TRC20/SPL token balance.
type Token struct {
	ContractAddress string `json:"contract_address" example:"0xdAC17F958D2ee523a2206206994597C13D831ec7"`
	Name            string `json:"name" example:"Tether USD"`
	Symbol          string `json:"symbol" example:"USDT"`
	Decimals        int    `json:"decimals" example:"6"`
	Balance         string `json:"balance" example:"1000000"`
	Logo            string `json:"logo,omitempty" example:"https://logo.cdn.example/usdt.png"`
	USDValue        string `json:"usd_value,omitempty" example:"1000.00"`
}

// TokenPage is a paginated list of tokens.
type TokenPage struct {
	Items      []Token `json:"items"`
	NextCursor string  `json:"next_cursor,omitempty"`
}

// Transaction represents a token transfer record.
type Transaction struct {
	Hash        string    `json:"hash" example:"0xabc123..."`
	From        string    `json:"from" example:"0x1234..."`
	To          string    `json:"to" example:"0x5678..."`
	Value       string    `json:"value" example:"1000000"`
	BlockNumber uint64    `json:"block_number" example:"18000000"`
	Timestamp   time.Time `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	Status      string    `json:"status" example:"success"`
}

// TransactionPage is a paginated list of transactions.
type TransactionPage struct {
	Items      []Transaction `json:"items"`
	NextCursor string        `json:"next_cursor,omitempty"`
}

// NFT represents a non-fungible token.
type NFT struct {
	ContractAddress string                 `json:"contract_address" example:"0x1234..."`
	TokenID         string                 `json:"token_id" example:"1234"`
	Name            string                 `json:"name" example:"CryptoPunk #1234"`
	Image           string                 `json:"image,omitempty" example:"https://img.cdn.example/1234.png"`
	CollectionName  string                 `json:"collection_name,omitempty" example:"CryptoPunks"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// NFTPage is a paginated list of NFTs.
type NFTPage struct {
	Items      []NFT  `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
}
