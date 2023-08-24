package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/reonardoleis/rinha-backend-2023/utils"
)

type Database struct {
	Conn *sql.DB
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

	postgresDB, err := sql.Open(
		"postgres",
		fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host,
			port,
			user,
			password,
			name),
	)

	postgresDB.SetMaxIdleConns(utils.GetIntEnv("MAX_IDLE_CONNS", 30))
	postgresDB.SetMaxOpenConns(utils.GetIntEnv("MAX_OPEN_CONNS", 450))

	if err != nil {
		log.Println(err)
		return err
	}

	singleton = &Database{Conn: postgresDB}

	return nil
}

func Instance() (*Database, error) {
	if singleton == nil {
		return singleton, Connect()
	}
	return singleton, nil
}
