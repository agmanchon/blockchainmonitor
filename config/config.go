package config

import (
	"errors"
	"fmt"
	"github.com/agmanchon/blockchainmonitor/libs/backend/pkg/postgresdb"
	"github.com/asaskevich/govalidator"
	"github.com/spf13/viper"
)

const (
	ConfigPath = "configPath"
)

type InstallConfig struct {
	PostgresConfig    postgresdb.PostgresConnectionFwOptions `mapstructure:"database" valid:"required"`
	KafkaConfig       KafkaOptions                           `mapstructure:"queue" valid:"required"`
	BlockchainNetwork BlockchainNetworkOptions               `mapstructure:"blockchainNetwork" valid:"required"`
}
type BlockchainNetworkOptions struct {
	Blockdaemon *BlockNetworkInfraOptions `mapstructure:"blockdaemon"`
}

type BlockNetworkInfraOptions struct {
	APIKey string                 `mapstructure:"apiKey" valid:"required~apiKey is mandatory in Kafka config"`
	Eth    *BlockchainNetworkInfo `mapstructure:"eth"`
	Btc    *BlockchainNetworkInfo `mapstructure:"btc"`
	Sol    *BlockchainNetworkInfo `mapstructure:"sol"`
}

type BlockchainNetworkInfo struct {
	Host string `mapstructure:"host" valid:"required~host is mandatory in Kafka config"`
}

type KafkaOptions struct {
	Kafka *KafkaInfo `mapstructure:"kafka"`
}
type KafkaInfo struct {
	Brokers  string `mapstructure:"brokers" valid:"required~port is mandatory in Kafka config"`
	Port     int    `mapstructure:"port" valid:"required~port is mandatory in Kafka config"`
	Username string `mapstructure:"username" json:"-" valid:"optional~username is optional in Kafka config"`
	Password string `mapstructure:"password" json:"-" valid:"optional~password is optional in Kafka config"`
	SSLMode  string `mapstructure:"sslMode" valid:"optional~sslMode is optional in Kafka config"`
	Topic    string `mapstructure:"topic" valid:"required~topic is mandatory in Kafka config"`
}

func ReadInstallConfig() (*InstallConfig, error) {
	// find config file
	viper.SetConfigType("yaml")
	var configPath = GetConfigPath()
	if len(configPath) < 1 {
		return nil, errors.New("configuration path is required")
	}

	// read config file
	viper.AddConfigPath(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("configuration could not be read: %w", err)
	}

	// unmarshall config file content
	var installConfig InstallConfig
	err = viper.Unmarshal(&installConfig)
	if err != nil {
		return nil, fmt.Errorf("configuration could not be unmarshalled: %w", err)
	}

	// validate config
	valid, err := validateInstallConfig(installConfig)
	if !valid {
		return nil, fmt.Errorf("configuration is not valid: %w", err)
	}

	return &installConfig, nil
}

func GetConfigPath() string {
	return viper.GetString(ConfigPath)
}

func validateInstallConfig(theConfig InstallConfig) (bool, error) {
	valid, err := govalidator.ValidateStruct(theConfig)
	if !valid || err != nil {
		return valid, err
	}
	return true, nil
}
