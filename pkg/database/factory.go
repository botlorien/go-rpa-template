package database

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewConnection cria uma conexão baseada no driver escolhido
func NewConnection(driver, dsn string) (*gorm.DB, error) {
	var dialect gorm.Dialector

	switch driver {
	case "postgres":
		dialect = postgres.Open(dsn)
	case "mysql":
		dialect = mysql.Open(dsn)
	case "sqlite":
		dialect = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("driver de banco de dados não suportado: %s", driver)
	}

	// Configuração GORM global
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(dialect, config)
	if err != nil {
		return nil, err
	}

	return db, nil
}