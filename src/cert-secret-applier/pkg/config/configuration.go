package config

import (
	"github.com/spf13/viper"
	"strings"
)

var config *Config

func Global() *Config {
	return config
}

type Config struct {
	Duckdns      Duckdns      `mapstructure:"duckdns"`
	Kubernetes   Kubernetes   `mapstructure:"kubernetes"`
	Logger       Logger       `mapstructure:"logger"`
}
  
type Duckdns struct {
	Domain string   `mapstructure:"string"`
}

type Kubernetes struct {
	Service Service   `mapstructure:"service"`
}

type Service struct {
	Host string   `mapstructure:"host"`
	Port Port     `mapstructure:"port"`
}

type Port struct {
	HTTP int   `mapstructure:"https"`
}

type Logger struct {
	Level string `mapstructure:"level"`
}

func FakeInit(cfg *Config) {
	config = cfg
}

func LoadConfig(path, fileName string) error {
	//overwrite config.yml with environment variables
	//example -> server-port -> SERVER_PORT
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.SetConfigName(fileName)
	viper.AddConfigPath(path)
	viper.AutomaticEnv()
	viper.SetConfigType("yml")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return viper.Unmarshal(&config)
}
