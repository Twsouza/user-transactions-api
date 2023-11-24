package handler

import (
	"net/http"
	"strconv"
	"user-transactions/application/dto"
	"user-transactions/application/presenters"
	"user-transactions/application/services"

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	TransactionService *services.TransactionService
}

func NewTransactionHandler(transactionService *services.TransactionService) *TransactionHandler {
	return &TransactionHandler{TransactionService: transactionService}
}

func (th *TransactionHandler) Save(c *gin.Context) {
	req := &dto.CreateTransactionReq{}
	if err := c.Bind(&req); err != nil {
		c.Negotiate(http.StatusBadRequest, gin.Negotiate{
			Offered: []string{"application/json", "application/xml"},
			Data:    presenters.TransformErrorToApiError(err),
		})
		return
	}

	transaction, errs := th.TransactionService.CreateTransaction(c, req)
	if len(errs) > 0 {
		c.Negotiate(http.StatusBadRequest, gin.Negotiate{
			Offered: []string{"application/json", "application/xml"},
			Data:    presenters.TransformErrorToApiError(errs...),
		})
		return
	}

	c.Negotiate(http.StatusCreated, gin.Negotiate{
		Offered: []string{"application/json", "application/xml"},
		Data:    presenters.TransformDataToApiFormat(transaction),
	})
}

func (th *TransactionHandler) Get(c *gin.Context) {
	id := c.Param("id")

	transaction, err := th.TransactionService.GetTransaction(c, id)
	if err != nil {
		c.Negotiate(http.StatusBadRequest, gin.Negotiate{
			Offered: []string{"application/json", "application/xml"},
			Data:    presenters.TransformErrorToApiError(err),
		})
		return
	}

	c.Negotiate(http.StatusOK, gin.Negotiate{
		Offered: []string{"application/json", "application/xml"},
		Data:    presenters.TransformDataToApiFormat(transaction),
	})
}

func (th *TransactionHandler) List(c *gin.Context) {
	// Get query parameters from URL
	queryParams := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}
	pageSizeStr, pageStr := c.Query("page_size"), c.Query("page")
	// Convert pageSize and offset to int
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		pageSize = 10
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 0
	}

	// Use query parameters as a filter for List method of TransactionService
	transactions, err := th.TransactionService.ListTransactions(c, pageSize, page, queryParams)
	if err != nil {
		c.Negotiate(http.StatusBadRequest, gin.Negotiate{
			Offered: []string{"application/json", "application/xml"},
			Data:    presenters.TransformErrorToApiError(err),
		})
		return
	}

	c.Negotiate(http.StatusOK, gin.Negotiate{
		Offered: []string{"application/json", "application/xml"},
		Data:    presenters.TransformDataToApiFormat(transactions).WithPagination(page, pageSize),
	})
}
