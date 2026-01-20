package integration

import (
	"fmt"

	"github.com/spf13/viper"
)

type (
	Config struct {
		DB       DBConfig       `mapstructure:"db"`
		RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
		Server   ServerConfig   `mapstructure:"server"`
	}

	DBConfig struct {
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Name     string `mapstructure:"name"`
		SSLMode  string `mapstructure:"sslmode"`
	}

	RabbitMQConfig struct {
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
	}

	ServerConfig struct {
		Host                     string `mapstructure:"host"`
		Port                     string `mapstructure:"port"`
		BroadcastIntervalSeconds string `mapstructure:"broadcast_interval_seconds"`
	}
)

func LoadConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("env.test")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("..")
	v.AddConfigPath("../..")
	v.AddConfigPath("../../..")
	v.AddConfigPath("./client")
	v.AddConfigPath("../client")
	v.AddConfigPath("../../client")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
