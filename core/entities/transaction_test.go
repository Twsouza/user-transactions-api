package entities_test

import (
	"fmt"
	"testing"
	"user-transactions/core/entities"

	"github.com/stretchr/testify/assert"
)

func Test_NewTransaction(t *testing.T) {
	t.Run("create credit transaction", func(t *testing.T) {
		origin := "desktop-web"
		userId := "123"
		amount := int64(1000)
		transactionType := entities.CREDIT

		transaction, err := entities.NewTransaction(origin, userId, amount, transactionType)
		assert.Equal(t, len(err), 0)
		assert.NotNil(t, transaction)
		assert.Equal(t, origin, transaction.Origin)
		assert.Equal(t, userId, transaction.UserID)
		assert.Equal(t, amount, transaction.Amount)
		assert.Equal(t, transactionType, transaction.Type)
		assert.NotEmpty(t, transaction.ID)
		assert.NotEmpty(t, transaction.CreatedAt)
	})

	t.Run("create debit transaction", func(t *testing.T) {
		origin := "desktop-web"
		userId := "123"
		amount := int64(-1000)
		transactionType := entities.DEBIT

		transaction, err := entities.NewTransaction(origin, userId, amount, transactionType)
		fmt.Printf("%+v\n", err)
		assert.Equal(t, len(err), 0)
		assert.NotNil(t, transaction)
		assert.Equal(t, origin, transaction.Origin)
		assert.Equal(t, userId, transaction.UserID)
		assert.Equal(t, amount, transaction.Amount)
		assert.Equal(t, transactionType, transaction.Type)
		assert.NotEmpty(t, transaction.ID)
		assert.NotEmpty(t, transaction.CreatedAt)
	})

	t.Run("create transaction with invalid origin", func(t *testing.T) {
		origin := ""
		userId := "123"
		amount := int64(1000)
		transactionType := entities.CREDIT

		transaction, err := entities.NewTransaction(origin, userId, amount, transactionType)
		assert.Equal(t, len(err), 1)
		assert.Nil(t, transaction)
		assert.Equal(t, err[0].Error(), "Origin is a required field")
	})

	t.Run("create transaction with invalid userId", func(t *testing.T) {
		origin := "desktop-web"
		userId := ""
		amount := int64(1000)
		transactionType := entities.CREDIT

		transaction, err := entities.NewTransaction(origin, userId, amount, transactionType)
		assert.Equal(t, len(err), 1)
		assert.Nil(t, transaction)
		assert.Equal(t, err[0].Error(), "UserID is a required field")
	})

	t.Run("create transaction with invalid amount", func(t *testing.T) {
		origin := "desktop-web"
		userId := "123"
		amount := int64(0)
		transactionType := entities.CREDIT

		transaction, err := entities.NewTransaction(origin, userId, amount, transactionType)
		assert.Equal(t, len(err), 1)
		assert.Nil(t, transaction)
		assert.Equal(t, "Amount is a required field", err[0].Error())
	})

	t.Run("create credit transaction with invalid amount", func(t *testing.T) {
		origin := "desktop-web"
		userId := "123"
		amount := int64(-1000)
		transactionType := entities.CREDIT

		transaction, err := entities.NewTransaction(origin, userId, amount, transactionType)
		assert.Equal(t, len(err), 1)
		assert.Nil(t, transaction)
		assert.Equal(t, err[0].Error(), "Amount must be positive for credit transactions")
	})

	t.Run("create debit transaction with invalid amount", func(t *testing.T) {
		origin := "desktop-web"
		userId := "123"
		amount := int64(1000)
		transactionType := entities.DEBIT

		transaction, err := entities.NewTransaction(origin, userId, amount, transactionType)
		assert.Equal(t, len(err), 1)
		assert.Nil(t, transaction)
		assert.Equal(t, "Amount must be negative for debit transactions", err[0].Error())
	})

	t.Run("create transaction with invalid type", func(t *testing.T) {
		origin := "desktop-web"
		userId := "123"
		amount := int64(1000)
		transactionType := entities.OperationType("invalid")

		transaction, err := entities.NewTransaction(origin, userId, amount, transactionType)
		assert.Equal(t, len(err), 1)
		assert.Nil(t, transaction)
		assert.Equal(t, "Type must be one of [debit credit]", err[0].Error())
	})
}

func Test_OperationType_String(t *testing.T) {
	debit := entities.DEBIT
	credit := entities.CREDIT

	assert.Equal(t, "debit", debit.String())
	assert.Equal(t, "credit", credit.String())
}
