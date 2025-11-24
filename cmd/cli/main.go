package main

import (
	"os"

	"github.com/botlorien/go-rpa-template/config"
	"github.com/botlorien/go-rpa-template/internal/robot"
	"github.com/botlorien/go-rpa-template/pkg/logger"
	"github.com/rs/zerolog/log"
)

func main() {
	// 1. Carregar Config
	cfg, err := config.Load()
	if err != nil {
		// Se falhar config, usamos panic pois não dá pra continuar
		panic("Falha ao carregar config: " + err.Error())
	}

	// 2. Setup do Logger (Passamos o ENV e o Nível)
	logger.Setup(cfg.LogLevel, cfg.Env)

	log.Info().Msg("Iniciando Worker de RPA via CLI...")

	// 3. Executar Robô
	bot := robot.NewService(cfg.TargetURL)
	_, err = bot.Execute()

	if err != nil {
		// Loga o erro com stacktrace se configurado e sai com código 1
		log.Fatal().Err(err).Msg("Falha crítica na execução do RPA")
	}

	log.Info().Msg("Worker finalizado com sucesso")
	os.Exit(0)
}
