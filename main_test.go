package main

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"sample-server/ent"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

var client *ent.Client
var ctx context.Context

func TestMain(m *testing.M) {
	ctx = context.Background()
	db := openDBConnection()
	defer db.Close()

	createTestDatabase(db)
	client = setupEntClient()
	defer client.Close()

	migrateSchema(client)

	exitVal := m.Run()

	os.Exit(exitVal)
}

func openDBConnection() *sql.DB {
	// TODO refactoring
	db, err := sql.Open("mysql", "root:password@tcp(localhost:3306)/")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	return db
}

func createTestDatabase(db *sql.DB) {
	// TODO refactoring
	_, err := db.Exec("CREATE DATABASE IF NOT EXISTS sample_test")
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
}

func setupEntClient() *ent.Client {
	// TODO refactoring
	dsn := "root:password@tcp(localhost:3306)/sample_test?parseTime=True"
	client, err := ent.Open("mysql", dsn)
	if err != nil {
		panic("failed opening connection to mysql: " + err.Error())
	}
	return client
}
func migrateSchema(client *ent.Client) {
	if err := client.Schema.Create(ctx); err != nil {
		panic(err)
	}
}

func withTransaction(t *testing.T, testFunc func(*ent.Tx)) {
	tx, err := client.Tx(ctx)
	if err != nil {
		t.Fatalf("failed starting a transaction: %v", err)
	}
	defer tx.Rollback()

	testFunc(tx)
}

func setupTestUserWithTransaction(ctx context.Context, tx *ent.Tx, t *testing.T, authUId string) *ent.User {
	testUser, err := tx.User.
		Create().
		SetAuth0UID(authUId).
		SetFirstName("Shohei").
		SetLastName("Otani").
		Save(ctx)
	if err != nil {
		t.Fatalf("failed creating test user: %v", err)
	}
	return testUser
}

func TestUserRepository_FindByAuth0UID_Exist(t *testing.T) {
	authUId := "test-auth0uid"
	withTransaction(t, func(tx *ent.Tx) {
		setupTestUserWithTransaction(ctx, tx, t, authUId)

		repo := NewUserRepository(tx.Client())
		user, err := repo.FindByAuth0UID(authUId)

		assert.NoError(t, err)
		assert.NotNil(t, user)
	})
}

func TestUserRepository_FindByAuth0UID_NotExist(t *testing.T) {
	authUId := "test-auth0uid"
	withTransaction(t, func(tx *ent.Tx) {
		setupTestUserWithTransaction(ctx, tx, t, authUId)

		repo := NewUserRepository(tx.Client())
		user, err := repo.FindByAuth0UID("notExist")

		assert.NoError(t, err)
		assert.Nil(t, user)
	})
}
