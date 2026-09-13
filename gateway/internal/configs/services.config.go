package configs

import (
	"log/slog"

	"github.com/spf13/viper"
)

type ServiceDefinition struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Protocol string `mapstructure:"protocol"`
}

type ServiceConfig struct {
	Services map[string]ServiceDefinition `mapstructure:"services"`
}

func LoadServiceConfig(logger *slog.Logger) *ServiceConfig {
	v := viper.New()

	v.SetConfigName("service_config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")

	if err := v.ReadInConfig(); err != nil {
		logger.Error("Failed to read service config file", "error", err)
		panic(err)
	}

	var cfg ServiceConfig
	if err := v.Unmarshal(&cfg); err != nil {
		logger.Error("Failed to unmarshal service config", "error", err)
		panic(err)
	}
	logger.Info("Successfully Loaded Services Config")
	return &cfg
}
