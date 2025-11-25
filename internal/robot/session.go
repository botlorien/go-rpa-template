package robot

import (
	"net/http"
	"net/http/cookiejar"
	"path/filepath"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/rs/zerolog/log"
)

// Session segura as conexões.
// Renomeei de "ScraperSession" para "Session" para ficar mais limpo.
type Session struct {
	HTTPClient *http.Client
	Browser    *rod.Browser
	UseRod     bool
	DownloadDir string
}

// NewSession inicializa o motor (Browser ou HTTP)
func NewSession(useRod bool, headless bool, downloadDir string) *Session {
	// Garante caminho absoluto para o Chrome não se perder
	absDownloadDir, err := filepath.Abs(downloadDir)
	if err != nil {
		absDownloadDir = downloadDir // Fallback
	}
	// Configura HTTP Client com Cookies (Jar)
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}

	sess := &Session{
		HTTPClient: client,
		UseRod:     useRod,
		DownloadDir: absDownloadDir,
	}

	if useRod {
		log.Info().Msg("Inicializando browser Rod...")
		u := launcher.New().Leakless(false).Headless(headless).NoSandbox(true).MustLaunch()
		browser := rod.New().ControlURL(u).MustConnect()
		// Isso configura o navegador para permitir downloads e salvar no path específico
		// sem abrir popup de confirmação.
		go func() {
			// É preciso rodar isso para cada Target (aba) nova ou na conexão principal
			// O MustSetDownloadBehavior envia o comando para o browser
			err := proto.BrowserSetDownloadBehavior{
				Behavior:         proto.BrowserSetDownloadBehaviorBehaviorAllow,
				DownloadPath:     absDownloadDir,
				EventsEnabled:    true, // Permite rastrear eventos de progresso
			}.Call(browser)
			
			if err != nil {
				log.Error().Err(err).Msg("Falha ao configurar diretório de download no Chrome")
			}
		}()
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