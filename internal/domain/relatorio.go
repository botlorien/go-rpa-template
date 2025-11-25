package domain

import (
	"time"
	"gorm.io/gorm"
)

type RelatorioPeso struct {
	gorm.Model
	Destino          string  `gorm:"index"` // Index ajuda na busca
	PesoCalculoTotal float64 `gorm:"type:decimal(15,2)"`
	DataProcessamento time.Time
}