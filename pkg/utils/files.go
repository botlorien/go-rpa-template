package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// EmptyDirectory remove todos os arquivos e subpastas de um diretório,
// mas mantém o diretório raiz.
func EmptyDirectory(dir string) error {
	// 1. Lê o conteúdo da pasta
	files, err := os.ReadDir(dir)
	if err != nil {
		// Se a pasta nem existe, tecnicamente ela já está "vazia"
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// 2. Itera sobre cada item e deleta
	for _, file := range files {
		path := filepath.Join(dir, file.Name())
		
		// RemoveAll apaga recursivamente (arquivos e subpastas)
		err := os.RemoveAll(path)
		if err != nil {
			return fmt.Errorf("falha ao deletar %s: %w", path, err)
		}
	}

	return nil
}