package http

import (
	"net/http"

	"github.com/botlorien/go-rpa-template/internal/robot"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Handler segura as dependências necessárias para lidar com as requisições
type Handler struct {
	Service *robot.Service
}

// NewHandler é o construtor
func NewHandler(s *robot.Service) *Handler {
	return &Handler{
		Service: s,
	}
}

// RegisterRoutes define as rotas que esse handler atende
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1") // Boa prática: versionamento
	{
		api.POST("/run", h.RunRPA)
		api.GET("/health", h.HealthCheck)
	}
}

// HealthCheck é um endpoint simples
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "up"})
}

// RunRPA é a função que o Gin vai chamar
func (h *Handler) RunRPA(c *gin.Context) {
	// 1. Log da entrada (Contexto HTTP)
	log.Info().
		Str("client_ip", c.ClientIP()).
		Msg("Recebida solicitação de execução via HTTP")

	// 2. Chama o Service (O Robô)
	// Note que o handler não sabe COMO o robô funciona, só pede para executar.
	data, err := h.Service.Execute()

	if err != nil {
		log.Error().Err(err).Msg("Erro na execução do serviço")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. Retorna a resposta (Tradução para HTTP)
	c.JSON(http.StatusOK, data)
}
