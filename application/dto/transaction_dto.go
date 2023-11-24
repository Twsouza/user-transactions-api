package dto

import "time"

type CreateTransactionReq struct {
	Origin string `json:"origin"`
	UserID string `json:"user_id"`
	Amount int64  `json:"amount"`
	Type   string `json:"type"`
}

type TransactionRes struct {
	ID        string    `json:"id"`
	Origin    string    `json:"origin"`
	UserID    string    `json:"user_id"`
	Amount    int64     `json:"amount"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}
