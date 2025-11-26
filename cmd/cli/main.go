package main

import (
	"os"

	"github.com/botlorien/go-rpa-template/config"
	"github.com/botlorien/go-rpa-template/internal/robot"
	"github.com/botlorien/go-rpa-template/internal/repository"
	"github.com/botlorien/go-rpa-template/pkg/botapp"
	"github.com/botlorien/go-rpa-template/pkg/logger"
	"github.com/botlorien/go-rpa-template/pkg/database"
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

	// 3. Prepara a config específica do BotApp
	botConfig := botapp.Config{
		APIURL:   cfg.BotAppURL,
		User:     cfg.BotAppUser,
		Password: cfg.BotAppPass,
	}

	// 4. Inicializa o Client injetando a config
	app, err := botapp.NewClient(botConfig)
	if err != nil {
		log.Warn().Msg("BotApp API não configurada. Rodando sem logs remotos.")
		app = nil // O código deve tratar app == nil
	} else {
		// 5. Registra o Bot (set_bot do Python)
		err = app.SetBot("Meu Bot Go", "Descrição do bot", "1.0.0", "TI")
		if err != nil {
			log.Error().Err(err).Msg("Falha ao registrar bot na dashboard")
			os.Exit(1)
		}
	}

	// 6. Infra: Banco de Dados
	dbConn, err := database.NewConnection(cfg.DBDriver, cfg.DBDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("Erro no banco")
	}

	// 7. Camada Repository
	relatorioRepo := repository.NewRelatorioRepository(dbConn)

	// 8. Preparar o Input
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

	// 9. Inicializar Infraestrutura (Browser/HTTP)
	scraperSession := robot.NewSession(cfg.UseRod, cfg.RodHeadless, cfg.PathDownload)
	defer scraperSession.Close()

	// 10. Executar Robô com os Inputs
	robotService := robot.NewService(scraperSession, relatorioRepo, app)
	
	// Passamos o input criado acima
	// 11. Define a tarefa que será executada (Sua lógica de negócio)
	// Isso seria o conteúdo da função def minha_tarefa():
	minhaTarefa := func() (any, error) {
		// Chama seu Service aqui dentro
		return robotService.Execute(input)
	}

	// 12. Executa a tarefa "Envelopada" 
	var resultado any
	
	if app != nil {
		// Se a API estiver configurada, roda via RunTask (com logs na dashboard)
		// "Extrair Dados SSW" -> Nome que vai aparecer na Dash
		resultado, err = app.RunTask(
			"Execução Geral", 
			"Executa pipeline completa", 
			minhaTarefa,
		)
	} else {
		// Fallback: roda local sem logs na dashboard
		resultado, err = minhaTarefa()
	}

	if err != nil {
		log.Fatal().Err(err).Msg("Falha crítica na execução do RPA")
	}

	// (Opcional) Logar resultado resumido
	log.Info().
		Interface("resultado", resultado).
		Msg("Worker finalizado com sucesso")
	
	os.Exit(0)
}