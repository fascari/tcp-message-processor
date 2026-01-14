package config

import (
	"github.com/spf13/viper"
)

type (
	Config struct {
		Database DatabaseConfig
		Server   ServerConfig
		RabbitMQ RabbitMQConfig
	}

	DatabaseConfig struct {
		Host     string
		Port     string
		User     string
		Password string
		Name     string
		SSLMode  string
	}

	ServerConfig struct {
		Port              string
		Host              string
		BroadcastInterval int
	}

	RabbitMQConfig struct {
		URL      string
		Exchange string
		Queue    string
	}
)

func Load() (Config, error) {
	viper.SetConfigName("env")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/")

	viper.AutomaticEnv()

	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "tcpuser")
	viper.SetDefault("DB_PASSWORD", "tcppass")
	viper.SetDefault("DB_NAME", "tcpprocessor")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("SERVER_PORT", "8888")
	viper.SetDefault("SERVER_HOST", "0.0.0.0")
	viper.SetDefault("BROADCAST_INTERVAL_SECONDS", 30)
	viper.SetDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	viper.SetDefault("RABBITMQ_EXCHANGE", "submissions")
	viper.SetDefault("RABBITMQ_QUEUE", "submissions.processing")

	_ = viper.ReadInConfig()

	cfg := Config{
		Database: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			Name:     viper.GetString("DB_NAME"),
			SSLMode:  viper.GetString("DB_SSLMODE"),
		},
		Server: ServerConfig{
			Port:              viper.GetString("SERVER_PORT"),
			Host:              viper.GetString("SERVER_HOST"),
			BroadcastInterval: viper.GetInt("BROADCAST_INTERVAL_SECONDS"),
		},
		RabbitMQ: RabbitMQConfig{
			URL:      viper.GetString("RABBITMQ_URL"),
			Exchange: viper.GetString("RABBITMQ_EXCHANGE"),
			Queue:    viper.GetString("RABBITMQ_QUEUE"),
		},
	}

	return cfg, nil
}
