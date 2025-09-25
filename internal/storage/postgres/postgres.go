// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package postgres // import "miniflux.app/v2/internal/storage/postgres"

import (
	"context"
	"database/sql"
	"time"
)

// NewConnectionPool configures the database connection pool.
func NewConnectionPool(dsn string, minConnections, maxConnections int, connectionLifetime time.Duration) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxConnections)
	db.SetMaxIdleConns(minConnections)
	db.SetConnMaxLifetime(connectionLifetime)

	return db, nil
}

// Postgres handles all operations related to the database.
type Postgres struct {
	db *sql.DB
}

// New returns a new Storage.
func New(db *sql.DB) *Postgres {
	return &Postgres{db}
}

// DatabaseVersion returns the version of the database which is in use.
func (p *Postgres) DatabaseVersion() string {
	var dbVersion string
	err := p.db.QueryRow(`SELECT current_setting('server_version')`).Scan(&dbVersion)
	if err != nil {
		return err.Error()
	}

	return dbVersion
}

// Ping checks if the database connection works.
func (p *Postgres) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.db.PingContext(ctx)
}

// DB returns the underlying database connection.
func (p *Postgres) DB() *sql.DB {
	return p.db
}

// DBStats returns database statistics.
func (p *Postgres) DBStats() sql.DBStats {
	return p.db.Stats()
}

// DBSize returns how much size the database is using in a pretty way.
func (p *Postgres) DBSize() (string, error) {
	var size string

	err := p.db.QueryRow("SELECT pg_size_pretty(pg_database_size(current_database()))").Scan(&size)
	if err != nil {
		return "", err
	}

	return size, nil
}
