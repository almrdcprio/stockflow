package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"stockflow/internal/database"
	"stockflow/internal/handler"
	"stockflow/internal/repository"

	"github.com/go-chi/chi/v5"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := database.NewPostgresPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	router := chi.NewRouter()
	productRepo := repository.NewProductRepository(pool)
	productHandler := handler.NewProductHandler(productRepo)
	orderRepo := repository.NewOrderRepository(pool)
	orderHandler := handler.NewOrderHandler(orderRepo)
	summaryRepo := repository.NewSummaryRepository(pool)
	summaryHandler := handler.NewSummaryHandler(summaryRepo)

	router.Get("/api/products", productHandler.ListProducts)
	router.Post("/api/products", productHandler.CreateProduct)
	router.Put("/api/products/{id}", productHandler.UpdateProduct)
	router.Delete("/api/products/{id}", productHandler.DeleteProduct)

	router.Post("/api/products/{id}/stock", productHandler.AdjustStock)
	router.Get("/api/products/{id}/movements", productHandler.GetMovements)

	router.Post("/api/orders", orderHandler.CreateOrder)
	router.Get("/api/orders/{id}", orderHandler.GetOrder)
	router.Post("/api/orders/{id}/cancel", orderHandler.CancelOrder)

	router.Get("/api/dashboard/summary", summaryHandler.GetSummary)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"status": "ok",
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	})

	router.Get("/health/db", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}

		response := map[string]string{
			"status":   "ok",
			"database": "connected",
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	})

	log.Println("StockFlow API running on :8080")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
