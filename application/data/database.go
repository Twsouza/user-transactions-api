package data

import (
	"log"
	"os"
	"time"
	"user-transactions/core"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	Db            *gorm.DB
	Logger        *log.Logger
	Dsn           string
	Debug         bool
	AutoMigrateDb bool
}

func (d *Database) Connect() (*gorm.DB, error) {
	var err error

	config := &gorm.Config{
		PrepareStmt: true,
	}
	if d.Debug {
		newLogger := logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Info,
				Colorful:                  true,
				IgnoreRecordNotFoundError: false,
				ParameterizedQueries:      true,
			},
		)

		config.Logger = newLogger
	}

	d.Db, err = gorm.Open(postgres.Open(d.Dsn), config)
	if err != nil {
		return nil, err
	}

	if d.AutoMigrateDb {
		d.Db.AutoMigrate(core.Transaction{})
	}

	return d.Db, nil
}
