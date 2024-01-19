package repositories

import (
	"context"
	"user-transactions/core/entities"
)

type TransactionRepository interface {
	Insert(ctx context.Context, transaction *entities.Transaction) (*entities.Transaction, error)
	Find(ctx context.Context, id string) (*entities.Transaction, error)
	List(ctx context.Context, pageSize, offset int, filter map[string]string) ([]*entities.Transaction, error)
}
