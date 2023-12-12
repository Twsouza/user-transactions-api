package repositories

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	"user-transactions/core"

	backoff "github.com/cenkalti/backoff/v4"
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
	CommitWg   sync.WaitGroup
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
				fmt.Printf("committing %d transactions with elapsed time of %v\n", len(bulk), time.Since(timer))

				r.CommitWg.Add(1)
				go r.CommitBulk(bulk...)

				bulk = nil
				timer = time.Now()
			}
		}
	}
}

func (r *TransactionRepositoryImpl) CommitBulk(transactions ...*core.Transaction) {
	defer r.CommitWg.Done()
	retryBo := backoff.NewExponentialBackOff()
	retryBo.MaxElapsedTime = 60 * time.Minute

	retryOp := func() error {
		err := r.Db.Create(transactions).Error
		if err != nil {
			fmt.Printf("error when committing %v transactions: %v, retrying in %v\n", len(transactions), err, retryBo.NextBackOff())
		}
		return err
	}

	if err := backoff.Retry(retryOp, retryBo); err != nil {
		fmt.Printf("error when committing %v transactions, aborting...", len(transactions))
		// here we have some options, send to a dead letter queue, another table or database, file, or retry again
		r.CommitBulk(transactions...)
	}
}

func (bc *TransactionRepositoryImpl) WithBulkConfig(maxBulkItems int, maxWaitingSeconds float64) *TransactionRepositoryImpl {
	bc.BulkConfig = &BulkConfig{
		MaxSize: maxBulkItems,
		MaxTime: maxWaitingSeconds,
	}

	return bc
}

func (r *TransactionRepositoryImpl) Shutdown(ctx context.Context) error {
	// since we close the HTTP server, InsertChan will not receive any more transactions
	// so we can close it safely without any data loss
	defer close(r.InsertChan)

	// Waiting all CommitBulk operations finish.
	done := make(chan struct{})
	go func() {
		r.CommitWg.Wait()
		close(done)
	}()

	log.Println("Waiting for all transactions to be committed...")
	select {
	case <-done:
		log.Println("All transactions committed, gracefully shutdown")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timed out waiting for transactions to be committed: %s", ctx.Err())
	}
}
