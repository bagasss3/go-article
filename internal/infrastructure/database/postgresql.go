package database

import (
	log "github.com/sirupsen/logrus"

	"database/sql"

	"github.com/bagasss3/go-article/internal/config"

	_ "github.com/lib/pq"
)

func NewDB() *sql.DB {
	db, err := sql.Open("postgres", config.DBDSN())
	if err != nil {
		log.WithField("dbDSN", config.DBDSN()).Fatal("Failed to connect:", err)
	}

	db.SetMaxIdleConns(config.MaxIdleConns())
	db.SetMaxOpenConns(config.MaxOpenConns())
	db.SetConnMaxLifetime(config.ConnMaxLifeTime())
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime())

	log.Info("Success connect database")
	return db
}
