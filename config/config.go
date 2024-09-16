package config

import (
	_ "github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

type Config struct {
	JackalConfig   JackalConfig    `yaml:"jackal_config" mapstructure:"jackal_config"`
	NetworksConfig []NetworkConfig `yaml:"networks_config" mapstructure:"networks_config"`
}

type JackalConfig struct {
	RPC      string `yaml:"rpc" mapstructure:"rpc"`
	GRPC     string `yaml:"grpc" mapstructure:"grpc"`
	SeedFile string `yaml:"seed_file" mapstructure:"seed_file"`
	Contract string `yaml:"contract" mapstructure:"contract"`
}

type NetworkConfig struct {
	Name     string `yaml:"name" mapstructure:"name"`
	RPC      string `yaml:"rpc" mapstructure:"rpc"`
	Contract string `yaml:"contract" mapstructure:"contract"`
	ChainID  int64  `yaml:"chain_id" mapstructure:"chain_id"`
	Finality int64  `yaml:"finality" mapstructure:"finality"`
}

func DefaultConfig() Config {
	return Config{
		JackalConfig: JackalConfig{
			RPC:      "https://testnet-rpc.jackalprotocol.com:443",
			GRPC:     "jackal-testnet-grpc.polkachu.com:17590",
			SeedFile: "seed.json",
			Contract: "jkl1tjaqmxerltfqgxwp5p9sdwr9hncmdvevg9mllac2qjqfrjte6wkqwnq3h9",
		},
		NetworksConfig: []NetworkConfig{
			{
				Name:     "Ethereum Sepolia",
				RPC:      "wss://ethereum-sepolia-rpc.publicnode.com",
				Contract: "0x730fdF2ee985Ac0F7792f90cb9e1E5485d340208",
				ChainID:  11155111,
				Finality: 2,
			},
			{
				Name:     "Base Sepolia",
				RPC:      "wss://base-sepolia-rpc.publicnode.com",
				Contract: "0x20738B8eaB736f24c7881bA48263ee60Eb2a0A2a",
				ChainID:  84532,
				Finality: 2,
			},
		},
	}
}

func (c Config) Export() ([]byte, error) {
	return yaml.Marshal(c)
}
