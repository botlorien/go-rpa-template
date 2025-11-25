package processor

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/xuri/excelize/v2"
	"golang.org/x/text/encoding/charmap" // go get golang.org/x/text
	"golang.org/x/text/transform"
)

// LoadFile detecta a extensão e carrega os dados normalizados
func LoadFile(path string) (*DataFrame, error) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".xlsx", ".xlsm":
		return loadExcel(path)
	case ".csv", ".txt":
		return loadCSV(path)
	default:
		return nil, fmt.Errorf("formato de arquivo não suportado: %s", ext)
	}
}

func loadExcel(path string) (*DataFrame, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Pega a primeira aba
	sheet := f.GetSheetList()[0]
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, err
	}

	return parseRawRows(rows)
}

// loadCSV agora detecta o encoding automaticamente
func loadCSV(path string) (*DataFrame, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 1. Criamos um Buffered Reader para poder "espiar" (Peek) os bytes
	// sem consumir o arquivo.
	br := bufio.NewReader(f)

	// 2. Espiamos os primeiros 1024 bytes (ou menos se o arquivo for pequeno)
	sample, err := br.Peek(1024)
	if err != nil && err != io.EOF {
		return nil, err
	}

	// 3. Decisão do Encoding
	var reader io.Reader

	if isUTF8(sample) {
		// É UTF-8 (Padrão Moderno)
		reader = br
	} else {
		// Não é UTF-8, então assumimos Windows-1252 (Padrão Excel Brasil)
		// Transformamos o reader para converter ANSI -> UTF-8 on-the-fly
		reader = transform.NewReader(br, charmap.Windows1252.NewDecoder())
	}

	// 4. Cria o CSV Reader usando o reader correto
	csvReader := csv.NewReader(reader)
	csvReader.Comma = ';' // Tenta ponto-e-vírgula primeiro
	csvReader.LazyQuotes = true
	
	// Dica Extra: Detectar separador automaticamente
	// Se a primeira linha não tiver ';', tenta ','
	if strings.Count(string(sample), ";") < strings.Count(string(sample), ",") {
		csvReader.Comma = ','
	}

	rawRows, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("falha ao ler CSV: %v", err)
	}

	return parseRawRows(rawRows)
}

// isUTF8 verifica se os bytes são válidos na tabela UTF-8
func isUTF8(data []byte) bool {
	// utf8.Valid retorna true se TODOS os bytes forem válidos.
	// ASCII puro (sem acento) também é UTF-8 válido, então funciona para ambos.
	// Se tiver um "ç" salvo em ANSI, isso aqui vai retornar false.
	return utf8.Valid(data)
}

// parseRawRows transforma matriz de string em nosso DataFrame (Map)
// E aplica o TrimSpace (Strip) em TUDO
func parseRawRows(rawRows [][]string) (*DataFrame, error) {
	if len(rawRows) < 1 {
		return nil, fmt.Errorf("arquivo vazio")
	}

	df := NewDataFrame()
	
	// 1. Processa Cabeçalho (Limpa e padroniza)
	headerRaw := rawRows[0]
	for _, h := range headerRaw {
		// Remove espaços e força minúsculo para facilitar acesso: " Valor Total " -> "valor_total"
		cleanHeader := strings.TrimSpace(h)
		// cleanHeader = strings.ToLower(cleanHeader) // Opcional: forçar lowercase
		df.Headers = append(df.Headers, cleanHeader)
	}

	// 2. Processa Linhas
	for _, rowSlice := range rawRows[1:] {
		rowMap := make(Row)
		for i, cell := range rowSlice {
			if i >= len(df.Headers) {
				continue // Ignora colunas extras sem cabeçalho
			}
			
			// AQUI ACONTECE O STRIP (PYTHON .strip())
			cleanVal := strings.TrimSpace(cell)
			
			rowMap[df.Headers[i]] = cleanVal
		}
		df.Rows = append(df.Rows, rowMap)
	}

	return df, nil
}