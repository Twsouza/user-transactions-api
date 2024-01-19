//go:build integration
// +build integration

package repositories_test

import (
	"context"
	"testing"
	"user-transactions/core/entities"
	"user-transactions/infrastructure/repositories"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupDB creates a new in-memory database and returns a gorm.DB
func setupDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	assert.NoError(t, err)
	assert.NoError(t, db.AutoMigrate(&entities.Transaction{}))

	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})

	return db
}

func Test_TransactionRepositoryImpl_Insert(t *testing.T) {
	db := setupDB(t)

	repo := repositories.NewTransactionRepository(db)

	transaction, errs := entities.NewTransaction("desktop-web", "user123", 200, entities.CREDIT)
	assert.Empty(t, errs)
	ctx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()

	result, err := repo.Insert(ctx, transaction)
	assert.NoError(t, err)
	assert.Equal(t, transaction, result)
}

func Test_TransactionRepositoryImpl_Find(t *testing.T) {
	db := setupDB(t)

	repo := repositories.NewTransactionRepository(db)

	t.Run("finding a transaction that exists", func(t *testing.T) {
		transaction, errs := entities.NewTransaction("desktop-web", "user123", 200, entities.CREDIT)
		assert.Empty(t, errs)

		err := db.Create(transaction).Error
		assert.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 5)
		defer cancel()

		found, err := repo.Find(ctx, transaction.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, transaction.ID, found.ID)
		assert.Equal(t, transaction.Origin, found.Origin)
		assert.Equal(t, transaction.UserID, found.UserID)
		assert.Equal(t, transaction.Amount, found.Amount)
		assert.Equal(t, transaction.Type, found.Type)
		assert.Equal(t, transaction.CreatedAt, found.CreatedAt)
	})

	t.Run("finding a transaction that does not exist", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5)
		defer cancel()

		found, err := repo.Find(ctx, "non-existing-id")
		assert.Error(t, err)
		assert.Equal(t, "record not found", err.Error())
		assert.Nil(t, found)
	})
}

func Test_TransactionRepositoryImpl_List(t *testing.T) {
	t.Run("listing transactions with no filters", func(t *testing.T) {
		db := setupDB(t)

		repo := repositories.NewTransactionRepository(db)

		transaction1, errs := entities.NewTransaction("desktop-web", "user123", 200, entities.CREDIT)
		assert.Empty(t, errs)
		err := db.Create(transaction1).Error
		assert.NoError(t, err)

		transaction2, errs := entities.NewTransaction("desktop-web", "user123", -200, entities.DEBIT)
		assert.Empty(t, errs)
		err = db.Create(transaction2).Error
		assert.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 5)
		defer cancel()

		found, err := repo.List(ctx, 10, 0, map[string]string{})
		assert.NoError(t, err)
		assert.Len(t, found, 2)
		assert.Equal(t, transaction1.ID, found[0].ID)
		assert.Equal(t, transaction2.ID, found[1].ID)
	})

	t.Run("listing transactions with filters", func(t *testing.T) {
		db := setupDB(t)

		repo := repositories.NewTransactionRepository(db)

		transaction1, errs := entities.NewTransaction("desktop-web", "user123", 200, entities.CREDIT)
		assert.Empty(t, errs)
		err := db.Create(transaction1).Error
		assert.NoError(t, err)

		transaction2, errs := entities.NewTransaction("desktop-web", "user456", -200, entities.DEBIT)
		assert.Empty(t, errs)
		err = db.Create(transaction2).Error
		assert.NoError(t, err)

		transaction3, errs := entities.NewTransaction("mobile-android", "user456", -200, entities.DEBIT)
		assert.Empty(t, errs)
		err = db.Create(transaction3).Error
		assert.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 5)
		defer cancel()

		found, err := repo.List(ctx, 10, 0, map[string]string{"origin": "desktop-web"})
		assert.NoError(t, err)
		assert.Len(t, found, 2)
		assert.Equal(t, transaction1.ID, found[0].ID)
		assert.Equal(t, transaction2.ID, found[1].ID)
	})

	t.Run("listing transactions with filters that doesn't have transactions", func(t *testing.T) {
		db := setupDB(t)

		repo := repositories.NewTransactionRepository(db)

		transaction1, errs := entities.NewTransaction("desktop-web", "user123", 200, entities.CREDIT)
		assert.Empty(t, errs)
		err := db.Create(transaction1).Error
		assert.NoError(t, err)

		transaction2, errs := entities.NewTransaction("desktop-web", "user456", -200, entities.DEBIT)
		assert.Empty(t, errs)
		err = db.Create(transaction2).Error
		assert.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 5)
		defer cancel()

		found, err := repo.List(ctx, 10, 0, map[string]string{"origin": "mobile-android"})
		assert.NoError(t, err)
		assert.Len(t, found, 0)
	})

	t.Run("listing transactions with pagination", func(t *testing.T) {
		db := setupDB(t)

		repo := repositories.NewTransactionRepository(db)

		transaction1, errs := entities.NewTransaction("desktop-web", "user123", 200, entities.CREDIT)
		assert.Empty(t, errs)
		err := db.Create(transaction1).Error
		assert.NoError(t, err)

		transaction2, errs := entities.NewTransaction("desktop-web", "user456", -200, entities.DEBIT)
		assert.Empty(t, errs)
		err = db.Create(transaction2).Error
		assert.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 5)
		defer cancel()

		found, err := repo.List(ctx, 1, 0, map[string]string{})
		assert.NoError(t, err)
		assert.Len(t, found, 1)
		assert.Equal(t, transaction1.ID, found[0].ID)

		found, err = repo.List(ctx, 1, 1, map[string]string{})
		assert.NoError(t, err)
		assert.Len(t, found, 1)
		assert.Equal(t, transaction2.ID, found[0].ID)
	})
}
