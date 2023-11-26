package main

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"sample-server/ent"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

var client *ent.Client
var ctx context.Context

func TestMain(m *testing.M) {
	client = setupMySQLDatabase()
	defer client.Close()
	ctx = context.Background()

	exitVal := m.Run()

	os.Exit(exitVal)
}

func setupMySQLDatabase() *ent.Client {
	// TODO
	dsn := "root:password@tcp(localhost:3306)/sample?parseTime=True"
	client, err := ent.Open("mysql", dsn)
	if err != nil {
		panic("failed opening connection to mysql: " + err.Error())
	}
	return client
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
