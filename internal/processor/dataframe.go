package processor

import (
	"os"
	"bytes"
	"encoding/csv"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/xuri/excelize/v2"
)

// Lista de layouts suportados, em ordem de prioridade.
// A ordem é CRUCIAL para resolver ambiguidades (ex: 01/02).
var DateLayouts = []string{
	// --- 1. Formatos ISO/Internacional (Inequívocos - Ano na frente) ---
	time.RFC3339,                 // "2006-01-02T15:04:05Z07:00"
	"2006-01-02 15:04:05.000000", // SQL Timestamp com microsegundos
	"2006-01-02 15:04:05",        // SQL/Excel Padrão
	"2006-01-02 15:04",           // Sem segundos
	"2006-01-02",                 // Data pura

	// --- 2. Formatos Brasileiros (Dia na frente) ---
	"02/01/2006 15:04:05", // BR Completo
	"02/01/2006 15:04",    // BR Hora Simples
	"02/01/2006",          // BR Padrão (Mais comum)
	"02-01-2006",          // Traços
	"02.01.2006",          // Pontos

	// --- 3. Formatos Legados/Compactos (Bancos) ---
	"020106",   // ddmmyy (Clássico)
	"02012006", // ddmmyyyy
	"20060102", // yyyymmdd

	// --- 4. Formatos Americanos (Mês na frente) ---
	// CUIDADO: 01/05 será lido como 1º de Maio (BR) antes de chegar aqui.
	// Este bloco só captura datas impossíveis no BR (ex: 05/31/2025).
	"01/02/2006 15:04:05 PM", // Com AM/PM
	"01/02/2006 15:04:05",    // 24h
	"01/02/2006",             // US Padrão
}

// Row representa uma linha do relatório (Coluna -> Valor)
type Row map[string]string

// DataFrame é a nossa estrutura em memória
type DataFrame struct {
	Headers []string
	Rows    []Row
}

// NewDataFrame cria uma estrutura vazia
func NewDataFrame() *DataFrame {
	return &DataFrame{
		Headers: []string{},
		Rows:    []Row{},
	}
}

// Count retorna o número de linhas
func (df *DataFrame) Count() int {
	return len(df.Rows)
}

// Filter aceita uma função lambda para filtrar linhas (igual df[df['col'] > 0])
func (df *DataFrame) Filter(condition func(row Row) bool) *DataFrame {
	filtered := NewDataFrame()
	filtered.Headers = df.Headers
	for _, row := range df.Rows {
		if condition(row) {
			filtered.Rows = append(filtered.Rows, row)
		}
	}
	return filtered
}

// --- Helpers de Conversão (Padrão Brasil) ---

// GetFloat converte "1.234,56" para float64
func (r Row) GetFloatBR(col string) float64 {
	val, ok := r[col]
	if !ok || val == "" {
		return 0.0
	}
	// Remove pontos de milhar e troca vírgula decimal por ponto
	clean := strings.ReplaceAll(val, ".", "")
	clean = strings.ReplaceAll(clean, ",", ".")
	
	f, _ := strconv.ParseFloat(clean, 64)
	return f
}


func (r Row) GetFloatStd(col string) float64 {
	val, ok := r[col]
	if !ok || val == "" {
		return 0.0
	}
	// ParseFloat nativo do Go já entende o ponto como decimal corretamente
	f, _ := strconv.ParseFloat(val, 64)
	return f
}

// GetInt converte string para int
func (r Row) GetInt(col string) int {
	val, ok := r[col]
	if !ok || val == "" {
		return 0
	}
	i, _ := strconv.Atoi(val)
	return i
}

// GetDate tenta converter a string em data testando múltiplos formatos
// Se layout for informado, ele tenta APENAS aquele layout.
// Se layout for vazio, ele tenta adivinhar usando a lista DateLayouts.
func (r Row) GetDate(col string, specificLayout string) time.Time {
	val, ok := r[col]
	if !ok || val == "" {
		return time.Time{}
	}

	// 1. Se o usuário passou um layout específico, respeitamos (Performance e Precisão)
	if specificLayout != "" {
		t, err := time.Parse(specificLayout, val)
		if err == nil {
			return t
		}
		// Se falhar o específico, podemos retornar zero ou tentar o fallback.
		// Aqui retornamos zero para indicar erro de contrato.
		return time.Time{}
	}

	// 2. Tentativa Automática (Brute Force inteligente)
	for _, layout := range DateLayouts {
		// Dica: time.Parse é estrito. Se sobrar caracter, ele falha.
		// Isso é bom para evitar falso positivo.
		t, err := time.Parse(layout, val)
		if err == nil {
			return t
		}
	}

	// 3. Casos Especiais (Excel Serial Date)
	// Às vezes o Excelize retorna "45250" em vez de "25/11/2023"
	if excelSerial, err := strconv.ParseFloat(val, 64); err == nil {
		// Data base do Excel (quase sempre é 30/Dez/1899)
		baseDate := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
		// Converte dias para Nanosegundos e soma
		daysInNano := time.Duration(excelSerial * 24 * float64(time.Hour))
		return baseDate.Add(daysInNano)
	}

	return time.Time{} // Falhou tudo
}

