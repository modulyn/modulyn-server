package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

// LoggerTx wraps a standard sql.Tx to add logging
type LoggerTx struct {
	*sql.Tx
	id string // Unique ID for the transaction for logging
}

func (ldb *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*LoggerTx, error) {
	tx, err := ldb.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	txID := fmt.Sprintf("tx-%s", ctx.Value(CorrelationKey))
	log.Printf("Transaction %s: Started", txID)
	return &LoggerTx{Tx: tx, id: txID}, nil
}

func (ltx *LoggerTx) Commit() error {
	log.Printf("Transaction %s: Committing", ltx.id)
	return ltx.Tx.Commit()
}

func (ltx *LoggerTx) Rollback() error {
	log.Printf("Transaction %s: Rolling back", ltx.id)
	return ltx.Tx.Rollback()
}

func (ltx *LoggerTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if EnableSqlLogging {
		log.Printf("Transaction %s: Executing query: %s", ltx.id, interpolateSQL(query, args...))
	} else {
		log.Printf("Transaction %s: Executing query", ltx.id)
	}
	return ltx.Tx.ExecContext(ctx, query, args...)
}

func (ltx *LoggerTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if EnableSqlLogging {
		log.Printf("Transaction %s: Querying: %s", ltx.id, interpolateSQL(query, args...))
	} else {
		log.Printf("Transaction %s: Querying", ltx.id)
	}
	return ltx.Tx.QueryContext(ctx, query, args...)
}
