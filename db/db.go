package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type Database struct {
	Conn *pgxpool.Pool
}

var singleton *Database

func Connect() error {
	var (
		name     = os.Getenv("DB_NAME")
		host     = os.Getenv("DB_HOST")
		port     = os.Getenv("DB_PORT")
		user     = os.Getenv("DB_USER")
		password = os.Getenv("DB_PASSWORD")
	)

	connString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host,
		port,
		user,
		password,
		name)

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Println(err)
		return err
	}

	config.MinConns = 4000
	config.MaxConns = 4500

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Println(err)
		return err
	}

	singleton = &Database{Conn: pool}

	return nil
}

func Instance() (*Database, error) {
	if singleton == nil {
		return singleton, Connect()
	}
	return singleton, nil
}
