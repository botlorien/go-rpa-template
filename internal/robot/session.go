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

// ApplyHeaders é um Helper genérico.
// Ele recebe um mapa de headers e os aplica na requisição.
// Se 'force' for false, ele só aplica se o header ainda não existir na requisição.
func (s *Session) ApplyHeaders(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		// Só define se estiver vazio, permitindo override na ação específica se necessário
		if req.Header.Get(k) == "" {
			req.Header.Set(k, v)
		}
	}
}