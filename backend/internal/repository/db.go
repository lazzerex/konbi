package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// db manager handles database connection and initialization
type DBManager struct {
	db     *sql.DB
	logger *logrus.Logger
}

// create new db manager
func NewDBManager(logger *logrus.Logger) *DBManager {
	return &DBManager{
		logger: logger,
	}
}

// initialize database connection
func (m *DBManager) Initialize(ctx context.Context, databaseURL string) (*sql.DB, error) {
	var err error

	if databaseURL != "" {
		// use postgresql (neon)
		m.logger.Info("connecting to postgresql database")
		m.db, err = sql.Open("postgres", databaseURL)
		if err != nil {
			m.logger.WithError(err).Error("failed to connect to postgresql")
			return nil, fmt.Errorf("failed to connect to postgresql: %w", err)
		}

		// test connection
		if err = m.db.PingContext(ctx); err != nil {
			m.logger.WithError(err).Error("failed to ping postgresql")
			return nil, fmt.Errorf("failed to ping postgresql: %w", err)
		}

		m.logger.Info("successfully connected to postgresql")
	} else {
		// use sqlite (development)
		m.logger.Info("using sqlite database")
		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = "./konbi.db"
		}

		m.db, err = sql.Open("sqlite3", dbPath)
		if err != nil {
			m.logger.WithError(err).Error("failed to open sqlite")
			return nil, fmt.Errorf("failed to open sqlite: %w", err)
		}

		m.logger.WithField("db_path", dbPath).Info("successfully opened sqlite database")
	}

	return m.db, nil
}

// run migrations creates tables and indexes
func (m *DBManager) RunMigrations(ctx context.Context) error {
	m.logger.Info("running database migrations")

	// determine if using postgresql or sqlite
	isPostgres := os.Getenv("DATABASE_URL") != ""

	var schema string
	if isPostgres {
		schema = `
		CREATE TABLE IF NOT EXISTS content (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			title TEXT,
			filename TEXT,
			filepath TEXT,
			filesize BIGINT,
			content TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NOT NULL,
			view_count INTEGER DEFAULT 0,
			deleted_at TIMESTAMP
		);
		`
		
		// add deleted_at column if it doesn't exist (for existing databases)
		m.db.ExecContext(ctx, "ALTER TABLE content ADD COLUMN deleted_at TIMESTAMP")
		
		schema += `
		CREATE INDEX IF NOT EXISTS idx_content_expires_at ON content(expires_at);
		CREATE INDEX IF NOT EXISTS idx_content_deleted_at ON content(deleted_at);
		CREATE INDEX IF NOT EXISTS idx_content_type ON content(type);
		CREATE INDEX IF NOT EXISTS idx_content_created_at ON content(created_at DESC);

		CREATE TABLE IF NOT EXISTS shortened_urls (
			id BIGSERIAL PRIMARY KEY,
			short_code TEXT UNIQUE NOT NULL,
			original_url TEXT NOT NULL,
			custom_alias TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			expires_at TIMESTAMP,
			click_count INTEGER DEFAULT 0,
			deleted_at TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_shortened_urls_short_code ON shortened_urls(short_code);
		CREATE INDEX IF NOT EXISTS idx_shortened_urls_expires_at ON shortened_urls(expires_at);
		CREATE INDEX IF NOT EXISTS idx_shortened_urls_deleted_at ON shortened_urls(deleted_at);

		CREATE TABLE IF NOT EXISTS url_clicks (
			id BIGSERIAL PRIMARY KEY,
			url_id BIGINT NOT NULL REFERENCES shortened_urls(id) ON DELETE CASCADE,
			clicked_at TIMESTAMP DEFAULT NOW(),
			ip_address TEXT,
			user_agent TEXT,
			referrer TEXT
		);

		CREATE INDEX IF NOT EXISTS idx_url_clicks_url_id ON url_clicks(url_id);
		CREATE INDEX IF NOT EXISTS idx_url_clicks_clicked_at ON url_clicks(clicked_at DESC);
		`
	} else {
		schema = `
		CREATE TABLE IF NOT EXISTS content (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			title TEXT,
			filename TEXT,
			filepath TEXT,
			filesize INTEGER,
			content TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME NOT NULL,
			view_count INTEGER DEFAULT 0,
			deleted_at DATETIME
		);
		`
		
		// add deleted_at column if it doesn't exist (for existing databases)
		m.db.ExecContext(ctx, "ALTER TABLE content ADD COLUMN deleted_at DATETIME")
		
		schema += `
		CREATE INDEX IF NOT EXISTS idx_content_expires_at ON content(expires_at);
		CREATE INDEX IF NOT EXISTS idx_content_deleted_at ON content(deleted_at);
		CREATE INDEX IF NOT EXISTS idx_content_type ON content(type);
		CREATE INDEX IF NOT EXISTS idx_content_created_at ON content(created_at DESC);

		CREATE TABLE IF NOT EXISTS shortened_urls (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			short_code TEXT UNIQUE NOT NULL,
			original_url TEXT NOT NULL,
			custom_alias TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME,
			click_count INTEGER DEFAULT 0,
			deleted_at DATETIME
		);

		CREATE INDEX IF NOT EXISTS idx_shortened_urls_short_code ON shortened_urls(short_code);
		CREATE INDEX IF NOT EXISTS idx_shortened_urls_expires_at ON shortened_urls(expires_at);
		CREATE INDEX IF NOT EXISTS idx_shortened_urls_deleted_at ON shortened_urls(deleted_at);

		CREATE TABLE IF NOT EXISTS url_clicks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			url_id INTEGER NOT NULL REFERENCES shortened_urls(id) ON DELETE CASCADE,
			clicked_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			ip_address TEXT,
			user_agent TEXT,
			referrer TEXT
		);

		CREATE INDEX IF NOT EXISTS idx_url_clicks_url_id ON url_clicks(url_id);
		CREATE INDEX IF NOT EXISTS idx_url_clicks_clicked_at ON url_clicks(clicked_at DESC);
		`
	}

	// split and execute each statement
	statements := strings.Split(schema, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if _, err := m.db.ExecContext(ctx, stmt); err != nil {
			m.logger.WithError(err).WithField("statement", stmt).Error("failed to execute migration")
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	m.logger.Info("database migrations completed successfully")
	return nil
}

// configure connection pool settings
func (m *DBManager) ConfigurePool(maxConns, maxIdle int, maxLifetime int) {
	m.db.SetMaxOpenConns(maxConns)
	m.db.SetMaxIdleConns(maxIdle)
	m.db.SetConnMaxLifetime(0) // connections don't expire

	m.logger.WithFields(logrus.Fields{
		"max_open_conns": maxConns,
		"max_idle_conns": maxIdle,
	}).Info("database connection pool configured")
}

// close database connection
func (m *DBManager) Close() error {
	if m.db != nil {
		m.logger.Info("closing database connection")
		return m.db.Close()
	}
	return nil
}

// get database instance
func (m *DBManager) GetDB() *sql.DB {
	return m.db
}
