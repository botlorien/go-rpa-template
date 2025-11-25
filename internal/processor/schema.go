package processor

import (
	"fmt"
	"strings"
	"unicode"
)

// SuggestStruct analisa o DataFrame e imprime uma sugestão de Struct Go compatível com GORM
func (df *DataFrame) SuggestStruct(structName string) string {
	if len(df.Headers) == 0 {
		return "// DataFrame vazio, impossível sugerir struct"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("type %s struct {\n", structName))
	sb.WriteString("\tgorm.Model\n") // Adiciona ID, CreatedAt, UpdatedAt, DeletedAt

	for _, col := range df.Headers {
		fieldName := toPascalCase(col)
		goType := "string" // Default
		gormTag := ""

		// Amostragem: Analisa as primeiras 20 linhas para adivinhar o tipo
		isFloat := true
		isDate := true
		hasData := false

		limit := 20
		if len(df.Rows) < limit {
			limit = len(df.Rows)
		}

		for i := 0; i < limit; i++ {
			val := df.Rows[i][col]
			if val == "" {
				continue
			}
			hasData = true

			// Teste Float
			if isFloat {
				if df.Rows[i].GetFloatBR(col) == 0 && val != "0" && val != "0,00" {
					// Se GetFloat retornou 0 mas o texto não é zero, falhou conversão
					isFloat = false
				}
			}

			// Teste Date
			if isDate {
				t := df.Rows[i].GetDate(col, "")
				if t.IsZero() {
					isDate = false
				}
			}
		}

		// Decisão do Tipo
		if hasData {
			if isDate {
				goType = "time.Time"
				gormTag = "`gorm:\"type:date\"`" // Ou datetime dependendo do banco
			} else if isFloat {
				goType = "float64"
				gormTag = "`gorm:\"type:decimal(10,2)\"`"
			}
		}

		// Monta a linha: NomeCampo Tipo `tag`
		// Ex: ValorFrete float64 `gorm:"type:decimal(10,2)"`
		if gormTag != "" {
			sb.WriteString(fmt.Sprintf("\t%s %s %s\n", fieldName, goType, gormTag))
		} else {
			sb.WriteString(fmt.Sprintf("\t%s %s\n", fieldName, goType))
		}
	}

	sb.WriteString("}")
	return sb.String()
}

// Helper para transformar "VALOR DO FRETE" em "ValorDoFrete"
func toPascalCase(s string) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return ' '
	}, s)
	
	parts := strings.Fields(s)
	for i, p := range parts {
		parts[i] = strings.Title(strings.ToLower(p))
	}
	return strings.Join(parts, "")
}