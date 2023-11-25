package repositories

import (
	"context"
	"fmt"
	"time"
	"user-transactions/core"

	"gorm.io/gorm"
)

type TransactionRepository interface {
	Insert(ctx context.Context, transaction *core.Transaction) (*core.Transaction, error)
	Find(ctx context.Context, id string) (*core.Transaction, error)
	List(ctx context.Context, pageSize, offset int, filter map[string]string) ([]*core.Transaction, error)
}

type BulkConfig struct {
	MaxSize int
	MaxTime float64
}

type TransactionRepositoryImpl struct {
	Db         *gorm.DB
	InsertChan chan *core.Transaction
	BulkConfig *BulkConfig
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepositoryImpl {
	return &TransactionRepositoryImpl{
		Db:         db,
		InsertChan: make(chan *core.Transaction),
	}
}

func (r *TransactionRepositoryImpl) Insert(ctx context.Context, transaction *core.Transaction) (*core.Transaction, error) {
	if r.BulkConfig != nil {
		r.InsertChan <- transaction
	} else {
		if err := r.Db.Create(transaction).Error; err != nil {
			return nil, err
		}
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

func (r *TransactionRepositoryImpl) RunGroupTransactions() {
	var bulk []*core.Transaction
	timer := time.Now()
	for {
		select {
		case transaction := <-r.InsertChan:
			bulk = append(bulk, transaction)
		default:
			if len(bulk) > 0 && (len(bulk) >= r.BulkConfig.MaxSize || time.Since(timer).Seconds() >= r.BulkConfig.MaxTime) {
				fmt.Printf("running bulk creation of %v transactions\n", len(bulk))

				r.Db.Create(bulk)
				bulk = nil
				timer = time.Now()
			}
		}
	}
}

func (bc *TransactionRepositoryImpl) WithBulkConfig(maxBulkItems int, maxWaitingSeconds float64) *TransactionRepositoryImpl {
	bc.BulkConfig = &BulkConfig{
		MaxSize: maxBulkItems,
		MaxTime: maxWaitingSeconds,
	}

	return bc
}
