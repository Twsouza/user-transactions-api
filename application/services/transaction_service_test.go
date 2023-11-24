package services_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"user-transactions/application/dto"
	mock_repositories "user-transactions/application/repositories/mock"
	"user-transactions/application/services"
	"user-transactions/core"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_NewTransactionService(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := mock_repositories.NewMockTransactionRepository(ctrl)

	t.Run("with default timeout", func(t *testing.T) {
		service, err := services.NewTransactionService(mock)
		assert.Nil(t, err)
		assert.NotNil(t, service)
		// testing default timeout
		assert.Equal(t, 5, service.Timeout)
	})

	t.Run("with custom timeout", func(t *testing.T) {
		os.Setenv("TIMEOUT_SERVICES", "10")
		service, err := services.NewTransactionService(mock)
		assert.Nil(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, 10, service.Timeout)
	})
}

func Test_TransactionService_CreateTransaction(t *testing.T) {
	ctx := context.Background()
	req := &dto.CreateTransactionReq{
		Amount: -150,
		Origin: "desktop-web",
		Type:   "debit",
		UserID: "user123",
	}
	expected := &core.Transaction{
		Origin: "desktop-web",
		UserID: "user123",
		Amount: -150,
		Type:   core.DEBIT,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mock_repositories.NewMockTransactionRepository(ctrl)

	service, err := services.NewTransactionService(mockRepo)
	assert.Nil(t, err)

	t.Run("create the new transaction", func(t *testing.T) {
		mockRepo.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(expected, nil)

		res, err := service.CreateTransaction(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, expected.Origin, res.Origin)
		assert.Equal(t, expected.UserID, res.UserID)
		assert.Equal(t, expected.Amount, res.Amount)
		assert.Equal(t, string(expected.Type), res.Type)
		assert.NotEmpty(t, res.CreatedAt)
		assert.NotEmpty(t, res.ID)
	})

	t.Run("don't create with missing required fields", func(t *testing.T) {
		req := &dto.CreateTransactionReq{}
		res, err := service.CreateTransaction(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func Test_TransactionService_GetTransaction(t *testing.T) {
	ctx := context.Background()
	id, err := uuid.NewRandom()
	assert.Nil(t, err)
	idStr := id.String()
	expected := &core.Transaction{
		ID:        id,
		Origin:    "desktop-web",
		UserID:    "user123",
		Amount:    150,
		Type:      "debit",
		CreatedAt: time.Now(),
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mock_repositories.NewMockTransactionRepository(ctrl)

	service, err := services.NewTransactionService(mockRepo)
	assert.Nil(t, err)

	t.Run("get an existing transaction", func(t *testing.T) {
		mockRepo.EXPECT().Find(gomock.Any(), idStr).Return(expected, nil)

		res, err := service.GetTransaction(ctx, idStr)

		assert.NoError(t, err)
		assert.Equal(t, expected.ID.String(), res.ID)
		assert.Equal(t, expected.Origin, res.Origin)
		assert.Equal(t, expected.UserID, res.UserID)
		assert.Equal(t, expected.Amount, res.Amount)
		assert.Equal(t, expected.Type.String(), res.Type)
		assert.Equal(t, expected.CreatedAt, res.CreatedAt)
	})

	t.Run("get a non-existing transaction", func(t *testing.T) {
		mockRepo.EXPECT().Find(gomock.Any(), idStr).Return(nil, errors.New("transaction not found"))

		res, err := service.GetTransaction(ctx, idStr)

		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func Test_TransactionService_ListTransactions(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mock_repositories.NewMockTransactionRepository(ctrl)

	service, err := services.NewTransactionService(mockRepo)
	assert.Nil(t, err)

	t.Run("list transactions with filter", func(t *testing.T) {
		filter := map[string]string{
			"origin": "desktop-web",
		}
		expected := []*core.Transaction{
			{
				ID:        uuid.New(),
				Origin:    "desktop-web",
				UserID:    "user123",
				Amount:    150,
				Type:      "debit",
				CreatedAt: time.Now(),
			},
		}

		mockRepo.EXPECT().List(gomock.Any(), 10, 0, filter).Return(expected, nil)

		res, err := service.ListTransactions(ctx, 10, 0, filter)

		assert.NoError(t, err)
		assert.Equal(t, len(expected), len(res))
		assert.Equal(t, expected[0].ID.String(), res[0].ID)
		assert.Equal(t, expected[0].Origin, res[0].Origin)
		assert.Equal(t, expected[0].UserID, res[0].UserID)
		assert.Equal(t, expected[0].Amount, res[0].Amount)
		assert.Equal(t, expected[0].Type.String(), res[0].Type)
		assert.Equal(t, expected[0].CreatedAt, res[0].CreatedAt)
	})

	t.Run("list transactions", func(t *testing.T) {
		expected := []*core.Transaction{
			{
				ID:        uuid.New(),
				Origin:    "desktop-web",
				UserID:    "user123",
				Amount:    150,
				Type:      "debit",
				CreatedAt: time.Now(),
			},
			{
				ID:        uuid.New(),
				Origin:    "mobile-app",
				UserID:    "user456",
				Amount:    200,
				Type:      "credit",
				CreatedAt: time.Now(),
			},
		}

		mockRepo.EXPECT().List(gomock.Any(), 10, 0, make(map[string]string)).Return(expected, nil)

		res, err := service.ListTransactions(ctx, 10, 0, nil)

		assert.NoError(t, err)
		assert.Equal(t, len(expected), len(res))
		assert.Equal(t, expected[0].ID.String(), res[0].ID)
		assert.Equal(t, expected[0].Origin, res[0].Origin)
		assert.Equal(t, expected[0].UserID, res[0].UserID)
		assert.Equal(t, expected[0].Amount, res[0].Amount)
		assert.Equal(t, expected[0].Type.String(), res[0].Type)
		assert.Equal(t, expected[0].CreatedAt, res[0].CreatedAt)
	})
}
