package robot

import (
	"errors"
	"time"
	"github.com/rs/zerolog/log"
)

// LoginSSW é uma "Ação" específica do sistema SSW
func (s *Session) Login(creds map[string]string) error {
	if user, ok := creds["username"]; ok {
		pass := creds["password"]
		if s.UseRod {
			return s.loginRod(user, pass)
		}
		return s.loginHTTP(user, pass)
	}
	return errors.New("nenhuma estratégia de autenticação válida encontrada")

}

// Implementação privada via ROD (Browser)
func (s *Session) loginRod(user, pass string) error {
	log.Debug().Msg("Realizando login via Browser")
	
	page := s.Browser.MustPage("https://ssw.inf.br/login")
	page.MustWaitLoad()

	// Exemplo hipotético de seletores
	page.MustElement("#username").Input(user)
	page.MustElement("#password").Input(pass)
	page.MustElement("#btn-entrar").MustClick()

	// Verifica se logou
	if hasError, _ := page.Element(".alert-error"); hasError != nil {
		return errors.New("usuário ou senha inválidos")
	}
	
	return nil
}

// Implementação privada via HTTP (Request)
func (s *Session) loginHTTP(user, pass string) error {
	log.Debug().Msg("Realizando login via HTTP Request")
	// Lógica de Post Form aqui...
	return nil
}

// Outra ação: Baixar Relatório
func (s *Session) BaixarRelatorio() ([]byte, error) {
    // Lógica para navegar até o relatório e baixar
    return []byte("conteudo do pdf"), nil
}