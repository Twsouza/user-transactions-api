package database

import (
	"log"
	"os"
	"time"
	"user-transactions/core/entities"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresDB struct {
	Db            *gorm.DB
	Logger        *log.Logger
	Dsn           string
	Debug         bool
	AutoMigrateDb bool
}

func (psql *PostgresDB) Connect() (*gorm.DB, error) {
	var err error

	config := &gorm.Config{
		PrepareStmt: true,
	}
	if psql.Debug {
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

	psql.Db, err = gorm.Open(postgres.Open(psql.Dsn), config)
	if err != nil {
		return nil, err
	}

	if psql.AutoMigrateDb {
		psql.Db.AutoMigrate(entities.Transaction{})
	}

	sqlDB, _ := psql.Db.DB()

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(20)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)

	return psql.Db, nil
}
