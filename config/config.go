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
	UseRod      bool `mapstructure:"USE_ROD"`      // true = usa browser, false = usa http puro
	RodHeadless bool `mapstructure:"ROD_HEADLESS"` // true = sem tela
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env") // Ou config.yaml
	viper.AutomaticEnv()

	// Defaults
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("APP_ENV", "local") // Por padrão é modo dev
	viper.SetDefault("USE_ROD", false)      // Padrão leve
	viper.SetDefault("ROD_HEADLESS", true)  // Padrão silencioso
	

	if err := viper.ReadInConfig(); err != nil {
		log.Println("Arquivo de config não encontrado, usando variáveis de ambiente.")
	}

	var cfg Config
	err := viper.Unmarshal(&cfg)
	return &cfg, err
}