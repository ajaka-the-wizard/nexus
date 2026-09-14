package configs

import (
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type ServiceDefinition struct {
	Host             string `mapstructure:"host" validate:"required,gte=1"`
	Port             int    `mapstructure:"port" validate:"required,gte=1"`
	Protocol         string `mapstructure:"protocol" validate:"required,gte=1"`
	RequestPerMinute uint   `mapstructure:"request_per_minute" validate:"required,gte=1"`
}

type ServiceConfig struct {
	Services map[string]ServiceDefinition `mapstructure:"services" validate:"required"`
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

	validator := validator.New()
	if err := validator.Struct(&cfg); err != nil {
		logger.Error("Service configuration validation failed", "error", err)
		panic(err)
	}

	for name, svc := range cfg.Services {
		if err := validator.Struct(svc); err != nil {
			logger.Error("Invalid service definition", "service", name, "error", err)
			panic(err)
		}
	}

	logger.Info("Successfully Loaded Services Config")
	return &cfg
}
