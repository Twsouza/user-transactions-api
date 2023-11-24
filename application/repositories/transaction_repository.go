package repositories

import (
	"context"
	"user-transactions/core"
)

type TransactionRepository interface {
	Insert(ctx context.Context, transaction *core.Transaction) (*core.Transaction, error)
	Find(ctx context.Context, id string) (*core.Transaction, error)
	List(ctx context.Context, pageSize, offset int, filter map[string]string) ([]*core.Transaction, error)
}
