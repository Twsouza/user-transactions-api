package dto

import (
	"encoding/xml"
	"time"
)

type CreateTransactionReq struct {
	XMLName xml.Name `json:"-" xml:"transaction"`
	Origin  string   `json:"origin" xml:"origin"`
	UserID  string   `json:"user_id" xml:"user_id"`
	Amount  int64    `json:"amount" xml:"amount"`
	Type    string   `json:"type" xml:"type"`
}

type TransactionRes struct {
	XMLName   xml.Name  `json:"-" xml:"transaction"`
	ID        string    `json:"id" xml:"id"`
	Origin    string    `json:"origin" xml:"origin"`
	UserID    string    `json:"user_id" xml:"user_id"`
	Amount    int64     `json:"amount" xml:"amount"`
	Type      string    `json:"type" xml:"type"`
	CreatedAt time.Time `json:"created_at" xml:"created_at"`
}
