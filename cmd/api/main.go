package main

import (
	"os"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/botlorien/go-rpa-template/config"
	"github.com/botlorien/go-rpa-template/internal/robot"
	"github.com/botlorien/go-rpa-template/internal/repository"
	transport "github.com/botlorien/go-rpa-template/internal/transport/http" // Alias para não confundir com net/http
	"github.com/botlorien/go-rpa-template/pkg/logger"
	"github.com/botlorien/go-rpa-template/pkg/database"
	"github.com/botlorien/go-rpa-template/pkg/botapp"
)

func main() {
	// 1. Configuração e Logger
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger.Setup(cfg.LogLevel, cfg.Env)

	// 2. Setup do Framework Web
	if cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	// Middleware de log simples (opcional, já que temos log no handler)
	r.Use(gin.Logger())

	botConfig := botapp.Config{
		APIURL:   cfg.BotAppURL,
		User:     cfg.BotAppUser,
		Password: cfg.BotAppPass,
	}

	// 3. Inicializa o Client injetando a config
	app, err := botapp.NewClient(botConfig)
	if err != nil {
		log.Warn().Msg("BotApp API não configurada. Rodando sem logs remotos.")
		app = nil // O código deve tratar app == nil
	} else {
		// 2. Registra o Bot (set_bot do Python)
		err = app.SetBot("Meu Bot Go", "Descrição do bot", "1.0.0", "TI")
		if err != nil {
			log.Error().Err(err).Msg("Falha ao registrar bot na dashboard")
			os.Exit(1)
		}
	}
	// 4. Infra: Banco de Dados
	dbConn, err := database.NewConnection(cfg.DBDriver, cfg.DBDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("Erro no banco")
	}

	// 5. Camada Repository
	relatorioRepo := repository.NewRelatorioRepository(dbConn)

	// 6. Inicializa o Scraper (Singleton)
   scraperSession := robot.NewSession(cfg.UseRod, cfg.RodHeadless, cfg.PathDownload)
    
    // IMPORTANTE: Fecha o browser quando a API cair
    defer scraperSession.Close()

    // 7. Injeta no Service
    robotService := robot.NewService(scraperSession, relatorioRepo, app)

	// 8. Criamos o Handler HTTP e injetamos o Robô nele
	httpHandler := transport.NewHandler(robotService)

	// 9. O Handler registra suas próprias rotas no servidor
	httpHandler.RegisterRoutes(r)

	// ---------------------------------------------------------

	log.Info().Str("port", cfg.AppPort).Msg("Servidor API iniciado")
	r.Run(":" + cfg.AppPort)
}
