package repository

import (
	"fmt"
	"github.com/botlorien/go-rpa-template/internal/domain"

	"gorm.io/gorm"
)

type RelatorioRepository struct {
	DB *gorm.DB
}

func NewRelatorioRepository(db *gorm.DB) *RelatorioRepository {
	// Garante que a tabela existe
	db.AutoMigrate(&domain.RelatorioPeso{})
	return &RelatorioRepository{DB: db}
}

func (r *RelatorioRepository) SaveBatch(dados []domain.RelatorioPeso) error {
	if len(dados) == 0 {
		return nil
	}

	// CreateInBatches é ideal para grandes volumes (ETL)
	result := r.DB.CreateInBatches(dados, 100)
	
	if result.Error != nil {
		return fmt.Errorf("erro ao salvar lote: %w", result.Error)
	}

	fmt.Printf("💾 [Repo] %d registros salvos com sucesso.\n", result.RowsAffected)
	return nil
}