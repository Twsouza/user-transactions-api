package repositories

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	"user-transactions/core/entities"

	backoff "github.com/cenkalti/backoff/v4"
	"gorm.io/gorm"
)

type BulkConfig struct {
	MaxSize int
	MaxTime float64
}

type TransactionRepository struct {
	Db         *gorm.DB
	InsertChan chan *entities.Transaction
	BulkConfig *BulkConfig
	CommitWg   sync.WaitGroup
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{
		Db:         db,
		InsertChan: make(chan *entities.Transaction),
	}
}

func (r *TransactionRepository) Insert(ctx context.Context, transaction *entities.Transaction) (*entities.Transaction, error) {
	if r.BulkConfig != nil {
		r.InsertChan <- transaction
	} else {
		if err := r.Db.Create(transaction).Error; err != nil {
			return nil, err
		}
	}

	return transaction, nil
}

func (r *TransactionRepository) Find(ctx context.Context, id string) (*entities.Transaction, error) {
	var transaction entities.Transaction
	if err := r.Db.Where("id = ?", id).First(&transaction).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *TransactionRepository) List(ctx context.Context, pageSize, offset int, filter map[string]string) ([]*entities.Transaction, error) {
	query := r.Db.Limit(pageSize).Offset(offset)
	for key, value := range filter {
		query = query.Where(fmt.Sprintf("%v = ?", key), value)
	}

	var transactions []*entities.Transaction
	if err := query.Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *TransactionRepository) RunGroupTransactions() {
	var bulk []*entities.Transaction
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

func (r *TransactionRepository) CommitBulk(transactions ...*entities.Transaction) {
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

func (r *TransactionRepository) WithBulkConfig(maxBulkItems int, maxWaitingSeconds float64) *TransactionRepository {
	r.BulkConfig = &BulkConfig{
		MaxSize: maxBulkItems,
		MaxTime: maxWaitingSeconds,
	}

	return r
}

func (r *TransactionRepository) Shutdown(ctx context.Context) error {
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
