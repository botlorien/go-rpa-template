package robot

import (
	"net/url"
	"regexp"
	"strings"
)

// ParseFormInputs varre o HTML em busca de tags <input> e retorna um mapa de dados.
// Prioridade: name > id.
func ParseFormInputs(html string) url.Values {
	formData := url.Values{}

	// Regex para encontrar a tag <input ... > inteira
	// (?i) = Case insensitive
	reTag := regexp.MustCompile(`(?i)<input\s+[^>]*>`)
	
	// Regex para extrair atributos específicos
	reName := regexp.MustCompile(`(?i)\bname=["']?([^"'\s>]+)["']?`)
	reID := regexp.MustCompile(`(?i)\bid=["']?([^"'\s>]+)["']?`)
	reValue := regexp.MustCompile(`(?i)\bvalue=["']?([^"']*)["']?`)

	// Encontra todas as tags input
	inputs := reTag.FindAllString(html, -1)

	for _, inputTag := range inputs {
		// Pula inputs do tipo submit ou button se necessário (opcional)
		// if strings.Contains(strings.ToLower(inputTag), "type=\"submit\"") { continue }

		// 1. Tenta extrair o Name
		key := ""
		matchName := reName.FindStringSubmatch(inputTag)
		if len(matchName) > 1 {
			key = matchName[1]
		}

		// 2. Se não tem Name, tenta o ID (conforme seu requisito)
		if key == "" {
			matchID := reID.FindStringSubmatch(inputTag)
			if len(matchID) > 1 {
				key = matchID[1]
			}
		}

		// Se não achou nem name nem id, ignora
		if key == "" {
			continue
		}

		// 3. Extrai o Value
		val := ""
		matchValue := reValue.FindStringSubmatch(inputTag)
		if len(matchValue) > 1 {
			val = matchValue[1]
		}

		// Decodifica HTML entities se necessário (ex: &quot; -> ")
		// O SSW as vezes manda value="ABRIR('X')" sem encodar, mas é bom prevenir
		val = strings.TrimSpace(val)

		formData.Set(key, val)
	}

	return formData
}