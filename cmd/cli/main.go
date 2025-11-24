package main

import (
	"os"

	"github.com/botlorien/go-rpa-template/config"
	"github.com/botlorien/go-rpa-template/internal/robot"
	"github.com/botlorien/go-rpa-template/pkg/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper" // Importante: Viper mantém o estado do config carregado
)

func main() {
	// 1. Carregar Config (Isso lê o .env e prepara o Viper)
	cfg, err := config.Load()
	if err != nil {
		panic("Falha ao carregar config: " + err.Error())
	}

	// 2. Setup do Logger
	logger.Setup(cfg.LogLevel, cfg.Env)

	log.Info().Msg("Iniciando Worker de RPA via CLI...")

	// 3. Preparar o Input
	// Mapeamos as variáveis de ambiente para a Struct de entrada do Robô.
	// O Viper pega tanto do .env quanto das vars do Sistema Operacional.
	input := robot.ExecutionInput{
		Auth: map[string]string{
			"username": viper.GetString("RPA_USERNAME"),
			"password": viper.GetString("RPA_PASSWORD"),
			"token":    viper.GetString("RPA_TOKEN"),
		},
		Params: map[string]any{
			"data_inicio": viper.GetString("RPA_DATA_INICIO"),
			"filtro_id":   viper.GetString("RPA_FILTRO_ID"),
			"baixar_pdf":  viper.GetBool("RPA_BAIXAR_PDF"), // Viper converte tipos!
		},
	}

	// 4. Inicializar Infraestrutura (Browser/HTTP)
	scraperSession := robot.NewSession(cfg.UseRod, cfg.RodHeadless)
	defer scraperSession.Close()

	// 5. Executar Robô com os Inputs
	robotService := robot.NewService(scraperSession)
	
	// Passamos o input criado acima
	data, err := robotService.Execute(input)

	if err != nil {
		log.Fatal().Err(err).Msg("Falha crítica na execução do RPA")
	}

	// (Opcional) Logar resultado resumido
	log.Info().
		Interface("resultado", data).
		Msg("Worker finalizado com sucesso")
	
	os.Exit(0)
}