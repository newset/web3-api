package service

import (
	"context"
	"fmt"
	"web3-api/internal/chain"
	"web3-api/internal/config"
	"web3-api/internal/model"
)

type Service struct {
	providers map[string]chain.ChainProvider
}

func New(cfg *config.Config) (*Service, error) {
	providers := make(map[string]chain.ChainProvider)
	for name, chainCfg := range cfg.Chains {
		p, err := chain.NewProvider(name, chainCfg)
		if err != nil {
			return nil, fmt.Errorf("init chain %s: %w", name, err)
		}
		providers[name] = p
	}
	return &Service{providers: providers}, nil
}

func (s *Service) GetTokens(ctx context.Context, chainName, address, cursor string) (*model.TokenPage, error) {
	p, ok := s.providers[chainName]
	if !ok {
		return nil, fmt.Errorf("unsupported chain: %s", chainName)
	}
	return p.GetTokens(ctx, address, cursor)
}

func (s *Service) GetTransactions(ctx context.Context, chainName, address, contractAddress, cursor string) (*model.TransactionPage, error) {
	p, ok := s.providers[chainName]
	if !ok {
		return nil, fmt.Errorf("unsupported chain: %s", chainName)
	}
	return p.GetTransactions(ctx, address, contractAddress, cursor)
}

func (s *Service) GetNativeTransactions(ctx context.Context, chainName, address, cursor string) (*model.TransactionPage, error) {
	p, ok := s.providers[chainName]
	if !ok {
		return nil, fmt.Errorf("unsupported chain: %s", chainName)
	}
	return p.GetNativeTransactions(ctx, address, cursor)
}

func (s *Service) GetNFTs(ctx context.Context, chainName, address, cursor string) (*model.NFTPage, error) {
	p, ok := s.providers[chainName]
	if !ok {
		return nil, fmt.Errorf("unsupported chain: %s", chainName)
	}
	return p.GetNFTs(ctx, address, cursor)
}
