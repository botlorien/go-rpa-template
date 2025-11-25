package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/botlorien/go-rpa-template/config"
	"github.com/botlorien/go-rpa-template/internal/robot"
	transport "github.com/botlorien/go-rpa-template/internal/transport/http" // Alias para não confundir com net/http
	"github.com/botlorien/go-rpa-template/pkg/logger"
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

	// A. Criamos o Robô (Core Domain)
	// 1. Inicializa o Scraper (Singleton)
   scraperSession := robot.NewSession(cfg.UseRod, cfg.RodHeadless, cfg.PathDownload)
    
    // IMPORTANTE: Fecha o browser quando a API cair
    defer scraperSession.Close()

    // 2. Injeta no Service
    robotService := robot.NewService(scraperSession)

	// B. Criamos o Handler HTTP e injetamos o Robô nele
	httpHandler := transport.NewHandler(robotService)

	// C. O Handler registra suas próprias rotas no servidor
	httpHandler.RegisterRoutes(r)

	// ---------------------------------------------------------

	log.Info().Str("port", cfg.AppPort).Msg("Servidor API iniciado")
	r.Run(":" + cfg.AppPort)
}