// String implementa a interface fmt.Stringer.
// Isso permite usar fmt.Println(df) e ter uma saída formatada igual Pandas.
func (df *DataFrame) String() string {
	if df == nil || len(df.Rows) == 0 {
		return "DataFrame Vazio []"
	}

	var buf bytes.Buffer
	
	// Configura o tabwriter:
	// minwidth=0, tabwidth=0, padding=2, padchar=' ', flags=0
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	// 1. Escreve o Cabeçalho
	// Junta os headers com \t (tabulação) para o writer alinhar
	fmt.Fprintln(w, strings.Join(df.Headers, "\t"))
	
	// (Opcional) Linha separadora visual, mas o Pandas não usa muito
	// fmt.Fprintln(w, strings.Repeat("-\t", len(df.Headers)))

	// 2. Escreve as Linhas (Limitando a 20 para não poluir o terminal, igual df.head(20))
	limit := 20
	count := 0
	
	for _, row := range df.Rows {
		if count >= limit {
			fmt.Fprintf(w, "... mais %d linhas ...\n", len(df.Rows)-limit)
			break
		}

		var line []string
		for _, colName := range df.Headers {
			val := row[colName]
			// Se for muito grande, corta para não quebrar o layout
			if len(val) > 50 {
				val = val[:47] + "..."
			}
			line = append(line, val)
		}
		fmt.Fprintln(w, strings.Join(line, "\t"))
		count++
	}

	// 3. Resumo final (igual Pandas: [5 rows x 3 columns])
	fmt.Fprintf(w, "\n[%d rows x %d columns]\n", len(df.Rows), len(df.Headers))

	w.Flush()
	return buf.String()
}

// Head é um helper para imprimir manualmente apenas o começo (se quiser variar o limite)
func (df *DataFrame) Head(n int) {
	fmt.Printf("--- Head (%d) ---\n", n)
	// Reutiliza a lógica do String() mas poderia ser customizado
	// Aqui só um print simples para debug rápido
	fmt.Println(df) 
}

// FromMap cria um DataFrame a partir de um mapa de colunas.
// Ex: {"DESTINO": ["MTZ", "SP"], "VALOR": [10.5, 20.0]}
func FromMap(data map[string][]any) *DataFrame {
	df := NewDataFrame()
	if len(data) == 0 {
		return df
	}

	// 1. Define os Headers e descobre o número de linhas
	maxRows := 0
	for col, values := range data {
		df.Headers = append(df.Headers, col)
		if len(values) > maxRows {
			maxRows = len(values)
		}
	}
	
	// Ordena headers para consistência visual (opcional)
	sort.Strings(df.Headers)

	// 2. Constrói as linhas pivotando as colunas
	for i := 0; i < maxRows; i++ {
		row := make(Row)
		for _, col := range df.Headers {
			values := data[col]
			valStr := ""
			
			// Proteção contra slices de tamanhos diferentes
			if i < len(values) {
				// Converte any para string de forma segura
				val := values[i]
				switch v := val.(type) {
				case float64:
					valStr = fmt.Sprintf("%.2f", v) // Formata float bonito
				case int:
					valStr = fmt.Sprintf("%d", v)
				case string:
					valStr = v
				default:
					valStr = fmt.Sprintf("%v", v)
				}
			}
			row[col] = valStr
		}
		df.Rows = append(df.Rows, row)
	}

	return df
}

// SortBy ordena o DataFrame baseado em uma coluna.
// Tenta ordenar numericamente se possível, senão usa ordem alfabética.
func (df *DataFrame) SortBy(col string, ascending bool) {
	sort.SliceStable(df.Rows, func(i, j int) bool {
		valI := df.Rows[i][col]
		valJ := df.Rows[j][col]

		// Tenta converter para float para ordenação numérica
		numI, errI := strconv.ParseFloat(valI, 64)
		numJ, errJ := strconv.ParseFloat(valJ, 64)

		if errI == nil && errJ == nil {
			if ascending {
				return numI < numJ
			}
			return numI > numJ
		}

		// Fallback para string (alfabética)
		if ascending {
			return valI < valJ
		}
		return valI > valJ
	})
}

// Export salva o DataFrame em CSV ou XLSX dependendo da extensão
func (df *DataFrame) Export(path string) error {
	if strings.HasSuffix(strings.ToLower(path), ".xlsx") {
		return df.toXLSX(path)
	}
	return df.toCSV(path)
}

// Implementação interna CSV
func (df *DataFrame) toCSV(path string) error {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.Comma = ';' // Padrão Excel Brasil

	// Escreve Header
	writer.Write(df.Headers)

	// Escreve Linhas
	for _, row := range df.Rows {
		record := make([]string, len(df.Headers))
		for i, col := range df.Headers {
			record[i] = row[col]
		}
		writer.Write(record)
	}
	writer.Flush()

	// Salva no disco (Aqui você poderia usar os.WriteFile com path absoluto)
	return os.WriteFile(path, buf.Bytes(), 0644)
}

// Implementação interna XLSX
func (df *DataFrame) toXLSX(path string) error {
	f := excelize.NewFile()
	sheet := "Sheet1"
	
	// Escreve Header
	for i, h := range df.Headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Escreve Linhas
	for rIdx, row := range df.Rows {
		for cIdx, col := range df.Headers {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, rIdx+2) // Linha começa em 2
			
			// Tenta salvar números como números reais no Excel
			val := row[col]
			if num, err := strconv.ParseFloat(val, 64); err == nil {
				f.SetCellValue(sheet, cell, num)
			} else {
				f.SetCellValue(sheet, cell, val)
			}
		}
	}

	return f.SaveAs(path)
}