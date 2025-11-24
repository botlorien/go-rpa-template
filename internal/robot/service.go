package robot

import (
	"io"
	"github.com/rs/zerolog/log"
)

type Service struct {
	TargetURL string
	Scraper   *ScraperSession
}

func NewService(url string, scraper *ScraperSession) *Service {
	return &Service{
		TargetURL: url,
		Scraper:   scraper,
	}
}

func (s *Service) Execute() (map[string]any, error) {
	log.Info().Str("url", s.TargetURL).Msg("Iniciando scraping...")

	// Decide qual estratégia usar
	if s.Scraper.UseRod {
		return s.runRodStrategy()
	}
	return s.runHTTPStrategy()
}

// Estratégia 1: Navegação via Browser (Rod)
func (s *Service) runRodStrategy() (map[string]any, error) {
	page := s.Scraper.Browser.MustPage(s.TargetURL)
	page.MustWaitLoad() // Espera o JS carregar

	// Exemplo: Pegar o título da página
	title := page.MustEval(`() => document.title`).Str()
	
	log.Info().Str("engine", "rod").Str("title", title).Msg("Dados extraídos")

	return map[string]any{"method": "rod", "title": title}, nil
}

// Estratégia 2: Requisição HTTP Pura (Session)
func (s *Service) runHTTPStrategy() (map[string]any, error) {
	resp, err := s.Scraper.HTTPClient.Get(s.TargetURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Lê o corpo (simples)
	body, _ := io.ReadAll(resp.Body)
	size := len(body)

	log.Info().Str("engine", "http").Int("size", size).Msg("Dados baixados")

	return map[string]any{"method": "http", "size": size}, nil
}