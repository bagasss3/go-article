package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bagasss3/go-article/internal/config"

	"github.com/jpillora/backoff"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	// PostgresDB represents gorm DB
	PostgresDB *gorm.DB

	ErrNilDatabase = errors.New("database connection is nil")

	// ctx is the background context for database operations
	ctx    context.Context
	cancel context.CancelFunc
)

type DBOptions struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func InitDB() {
	ctx, cancel = context.WithCancel(context.Background())

	opts := getDBOptions()
	conn, err := openDB(config.DBDSN(), opts)
	if err != nil {
		log.WithFields(log.Fields{
			"dbDSN": config.DBDSN(),
			"error": err,
		}).Fatal("Failed to connect to database")
	}

	PostgresDB = conn

	// Start connection health check
	go checkConnection(ctx, time.NewTicker(config.DBPingInterval()))

	log.Info("Successfully connected to PostgreSQL database")
}

// CloseDB properly closes the database connection
func CloseDB() {
	if cancel != nil {
		cancel()
	}

	if PostgresDB != nil {
		db, err := PostgresDB.DB()
		if err != nil {
			log.WithError(err).Error("Error getting database instance while closing")
			return
		}
		if err := db.Close(); err != nil {
			log.WithError(err).Error("Error closing database connection")
		}
	}
}

// checkConnection periodically checks the database connection health
func checkConnection(ctx context.Context, ticker *time.Ticker) {
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Stopping database health check")
			return
		case <-ticker.C:
			if err := pingDB(); err != nil {
				log.WithError(err).Error("Database ping failed")
				if err := reconnectPostgresDBConn(ctx); err != nil {
					log.WithError(err).Error("Failed to reconnect to database")
				}
			}
			recordConnectionPoolMetrics()
		}
	}
}

// pingDB checks if the database connection is alive
func pingDB() error {
	if PostgresDB == nil {
		return ErrNilDatabase
	}

	db, err := PostgresDB.DB()
	if err != nil {
		return fmt.Errorf("getting database instance: %w", err)
	}

	start := time.Now()
	err = db.Ping()
	pingDuration := time.Since(start)

	if err != nil {
		return fmt.Errorf("pinging database: %w", err)
	}

	log.WithField("latency", pingDuration).Debug("Database ping successful")
	return nil
}

// reconnectPostgresDBConn attempts to reconnect to the database with exponential backoff
func reconnectPostgresDBConn(ctx context.Context) error {
	b := backoff.Backoff{
		Factor: 2,
		Jitter: true,
		Min:    100 * time.Millisecond,
		Max:    1 * time.Second,
	}
	maxAttempts := config.DBRetryAttempts()

	for attempt := 0; attempt < maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			log.WithField("attempt", attempt+1).Info("Attempting database reconnection")

			conn, err := openDB(config.DBDSN(), getDBOptions())
			if err == nil && conn != nil {
				PostgresDB = conn
				log.Info("Successfully reconnected to database")
				return nil
			}

			if err != nil {
				log.WithError(err).Error("Reconnection attempt failed")
			}

			time.Sleep(b.Duration())
			b.Reset()
		}
	}

	return fmt.Errorf("failed to reconnect after %d attempts", maxAttempts)
}

// openDB creates a new database connection with the given options
func openDB(dsn string, opts DBOptions) (*gorm.DB, error) {
	dialect := postgres.Open(dsn)
	db, err := gorm.Open(dialect, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("getting database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(opts.MaxIdleConns)
	sqlDB.SetMaxOpenConns(opts.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(opts.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(opts.ConnMaxIdleTime)

	return db, nil
}

// getDBOptions returns the database options from config
func getDBOptions() DBOptions {
	return DBOptions{
		MaxIdleConns:    config.MaxIdleConns(),
		MaxOpenConns:    config.MaxOpenConns(),
		ConnMaxLifetime: config.ConnMaxLifeTime(),
		ConnMaxIdleTime: config.ConnMaxIdleTime(),
	}
}

// recordConnectionPoolMetrics records database connection pool metrics
func recordConnectionPoolMetrics() {
	if PostgresDB == nil {
		return
	}

	db, err := PostgresDB.DB()
	if err != nil {
		log.WithError(err).Error("Failed to get database stats")
		return
	}

	stats := db.Stats()
	log.WithFields(log.Fields{
		"maxOpenConnections": stats.MaxOpenConnections,
		"openConnections":    stats.OpenConnections,
		"inUse":              stats.InUse,
		"idle":               stats.Idle,
		"waitCount":          stats.WaitCount,
		"waitDuration":       stats.WaitDuration,
	}).Debug("Database connection pool stats")
}
