package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/botlorien/go-rpa-template/config"
	"github.com/botlorien/go-rpa-template/internal/robot"
	"github.com/botlorien/go-rpa-template/internal/repository"
	transport "github.com/botlorien/go-rpa-template/internal/transport/http" // Alias para não confundir com net/http
	"github.com/botlorien/go-rpa-template/pkg/logger"
	"github.com/botlorien/go-rpa-template/pkg/database"
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

	// ---------------------------------------------------------
	// 3. INJEÇÃO DE DEPENDÊNCIA (A mágica acontece aqui)
	// ---------------------------------------------------------
	// 1. Infra: Banco de Dados
	dbConn, err := database.NewConnection(cfg.DBDriver, cfg.DBDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("Erro no banco")
	}

	// 2. Camada Repository
	relatorioRepo := repository.NewRelatorioRepository(dbConn)
	// A. Criamos o Robô (Core Domain)
	// 1. Inicializa o Scraper (Singleton)
   scraperSession := robot.NewSession(cfg.UseRod, cfg.RodHeadless, cfg.PathDownload)
    
    // IMPORTANTE: Fecha o browser quando a API cair
    defer scraperSession.Close()

    // 2. Injeta no Service
    robotService := robot.NewService(scraperSession, relatorioRepo)

	// B. Criamos o Handler HTTP e injetamos o Robô nele
	httpHandler := transport.NewHandler(robotService)

	// C. O Handler registra suas próprias rotas no servidor
	httpHandler.RegisterRoutes(r)

	// ---------------------------------------------------------

	log.Info().Str("port", cfg.AppPort).Msg("Servidor API iniciado")
	r.Run(":" + cfg.AppPort)
}
