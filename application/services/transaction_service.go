package services

import (
	"context"
	"os"
	"strconv"
	"time"

	"user-transactions/application/dto"
	"user-transactions/application/repositories"
	"user-transactions/core/entities"
)

type TransactionService struct {
	Timeout               int
	TransactionRepository repositories.TransactionRepository
}

func NewTransactionService(tr repositories.TransactionRepository) (*TransactionService, error) {
	timeout, err := strconv.Atoi(os.Getenv("TIMEOUT_SERVICES"))
	if err != nil {
		timeout = 5
	}

	return &TransactionService{
		Timeout:               timeout,
		TransactionRepository: tr,
	}, nil
}

func (ts *TransactionService) CreateTransaction(c context.Context, req *dto.CreateTransactionReq) (*dto.TransactionRes, []error) {
	ctx, cancel := context.WithTimeout(c, time.Duration(ts.Timeout)*time.Second)
	defer cancel()

	transaction, errs := entities.NewTransaction(req.Origin, req.UserID, req.Amount, entities.OperationType(req.Type))
	if errs != nil {
		return nil, errs
	}

	_, err := ts.TransactionRepository.Insert(ctx, transaction)
	if err != nil {
		return nil, []error{err}
	}

	return &dto.TransactionRes{
		ID:        transaction.ID.String(),
		Origin:    transaction.Origin,
		UserID:    transaction.UserID,
		Amount:    transaction.Amount,
		Type:      transaction.Type.String(),
		CreatedAt: transaction.CreatedAt,
	}, nil
}

func (ts *TransactionService) GetTransaction(c context.Context, id string) (*dto.TransactionRes, error) {
	ctx, cancel := context.WithTimeout(c, time.Duration(ts.Timeout)*time.Second)
	defer cancel()

	transaction, err := ts.TransactionRepository.Find(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dto.TransactionRes{
		ID:        transaction.ID.String(),
		Origin:    transaction.Origin,
		UserID:    transaction.UserID,
		Amount:    transaction.Amount,
		Type:      transaction.Type.String(),
		CreatedAt: transaction.CreatedAt,
	}, nil
}

func (ts *TransactionService) ListTransactions(c context.Context, pageSize, offset int, filter map[string]string) ([]*dto.TransactionRes, error) {
	ctx, cancel := context.WithTimeout(c, time.Duration(ts.Timeout)*time.Second)
	defer cancel()

	// only allow certain filters
	allowedFilters := map[string]bool{
		"origin":  true,
		"user_id": true,
		"type":    true,
	}
	validFilters := map[string]string{}
	for key, value := range filter {
		if allowedFilters[key] {
			validFilters[key] = value
		}
	}

	transactions, err := ts.TransactionRepository.List(ctx, pageSize, offset, validFilters)
	if err != nil {
		return nil, err
	}

	var res []*dto.TransactionRes
	for _, transaction := range transactions {
		res = append(res, &dto.TransactionRes{
			ID:        transaction.ID.String(),
			Origin:    transaction.Origin,
			UserID:    transaction.UserID,
			Amount:    transaction.Amount,
			Type:      transaction.Type.String(),
			CreatedAt: transaction.CreatedAt,
		})
	}

	return res, nil
}
