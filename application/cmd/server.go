package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"user-transactions/application/data"
	"user-transactions/application/handler"
	"user-transactions/application/repositories"
	"user-transactions/application/router"
	"user-transactions/application/services"

	"github.com/joho/godotenv"
)

var (
	db   data.Database
	port string
)

func init() {
	godotenv.Load()

	autoMigrateDb, err := strconv.ParseBool(os.Getenv("AUTO_MIGRATE_DB"))
	if err != nil {
		log.Fatalf("error loading AUTO_MIGRATE_DB env var: %s", os.Getenv("AUTO_MIGRATE_DB"))
	}

	debug, err := strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		log.Fatalf("error loading DEBUG env var: %s", os.Getenv("DEBUG"))
	}

	db.Debug = debug
	db.AutoMigrateDb = autoMigrateDb
	db.Dsn = os.Getenv("DSN")
	port = os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
}

func main() {
	dbConn, err := db.Connect()
	if err != nil {
		log.Fatalf("error connecting to database: %s", err)
	}

	transactionRepo := repositories.NewTransactionRepository(dbConn)
	transactionSvc, _ := services.NewTransactionService(transactionRepo)
	handler := handler.NewTransactionHandler(transactionSvc)

	routes := router.SetupRouter(handler)
	fmt.Printf("Server running on port %s\n", port)
	routes.Run(":" + port)
}
