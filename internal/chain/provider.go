package chain

import (
	"context"
	"fmt"
	"web3-api/internal/config"
	"web3-api/internal/model"
)

type ChainProvider interface {
	GetTokens(ctx context.Context, address string, cursor string) (*model.TokenPage, error)
	GetTransactions(ctx context.Context, address, contractAddress string, cursor string) (*model.TransactionPage, error)
	GetNativeTransactions(ctx context.Context, address string, cursor string) (*model.TransactionPage, error)
	GetNFTs(ctx context.Context, address string, cursor string) (*model.NFTPage, error)
}

type ProviderFactory func(cfg config.ChainConfig) ChainProvider

var registry = map[string]ProviderFactory{
	"eth":          newEVMProvider("eth"),
	"eth_sepolia":  newEVMProvider("eth"),
	"bnb":          newEVMProvider("bnb"),
	"bnb_testnet":  newEVMProvider("bnb"),
	"tron":         NewTronProvider,
	"tron_shasta":  NewTronProvider,
	"solana":       NewSolanaProvider,
	"solana_devnet": NewSolanaProvider,
}

func NewProvider(chain string, cfg config.ChainConfig) (ChainProvider, error) {
	factory, ok := registry[chain]
	if !ok {
		return nil, fmt.Errorf("unsupported chain: %s", chain)
	}
	return factory(cfg), nil
}

func RegisterChain(name string, factory ProviderFactory) {
	registry[name] = factory
}
