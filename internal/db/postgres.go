package db

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Connect() {
	dsn := os.Getenv("DATABASE_URL")
	config, err := pgxpool.ParseConfig(dsn)

	if err != nil {
		log.Fatal("unable to parse DATABASE_URL:", err)
	}

	config.MaxConns = 10
	config.MaxConnLifetime = time.Hour

	Pool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatal("unable to create connection pool:", err)
	}
	log.Println("Connected to Postgres")
}

func Close() {
	Pool.Close()
}
