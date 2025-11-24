package robot

import (
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/rs/zerolog/log"
)

// Session segura as conexões.
// Renomeei de "ScraperSession" para "Session" para ficar mais limpo.
type Session struct {
	HTTPClient *http.Client
	Browser    *rod.Browser
	UseRod     bool
}

// NewSession inicializa o motor (Browser ou HTTP)
func NewSession(useRod bool, headless bool) *Session {
	// Configura HTTP Client com Cookies (Jar)
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}

	sess := &Session{
		HTTPClient: client,
		UseRod:     useRod,
	}

	if useRod {
		log.Info().Msg("Inicializando browser Rod...")
		u := launcher.New().Headless(headless).MustLaunch()
		browser := rod.New().ControlURL(u).MustConnect()
		sess.Browser = browser
	}

	return sess
}

func (s *Session) Close() {
	if s.Browser != nil {
		s.Browser.MustClose()
	}
}