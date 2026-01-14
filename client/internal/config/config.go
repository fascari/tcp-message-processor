package config

import (
	"tcp-message-processor/common/pkg/env"

	"github.com/spf13/viper"
)

type (
	Config struct {
		Server ServerConfig
		Client ClientConfig
	}

	ServerConfig struct {
		Host string
		Port string
	}

	ClientConfig struct {
		Username             string
		SubmissionMinSeconds int
		SubmissionMaxSeconds int
	}
)

func Load() (Config, error) {
	viper.SetConfigName("env")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("../server")

	viper.AutomaticEnv()

	viper.SetDefault("SERVER_HOST", "localhost")
	viper.SetDefault("SERVER_PORT", "8888")
	viper.SetDefault("CLIENT_USERNAME", env.Lookup("USER", "client1"))
	viper.SetDefault("SUBMISSION_MIN_SECONDS", 1)
	viper.SetDefault("SUBMISSION_MAX_SECONDS", 1)

	_ = viper.ReadInConfig()

	return Config{
		Server: ServerConfig{
			Host: viper.GetString("SERVER_HOST"),
			Port: viper.GetString("SERVER_PORT"),
		},
		Client: ClientConfig{
			Username:             viper.GetString("CLIENT_USERNAME"),
			SubmissionMinSeconds: viper.GetInt("SUBMISSION_MIN_SECONDS"),
			SubmissionMaxSeconds: viper.GetInt("SUBMISSION_MAX_SECONDS"),
		},
	}, nil
}
