//go:build integration
// +build integration

package handler_test

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"user-transactions/application/dto"
	"user-transactions/application/handler"
	"user-transactions/application/presenters"
	"user-transactions/application/repositories"
	"user-transactions/application/services"
	"user-transactions/core"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupService creates a new in-memory database and returns a gorm.DB
func setupService(t *testing.T) *services.TransactionService {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	assert.NoError(t, err)
	assert.NoError(t, db.AutoMigrate(&core.Transaction{}))

	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})

	tr := repositories.NewTransactionRepository(db)
	s, _ := services.NewTransactionService(tr)

	return s
}

func Test_TransactionHandler_Save(t *testing.T) {
	s := setupService(t)
	h := handler.NewTransactionHandler(s)

	// Create a new Gin router
	router := gin.Default()
	router.POST("/transactions", h.Save)

	t.Run("saving a transaction with valid payload", func(t *testing.T) {
		// Create a new HTTP request
		payload := `{
			"origin": "desktop-web",
			"user_id": "user123",
			"amount": 200,
			"type": "credit"
		}`
		req, err := http.NewRequest("POST", "/transactions", strings.NewReader(payload))
		assert.NoError(t, err)

		// Set the request content type
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		// Create a new HTTP response recorder
		res := httptest.NewRecorder()

		// Serve the HTTP request
		router.ServeHTTP(res, req)

		// Assert the response status code
		assert.Equal(t, http.StatusCreated, res.Code)

		// Assert the response body
		var result struct {
			Data dto.TransactionRes `json:"data"`
		}
		err = json.NewDecoder(res.Body).Decode(&result)
		assert.NoError(t, err)
		assert.NotEmpty(t, result.Data.ID)
		assert.Equal(t, "desktop-web", result.Data.Origin)
		assert.Equal(t, "user123", result.Data.UserID)
		assert.Equal(t, int64(200), result.Data.Amount)
		assert.Equal(t, "credit", result.Data.Type)
	})

	t.Run("saving a transaction with empty payload", func(t *testing.T) {
		// Create a new HTTP request
		payload := `{}`
		req, err := http.NewRequest("POST", "/transactions", strings.NewReader(payload))
		assert.NoError(t, err)

		// Set the request content type
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		// Create a new HTTP response recorder
		res := httptest.NewRecorder()

		// Serve the HTTP request
		router.ServeHTTP(res, req)

		// Assert the response status code
		assert.Equal(t, http.StatusBadRequest, res.Code)

		// Assert the response body
		assert.Contains(t, res.Body.String(), "Origin is a required field")
		assert.Contains(t, res.Body.String(), "UserID is a required field")
		assert.Contains(t, res.Body.String(), "Amount is a required field")
		assert.Contains(t, res.Body.String(), "Type is a required field")
	})

	t.Run("saving a transaction with invalid type", func(t *testing.T) {
		// Create a new HTTP request
		payload := `{
			"origin": "desktop-web",
			"user_id": "user123",
			"amount": 200,
			"type": "invalid"
		}`
		req, err := http.NewRequest("POST", "/transactions", strings.NewReader(payload))
		assert.NoError(t, err)

		// Set the request content type
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		// Create a new HTTP response recorder
		res := httptest.NewRecorder()

		// Serve the HTTP request
		router.ServeHTTP(res, req)

		// Assert the response status code
		assert.Equal(t, http.StatusBadRequest, res.Code)

		// Assert the response body
		assert.Contains(t, res.Body.String(), "Type must be one of [debit credit]")
	})

	t.Run("saving a transaction in XML format", func(t *testing.T) {
		// Create a new HTTP request
		payload := `<?xml version="1.0" encoding="UTF-8"?>
		<transaction>
			<origin>desktop-web</origin>
			<user_id>user123</user_id>
			<amount>200</amount>
			<type>credit</type>
		</transaction>`
		req, err := http.NewRequest("POST", "/transactions", strings.NewReader(payload))
		assert.NoError(t, err)

		// Set the request content type
		req.Header.Set("Content-Type", "application/xml")
		req.Header.Set("Accept", "application/xml")

		// Create a new HTTP response recorder
		res := httptest.NewRecorder()

		// Serve the HTTP request
		router.ServeHTTP(res, req)

		// Assert the response status code
		assert.Equal(t, http.StatusCreated, res.Code)

		// Assert the response body
		var result struct {
			Data dto.TransactionRes `xml:"transaction"`
		}
		err = xml.Unmarshal(res.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.NotEmpty(t, result.Data.ID)
		assert.Equal(t, "desktop-web", result.Data.Origin)
		assert.Equal(t, "user123", result.Data.UserID)
		assert.Equal(t, int64(200), result.Data.Amount)
		assert.Equal(t, "credit", result.Data.Type)
		assert.NotEmpty(t, result.Data.CreatedAt)
	})
}

func Test_TransactionHandler_Get(t *testing.T) {
	s := setupService(t)
	h := handler.NewTransactionHandler(s)

	// Create a new Gin router
	router := gin.Default()
	router.GET("/transactions/:id", h.Get)

	t.Run("getting a transaction by ID", func(t *testing.T) {

		// Create a new transaction
		transaction, errs := core.NewTransaction("desktop-web", "user123", 200, core.CREDIT)
		assert.Empty(t, errs)
		_, err := s.TransactionRepository.Insert(context.Background(), transaction)
		assert.NoError(t, err)

		// Create a new HTTP request
		req, err := http.NewRequest("GET", fmt.Sprintf("/transactions/%s", transaction.ID), nil)
		assert.NoError(t, err)

		// Set the request content type
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		// Create a new HTTP response recorder
		res := httptest.NewRecorder()

		// Serve the HTTP request
		router.ServeHTTP(res, req)

		// Assert the response status code
		assert.Equal(t, http.StatusOK, res.Code)

		// Assert the response body
		var result struct {
			Data dto.TransactionRes `json:"data"`
		}
		err = json.NewDecoder(res.Body).Decode(&result)
		assert.NoError(t, err)
		assert.Equal(t, transaction.ID.String(), result.Data.ID)
		assert.Equal(t, transaction.Origin, result.Data.Origin)
		assert.Equal(t, transaction.UserID, result.Data.UserID)
		assert.Equal(t, transaction.Amount, result.Data.Amount)
		assert.Equal(t, transaction.Type.String(), result.Data.Type)
		assert.Equal(t, transaction.CreatedAt, result.Data.CreatedAt)
	})

	t.Run("getting a transaction by ID accepting XML", func(t *testing.T) {

		// Create a new transaction
		transaction, errs := core.NewTransaction("desktop-web", "user123", 200, core.CREDIT)
		assert.Empty(t, errs)
		_, err := s.TransactionRepository.Insert(context.Background(), transaction)
		assert.NoError(t, err)

		// Create a new HTTP request
		req, err := http.NewRequest("GET", fmt.Sprintf("/transactions/%s", transaction.ID), nil)
		assert.NoError(t, err)

		// Set the request content type
		req.Header.Set("Content-Type", "application/xml")
		req.Header.Set("Accept", "application/xml")

		// Create a new HTTP response recorder
		res := httptest.NewRecorder()

		// Serve the HTTP request
		router.ServeHTTP(res, req)

		// Assert the response status code
		assert.Equal(t, http.StatusOK, res.Code)

		// Assert the response body
		var result struct {
			Data dto.TransactionRes `xml:"transaction"`
		}
		err = xml.NewDecoder(res.Body).Decode(&result)
		assert.NoError(t, err)
		assert.Equal(t, transaction.ID.String(), result.Data.ID)
		assert.Equal(t, transaction.Origin, result.Data.Origin)
		assert.Equal(t, transaction.UserID, result.Data.UserID)
		assert.Equal(t, transaction.Amount, result.Data.Amount)
		assert.Equal(t, transaction.Type.String(), result.Data.Type)
		assert.Equal(t, transaction.CreatedAt, result.Data.CreatedAt)
	})

	t.Run("getting a transaction by invalid ID", func(t *testing.T) {

		// Create a new HTTP request
		req, err := http.NewRequest("GET", "/transactions/invalid", nil)
		assert.NoError(t, err)

		// Set the request content type
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		// Create a new HTTP response recorder
		res := httptest.NewRecorder()

		// Serve the HTTP request
		router.ServeHTTP(res, req)

		// Assert the response status code
		assert.Equal(t, http.StatusBadRequest, res.Code)

		// Assert the response body
		assert.Contains(t, res.Body.String(), "record not found")
	})
}

func Test_TransactionHandler_List(t *testing.T) {
	s := setupService(t)
	h := handler.NewTransactionHandler(s)

	// Create a new Gin router
	router := gin.Default()
	router.GET("/transactions", h.List)

	for i := 0; i < 5; i++ {
		// Create a new transaction with desktop-web origin
		transaction, errs := core.NewTransaction("desktop-web", "user123", 100, core.CREDIT)
		assert.Empty(t, errs)
		_, err := s.TransactionRepository.Insert(context.Background(), transaction)
		assert.NoError(t, err)
	}
	// Create a new transaction with mobile-android origin
	transaction, errs := core.NewTransaction("mobile-android", "user123", 100, core.CREDIT)
	assert.Empty(t, errs)
	_, err := s.TransactionRepository.Insert(context.Background(), transaction)
	assert.NoError(t, err)

	t.Run("listing transactions", func(t *testing.T) {
		// Create a new HTTP request
		req, err := http.NewRequest("GET", "/transactions", nil)
		assert.NoError(t, err)

		// Set the request content type
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		// Create a new HTTP response recorder
		res := httptest.NewRecorder()

		// Serve the HTTP request
		router.ServeHTTP(res, req)

		// Assert the response status code
		assert.Equal(t, http.StatusOK, res.Code)

		// Assert the response body
		var result struct {
			Data []*dto.TransactionRes `json:"data"`
		}
		err = json.NewDecoder(res.Body).Decode(&result)
		assert.NoError(t, err)
		assert.Len(t, result.Data, 6)
	})

	t.Run("listing transactions with query params", func(t *testing.T) {
		// Create a new HTTP request
		req, err := http.NewRequest("GET", "/transactions?origin=desktop-web", nil)
		assert.NoError(t, err)

		// Set the request content type
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		// Create a new HTTP response recorder
		res := httptest.NewRecorder()

		// Serve the HTTP request
		router.ServeHTTP(res, req)

		// Assert the response status code
		assert.Equal(t, http.StatusOK, res.Code)

		// Assert the response body
		var result struct {
			Data []*dto.TransactionRes `json:"data"`
		}
		err = json.NewDecoder(res.Body).Decode(&result)
		assert.NoError(t, err)
		assert.Len(t, result.Data, 5)
		assert.Equal(t, "desktop-web", result.Data[0].Origin)
	})

	t.Run("listing transactions with pagination", func(t *testing.T) {
		// Create a new HTTP request
		req, err := http.NewRequest("GET", "/transactions?page=1&page_size=2", nil)
		assert.NoError(t, err)

		// Set the request content type
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		// Create a new HTTP response recorder
		res := httptest.NewRecorder()

		// Serve the HTTP request
		router.ServeHTTP(res, req)

		// Assert the response status code
		assert.Equal(t, http.StatusOK, res.Code)

		// Assert the response body
		var result struct {
			Data       []*dto.TransactionRes  `json:"data"`
			Pagination *presenters.Pagination `json:"pagination"`
		}
		err = json.NewDecoder(res.Body).Decode(&result)
		assert.NoError(t, err)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, 1, result.Pagination.Page)
		assert.Equal(t, 2, result.Pagination.PageSize)
	})
}
