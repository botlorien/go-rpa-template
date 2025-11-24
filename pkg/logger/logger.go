package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Setup configura o logger globalmente
func Setup(level string, env string) {
	// 1. Define o nível (debug, info, warn, error)
	zerologLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		zerologLevel = zerolog.InfoLevel // Default
	}
	zerolog.SetGlobalLevel(zerologLevel)

	// 2. Define o formato de saída
	if env == "local" || env == "dev" {
		// Output bonito para terminal (colorido)
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		})
	} else {
		// Output JSON para Produção (Docker/Kubernetes)
		// É mais rápido e máquinas conseguem fazer query
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}
}