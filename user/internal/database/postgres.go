package database

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"time"
)

func ConnectPostgres(env string, ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("pgx", env)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Minute * 5)

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("failed to ping DB: %v", err)
		return nil, err
	}

	log.Println("✅ PostgreSQL connected")
	return db, nil
}
