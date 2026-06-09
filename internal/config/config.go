package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type ChainConfig struct {
	BaseURL string `mapstructure:"base_url"`
	APIKey  string `mapstructure:"api_key"`
	Network string `mapstructure:"network"` // Moralis chain param: sepolia, bsc testnet, etc.
}

type Config struct {
	Server struct {
		Port int    `mapstructure:"port"`
		Mode string `mapstructure:"mode"`
	} `mapstructure:"server"`
	Chains map[string]ChainConfig `mapstructure:"chains"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.SetEnvPrefix("WEB3")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// expand env vars in api keys
	for name, chain := range cfg.Chains {
		if strings.HasPrefix(chain.APIKey, "${") && strings.HasSuffix(chain.APIKey, "}") {
			envKey := chain.APIKey[2 : len(chain.APIKey)-1]
			chain.APIKey = viper.GetString(envKey)
			cfg.Chains[name] = chain
		}
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}

	return &cfg, nil
}
