package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/bagasss3/go-article/internal/config"

	"github.com/jpillora/backoff"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

func InitDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := connectWithRetry(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	configurePool(db)
	go checkConnection(ctx, db, time.NewTicker(config.DBPingInterval()))

	log.Info("Successfully connected to postgresql database")
	return db, nil
}

func connectWithRetry(ctx context.Context, dsn string) (*sql.DB, error) {
	b := &backoff.Backoff{
		Min:    200 * time.Millisecond,
		Max:    2 * time.Second,
		Factor: 2,
		Jitter: true,
	}

	maxAttempts := config.DBRetryAttempts()
	var conn *sql.DB
	var err error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			conn, err = sql.Open("postgres", dsn)
			if err == nil {
				err = conn.PingContext(ctx)
			}
			if err == nil {
				log.WithField("attempt", attempt).Info("Database ping successful")
				return conn, nil
			}

			log.WithFields(log.Fields{
				"attempt": attempt,
				"error":   err,
			}).Warn("Database connection failed, retrying...")

			time.Sleep(b.Duration())
		}
	}

	return nil, fmt.Errorf("could not connect to DB after %d attempts: %w", maxAttempts, err)
}

func checkConnection(ctx context.Context, db *sql.DB, ticker *time.Ticker) {
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Stopping database health check")
			return
		case <-ticker.C:
			if err := pingDB(db); err != nil {
				log.WithError(err).Error("Database ping failed")
			} else {
				recordConnectionPoolMetrics(db)
			}
		}
	}
}

func pingDB(db *sql.DB) error {
	if db == nil {
		return errors.New("database connection is nil")
	}
	ctxPing, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	start := time.Now()
	err := db.PingContext(ctxPing)
	if err != nil {
		return err
	}
	log.WithField("latency", time.Since(start)).Debug("Database ping successful")
	return nil
}

func configurePool(db *sql.DB) {
	db.SetMaxIdleConns(config.MaxIdleConns())
	db.SetMaxOpenConns(config.MaxOpenConns())
	db.SetConnMaxLifetime(config.ConnMaxLifeTime())
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime())
}

func recordConnectionPoolMetrics(db *sql.DB) {
	if db == nil {
		return
	}
	stats := db.Stats()
	log.WithFields(log.Fields{
		"maxOpen": stats.MaxOpenConnections,
		"open":    stats.OpenConnections,
		"inUse":   stats.InUse,
		"idle":    stats.Idle,
		"waits":   stats.WaitCount,
		"waitDur": stats.WaitDuration,
	}).Debug("DB connection pool stats")
}
