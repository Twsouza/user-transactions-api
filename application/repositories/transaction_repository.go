package repositories

import (
	"context"
	"fmt"
	"user-transactions/core"

	"gorm.io/gorm"
)

type TransactionRepository interface {
	Insert(ctx context.Context, transaction *core.Transaction) (*core.Transaction, error)
	Find(ctx context.Context, id string) (*core.Transaction, error)
	List(ctx context.Context, pageSize, offset int, filter map[string]string) ([]*core.Transaction, error)
}

type TransactionRepositoryImpl struct {
	Db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepositoryImpl {
	return &TransactionRepositoryImpl{Db: db}
}

func (r *TransactionRepositoryImpl) Insert(ctx context.Context, transaction *core.Transaction) (*core.Transaction, error) {
	if err := r.Db.Create(transaction).Error; err != nil {
		return nil, err
	}
	return transaction, nil
}

func (r *TransactionRepositoryImpl) Find(ctx context.Context, id string) (*core.Transaction, error) {
	var transaction core.Transaction
	if err := r.Db.Where("id = ?", id).First(&transaction).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *TransactionRepositoryImpl) List(ctx context.Context, pageSize, offset int, filter map[string]string) ([]*core.Transaction, error) {
	query := r.Db.Limit(pageSize).Offset(offset)
	for key, value := range filter {
		query = query.Where(fmt.Sprintf("%v = ?", key), value)
	}

	var transactions []*core.Transaction
	if err := query.Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}
