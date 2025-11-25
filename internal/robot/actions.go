package robot

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"os"

	"github.com/rs/zerolog/log"
)

var DefaultSSWHeaders = map[string]string{
	"Accept":             "*/*",
	"Accept-Language":    "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7",
	"Sec-Ch-Ua":          "\"Google Chrome\";v=\"141\", \"Not?A_Brand\";v=\"8\", \"Chromium\";v=\"141\"",
	"Sec-Ch-Ua-Mobile":   "?0",
	"Sec-Ch-Ua-Platform": "\"Windows\"",
	"Sec-Fetch-Dest":     "empty",
	"Sec-Fetch-Mode":     "cors",
	"Sec-Fetch-Site":     "same-origin",
	"User-Agent":         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36",
	"Content-Type":		  "application/x-www-form-urlencoded",
	"Origin":			  "https://targetUrl.com.br",
	"Referer":			  "https://targetUrl.com.br",
}

// Login
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
	
	page := s.Browser.MustPage("https://targetUrl.com.br/login")
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
	log.Info().Msg("Iniciando Login via HTTP (SSW)")

	targetURL := "https://targetUrl.com.br/login"

	// 1. Construção do Payload (Form Data)
	// Usamos url.Values para garantir que caracteres especiais na senha sejam escapados corretamente
	formData := url.Values{}
	formData.Set("user", user)
	formData.Set("pass", pass)
	formData.Set("dummy", fmt.Sprintf("%d", time.Now().UnixMilli()))

	// 2. Criação da Requisição
	// formData.Encode() transforma o mapa em string: "act=L&f1=..."
	log.Debug().Str("form_data", formData.Encode()).Msg("Payload do login HTTP")
	req, err := http.NewRequest("POST", targetURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}


	// 2. Aplica os headers padrão do robô (User-Agent, Sec-CH, etc)
	s.ApplyHeaders(req, DefaultSSWHeaders)

	// 4. Execução
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 5. Validação
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	log.Debug().Str("response_body", bodyString).Msg("Resposta do login HTTP")

	if resp.StatusCode != 200 {
		log.Error().Int("status", resp.StatusCode).Msg("Falha na requisição de login")
		return errors.New("status code inválido no login")
	}

	// Verificação básica de sucesso (SSW geralmente retorna o menu ou uma mensagem de erro no HTML)
	if strings.Contains(bodyString, "Login inv&aacute;lido") {
		return errors.New("credenciais inválidas ou erro no login targetUrl")
	}

	log.Info().Msg("Login HTTP realizado com sucesso (Sessão capturada no CookieJar)")
	return nil
}

// Outra ação: Baixar Relatório
func (s *Session) BaixarRelatorio(pathDownload string) (string, error) {
    // Lógica para navegar até o relatório e baixar

	log.Info().Msgf("Iniciando extração do relatório")
	endpoint := "https://targetUrl.com.br/download" // URL é a mesma para os dois passos


	step1Data := url.Values{}
	step1Data.Set("sequencia", "19")
	step1Data.Set("dummy", fmt.Sprintf("%d", time.Now().UnixMilli()))

	req1, err := http.NewRequest("POST", endpoint, strings.NewReader(step1Data.Encode()))
	if err != nil {
		return "", err
	}

	// Headers específicos do Passo 1
	req1.Header.Set("Referer", "https://targetUrl.com.br/menu") // Veio do Menu
	
	s.ApplyHeaders(req1, DefaultSSWHeaders) // Injeta User-Agent, etc

	resp1, err := s.HTTPClient.Do(req1)
	if err != nil {
		return "", fmt.Errorf("erro de conexão no passo 1: %v", err)
	}
	defer resp1.Body.Close()

	if resp1.StatusCode != 200 {
		return "", fmt.Errorf("erro ao inicializar tela 019. Status: %d", resp1.StatusCode)
	}

	resp1Bytes, err := io.ReadAll(resp1.Body)
	if err != nil {
		return "", err
	}
	resp1String := string(resp1Bytes)
	payloadDinamico := ParseFormInputs(resp1String)
	log.Debug().Interface("form_data", payloadDinamico).Msg("Form data extraído do passo 1")
	// ========================================================================
	// PASSO 2: FILTRO E DOWNLOAD (Gera o Excel)
	// ========================================================================
	log.Debug().Msg("Passo 2: Solicitando relatório Excel...")

	// Formatação de Datas Dinâmicas (ddMMyy)
	hoje := time.Now()
	amanha := hoje.Add(24 * time.Hour)
	layoutData := "020106" // ddMMyy
	
	// Sobrescreve/Adiciona campos no payload capturado
	payloadDinamico.Set("end_date", amanha.Format(layoutData))
	payloadDinamico.Set("start_date", hoje.Format(layoutData))
	payloadDinamico.Set("relatorio_excel", "s") 
	payloadDinamico.Set("dummy", fmt.Sprintf("%d", time.Now().UnixMilli()))
	req2, err := http.NewRequest("POST", endpoint, strings.NewReader(payloadDinamico.Encode()))
	if err != nil {
		return "", err
	}

	// Headers específicos do Passo 2
	req1.Header.Set("Referer", "https://targetUrl.com.br/download") // Referer agora é a própria tela
	
	s.ApplyHeaders(req2, DefaultSSWHeaders)

	resp2, err := s.HTTPClient.Do(req2)
	if err != nil {
		return "", fmt.Errorf("erro ao baixar relatório: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != 200 {
		return "", fmt.Errorf("servidor rejeitou o pedido do relatório. Status: %d", resp2.StatusCode)
	}


	finalBytes, err := io.ReadAll(resp2.Body)
	if err != nil {
		return "", err
	}

	if finalBytes == nil {
        return "", nil
    }

	// Simula o download do arquivo
	pathFile := pathDownload + string(os.PathSeparator) + "arquivo.csv"
    os.WriteFile(pathFile, finalBytes, 0644)

	log.Info().Int("bytes", len(finalBytes)).Str("arquivo", pathFile).Msg("Download concluído com sucesso")
	return pathFile, nil
}