package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Config -  and validate
type Config struct {
	HttpPort      int    `mapstructure:"http_port" validate:"required,min=4000,max=4040"`
	Directory     string `mapstructure:"directory" validate:"required,dir"`
	CheckInterval int    `mapstructure:"check_interval" validate:"required,min=1"`
	APIEndpoint   string `mapstructure:"api_endpoint" validate:"required,url"`
	QueueSize     int    `mapstructure:"queue_size" validate:"required,min=1"`
}

var config Config

// GetConfig - get config instance pointer
func GetConfig() (*Config, error) {
	// load config
	if err := LoadConfig(); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadConfig - check yaml file exists with viper and validate
func LoadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file - %w", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		return fmt.Errorf("error unmarshalling config - %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return err
	}

	return nil
}
