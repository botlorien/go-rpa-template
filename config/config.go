package config

import (
	"log"
	"os"
	"github.com/spf13/viper"
)

type Config struct {
	BaseDir   string `mapstructure:"BASE_DIR"`
	PathDownload   string `mapstructure:"PATH_DOWNLOAD"`
	pathReports   string `mapstructure:"PATH_REPORTS"`
	AppPort  string `mapstructure:"APP_PORT"`
	TargetURL string `mapstructure:"TARGET_URL"`
	LogLevel string `mapstructure:"LOG_LEVEL"`
	Env       string `mapstructure:"APP_ENV"`    // local, prod
	UseRod      bool `mapstructure:"USE_ROD"`      // true = usa browser, false = usa http puro
	RodHeadless bool `mapstructure:"ROD_HEADLESS"` // true = sem tela
	BotAppURL  string `mapstructure:"BOTAPP_API_URL"`
	BotAppUser string `mapstructure:"BOTAPP_API_USUARIO"`
	BotAppPass string `mapstructure:"BOTAPP_API_SENHA"`
	DBDriver string `mapstructure:"DB_DRIVER"` // postgres, mysql, sqlite
    DBDSN    string `mapstructure:"DB_DSN"`    // Connection String
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env") // Ou config.yaml
	viper.AutomaticEnv()

	// Defaults
	rootDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Erro ao obter diretório atual: %v", err)
	}
	pathDownload := rootDir + string(os.PathSeparator) + "downloads"
	os.MkdirAll(pathDownload, 0755)
	pathReports := rootDir + string(os.PathSeparator) + "reports"
	os.MkdirAll(pathReports, 0755)

	viper.SetDefault("BASE_DIR", rootDir)
	viper.SetDefault("PATH_DOWNLOAD", pathDownload)
	viper.SetDefault("PATH_REPORTS", pathReports)

	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("APP_ENV", "local") // Por padrão é modo dev
	viper.SetDefault("USE_ROD", false)      // Padrão leve
	viper.SetDefault("ROD_HEADLESS", true)  // Padrão silencioso
	

	if err := viper.ReadInConfig(); err != nil {
		log.Println("Arquivo de config não encontrado, usando variáveis de ambiente.")
	}

	var cfg Config
	err = viper.Unmarshal(&cfg)
	return &cfg, err
}