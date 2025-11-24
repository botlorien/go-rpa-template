package robot

import (
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/rs/zerolog/log"
)

// ScraperSession segura tanto o cliente HTTP quanto o Browser
type ScraperSession struct {
	// Cliente HTTP padrão (com Cookies persistentes)
	HTTPClient *http.Client
	
	// Instância do Browser (Rod) - Só é iniciada se necessário
	Browser    *rod.Browser
	
	UseRod     bool
}

// NewScraper inicializa o motor escolhido
func NewScraper(useRod bool, headless bool) *ScraperSession {
	// 1. Configura Sessão HTTP (Sempre útil)
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}

	session := &ScraperSession{
		HTTPClient: client,
		UseRod:     useRod,
	}

	// 2. Se for usar Rod, inicializa o Browser
	if useRod {
		log.Info().Msg("Inicializando browser Rod...")
		
		// Launcher ajuda a encontrar o binário do Chrome no Docker/Local
		u := launcher.New().
			Headless(headless).
			MustLaunch()

		browser := rod.New().ControlURL(u).MustConnect()
		session.Browser = browser
	}

	return session
}

// Close garante a limpeza de recursos do browser
func (s *ScraperSession) Close() {
	if s.Browser != nil {
		s.Browser.MustClose()
	}
}