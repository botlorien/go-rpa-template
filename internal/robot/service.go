package robot

import (
	"errors"
	"github.com/rs/zerolog/log"
	"github.com/botlorien/go-rpa-template/internal/repository"
	"github.com/botlorien/go-rpa-template/pkg/utils"
	"github.com/botlorien/go-rpa-template/pkg/botapp"
)

type Service struct {
	Session *Session
	Repo    *repository.RelatorioRepository
	App     *botapp.Client
}

func NewService(s *Session, r *repository.RelatorioRepository, a *botapp.Client) *Service {
	return &Service{
		Session: s,
		Repo:    r,
		App:	 a,
	}
}

// Execute agora aceita o input genérico
func (s *Service) Execute(input ExecutionInput) (any, error) {
	    log.Info().Str("dir", s.Session.DownloadDir).Msg("Limpando diretório de trabalho...")
    
    if err := utils.EmptyDirectory(s.Session.DownloadDir); err != nil {
        // Se não conseguir limpar, é perigoso continuar
        log.Error().Err(err).Msg("Falha ao limpar pasta de downloads")
        return nil, err
    }
	log.Info().Msg("Iniciando execução com parâmetros dinâmicos")

	// 1. Validação (Defensive Programming)
	// O Service decide O QUE é obrigatório para ESSE robô específico
	user := input.GetCredential("username")
	pass := input.GetCredential("password")

	if user == "" || pass == "" {
		return nil, errors.New("credenciais 'username' e 'password' são obrigatórias")
	}

	// various tasks can be added here
	var resultado any
	var err error

	loginTask := func() (any, error){
		if err := s.Session.Login(input.Auth); err != nil {
			return nil, err
		}
		return nil, nil
	}

	if s.App != nil {
		resultado, err = s.App.RunTask(
			"LoginTask",
			"Description Task", 
			loginTask,
	)

	} else {
		resultado, err = loginTask()
	}
	return resultado, err
}