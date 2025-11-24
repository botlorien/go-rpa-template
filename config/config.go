package config

import (
	"log"
	"github.com/spf13/viper"
)

type Config struct {
	AppPort  string `mapstructure:"APP_PORT"`
	TargetURL string `mapstructure:"TARGET_URL"`
	LogLevel string `mapstructure:"LOG_LEVEL"`
	Env       string `mapstructure:"APP_ENV"`    // local, prod
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env") // Ou config.yaml
	viper.AutomaticEnv()

	// Defaults
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("APP_ENV", "local") // Por padrão é modo dev

	if err := viper.ReadInConfig(); err != nil {
		log.Println("Arquivo de config não encontrado, usando variáveis de ambiente.")
	}

	var cfg Config
	err := viper.Unmarshal(&cfg)
	return &cfg, err
}