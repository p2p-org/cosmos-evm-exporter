package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	EVMAddress      string `toml:"evm_address"`
	TargetValidator string `toml:"target_validator"`
	ETHEndpoint     string `toml:"eth_endpoint"`
	RPCEndpoint     string `toml:"rpc_endpoint"`
	MetricsPort     string `toml:"metrics_port"`
	LogFile         string `toml:"log_file"`
	EnableFileLog   bool   `toml:"enable_file_log"`
	EnableStdout    bool   `toml:"enable_stdout"`
}

func LoadConfig(path string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
