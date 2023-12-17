package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"os"
)

var (
	dbHost = os.Getenv("DB_HOST")
	dbPort = os.Getenv("DB_PORT")
	dbUser = os.Getenv("DB_USER")
	dbPass = os.Getenv("DB_PASS")
	dbName = os.Getenv("DB_NAME")
)

func main() {
	endpoint := fmt.Sprintf("tcp(%s:%s)", dbHost, dbPort)
	dsn := fmt.Sprintf("%s:%s@%s/%s?charset=utf8&parseTime=true&multiStatements=true",
		dbUser,
		dbPass,
		endpoint,
		dbName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println(err)
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		fmt.Println(err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://files",
		"mysql",
		driver,
	)
	if err != nil {
		fmt.Println(err)
	}

	err = m.Up()
	if err != nil {
		fmt.Println(err)
	}
}
