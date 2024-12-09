package config

import (
	"online-library/internal/logger"

	"github.com/spf13/viper"
)

type Config struct {
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	APIFullURL string `mapstructure:"EXTERNAL_API_FULL_URL"`
	ServerPort string `mapstructure:"SERVER_PORT"`
	Method     string `mapstructure:"EXTERNAL_API_METHOD"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		logger.Log.Fatalf("error reading config file: %v", err)
	}
	var config Config

	logger.Log.Debug("Unmarshalling config data into struct...")

	if err := viper.Unmarshal(&config); err != nil {
		logger.Log.Errorf("Error unmarshalling config data: %v", err)
		return nil, err
	}

	logger.Log.Info("Configuration loaded successfully")
	return &config, nil
}
