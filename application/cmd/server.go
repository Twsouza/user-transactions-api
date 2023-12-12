package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
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
	go transactionRepo.WithBulkConfig(100, 1).RunGroupTransactions()

	routes := router.SetupRouter(handler)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: routes,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server running on port %s\n", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %s", err)
		}
	}()

	gracefulShutdown(quit, srv, transactionRepo)
	log.Println("Server exited")
}

func gracefulShutdown(quit chan os.Signal, srv *http.Server, transactionRepo *repositories.TransactionRepositoryImpl) {
	log.Println("Press Ctrl+C to shutdown server")
	<-quit
	log.Println("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("HTTP server exited")

	ctx, cancel = context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := transactionRepo.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Transaction repository exited")
}
