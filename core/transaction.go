package core

import (
	"fmt"
	"time"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/google/uuid"
)

type OperationType string

const (
	DEBIT  OperationType = "debit"
	CREDIT OperationType = "credit"
)

type Transaction struct {
	ID        uuid.UUID
	Origin    string        `validate:"required"`
	UserID    string        `validate:"required"`
	Amount    int64         `validate:"required,numeric"` // cents, 0 is not allowed
	Type      OperationType `validate:"required,oneof=debit credit"`
	CreatedAt time.Time
}

var (
	uni      *ut.UniversalTranslator
	trans    ut.Translator
	validate *validator.Validate
)

func init() {
	en := en.New()
	uni = ut.New(en, en)
	trans, _ = uni.GetTranslator("en")

	validate = validator.New(validator.WithRequiredStructEnabled())
	en_translations.RegisterDefaultTranslations(validate, trans)
}

func NewTransaction(origin, userId string, amount int64, opType OperationType) (*Transaction, []error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, []error{err}
	}

	t := &Transaction{
		ID:        id,
		Origin:    origin,
		UserID:    userId,
		Amount:    amount,
		Type:      opType,
		CreatedAt: time.Now(),
	}

	if err := t.validate(); err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Transaction) validate() (errs []error) {
	if t.Type == DEBIT && t.Amount > 0 {
		errs = append(errs, fmt.Errorf("Amount must be negative for debit transactions"))
	}

	if t.Type == CREDIT && t.Amount < 0 {
		errs = append(errs, fmt.Errorf("Amount must be positive for credit transactions"))
	}

	err := validate.Struct(t)
	if err == nil {
		return
	}

	verrs := err.(validator.ValidationErrors).Translate(trans)
	for _, v := range verrs {
		errs = append(errs, fmt.Errorf(v))
	}

	return
}
