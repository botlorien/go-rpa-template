package robot

import (
	"errors"
	"github.com/rs/zerolog/log"
	"github.com/botlorien/go-rpa-template/internal/repository"
)

type Service struct {
	Session *Session
	Repo    *repository.RelatorioRepository
}

func NewService(s *Session, r *repository.RelatorioRepository) *Service {
	return &Service{
		Session: s,
		Repo:    r,
	}
}

// Execute agora aceita o input genérico
func (s *Service) Execute(input ExecutionInput) (map[string]any, error) {
	log.Info().Msg("Iniciando execução com parâmetros dinâmicos")

	// 1. Validação (Defensive Programming)
	// O Service decide O QUE é obrigatório para ESSE robô específico
	user := input.GetCredential("username")
	pass := input.GetCredential("password")

	if user == "" || pass == "" {
		return nil, errors.New("credenciais 'username' e 'password' são obrigatórias")
	}

	// 2. Chama a Ação de Login passando o mapa inteiro ou só o necessário
	if err := s.Session.Login(input.Auth); err != nil {
		return nil, err
	}

	// 3. Usa parâmetros extras (ex: Data do relatório)
	dataRelatorio, ok := input.Params["data_inicio"].(string)
	if !ok {
		dataRelatorio = "hoje" // valor default
	}

	log.Info().Str("data_alvo", dataRelatorio).Msg("Processando...")

	return map[string]any{"status": "ok"}, nil
}