package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"stockflow/internal/repository"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type productHandler struct {
	repo *repository.ProductRepository
}

func NewProductHandler(repo *repository.ProductRepository) *productHandler {
	return &productHandler{
		repo: repo,
	}
}

type productListResponse struct {
	Data  []repository.Product `json:"data"`
	Page  int                  `json:"page"`
	Limit int                  `json:"limit"`
	Total int                  `json:"total"`
}

type productCreateRequest struct {
	SKU   string `json:"sku"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

type AdjustStockRequest struct {
	Type string `json:"type"`
	Qty  int    `json:"qty"`
	Note string `json:"note"`
}

func (h *productHandler) ListProducts(
	w http.ResponseWriter, r *http.Request,
) {
	query := r.URL.Query()
	search := strings.TrimSpace(query.Get("search"))
	page := 1
	limit := 10

	if value := query.Get("page"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			http.Error(w, "invalid page parameter", http.StatusBadRequest)
			return
		}
		page = parsed
	}

	if value := query.Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			http.Error(w, "invalid limit parameter", http.StatusBadRequest)
			return
		}
	}

	products, total, err := h.repo.GetProducts(
		r.Context(),
		search,
		page,
		limit,
	)
	if err != nil {
		log.Printf("failed to fetch products: %v", err)
		http.Error(w, "failed to fetch products", http.StatusInternalServerError)
		return
	}

	response := productListResponse{
		Data:  products,
		Page:  page,
		Limit: limit,
		Total: total,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *productHandler) CreateProduct(
	w http.ResponseWriter, r *http.Request,
) {
	var request productCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	request.SKU = strings.TrimSpace(request.SKU)
	request.Name = strings.TrimSpace(request.Name)

	if request.SKU == "" {
		http.Error(w, "Product SKU is required", http.StatusBadRequest)
		return
	}

	if request.Name == "" {
		http.Error(w, "Product name is required", http.StatusBadRequest)
		return
	}

	if request.Price < 0 {
		http.Error(w, "Product price must be greater than or equal to zero", http.StatusBadRequest)
		return
	}

	product, err := h.repo.CreateProduct(
		r.Context(),
		request.SKU,
		request.Name,
		request.Price,
	)

	if err != nil {
		if errors.Is(err, repository.ErrDuplicateSKU) {
			http.Error(w, "Product SKU already exists", http.StatusConflict)
			return
		}

		http.Error(w, "failed to create product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(product); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

}

func (h *productHandler) UpdateProduct(
	w http.ResponseWriter, r *http.Request,
) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var request productCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	request.SKU = strings.TrimSpace(request.SKU)
	request.Name = strings.TrimSpace(request.Name)

	if request.SKU == "" {
		http.Error(w, "Product SKU is required", http.StatusBadRequest)
		return
	}

	if request.Name == "" {
		http.Error(w, "Product name is required", http.StatusBadRequest)
		return
	}

	if request.Price < 0 {
		http.Error(w, "Product price must be greater than or equal to zero", http.StatusBadRequest)
		return
	}

	product, err := h.repo.UpdateProduct(
		r.Context(),
		id,
		request.SKU,
		request.Name,
		request.Price,
	)

	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			respondError(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found")
			return
		}

		if errors.Is(err, repository.ErrDuplicateSKU) {
			respondError(w, http.StatusConflict, "DUPLICATE_SKU", "Product SKU already exists")
			return
		}

		log.Printf("failed to update product: %v", err)
		respondError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(product); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *productHandler) DeleteProduct(
	w http.ResponseWriter, r *http.Request,
) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	err = h.repo.DeleteProduct(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)

			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]string{
					"code": "PRODUCT_NOT_FOUND",
				},
			})
			return
		}

		if errors.Is(err, repository.ErrProductHasHistory) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)

			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]string{
					"code": "PRODUCT_HAS_HISTORY",
				},
			})
			return
		}

		log.Printf("failed to delete product: %v", err)
		http.Error(w, "failed to delete product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

func (h *productHandler) AdjustStock(
	w http.ResponseWriter, r *http.Request,
) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var request AdjustStockRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if request.Type != "in" && request.Type != "out" {
		http.Error(w, "Invalid type! Must be either 'in' or 'out'", http.StatusBadRequest)
		return
	}

	if request.Qty <= 0 {
		http.Error(w, "Quantity must be greater than zero", http.StatusBadRequest)
		return
	}

	newStock, err := h.repo.AdjustStock(
		r.Context(),
		id,
		request.Type,
		request.Qty,
		request.Note,
	)

	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}

		if errors.Is(err, repository.ErrInsufficientStock) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)

			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]any{
					"code":            "INSUFFICIENT_STOCK",
					"available_stock": newStock,
				},
			})
			return
		}

		log.Printf("failed to adjust stock: %v", err)
		http.Error(w, "failed to adjust stock", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"id":    id,
		"type":  request.Type,
		"qty":   request.Qty,
		"stock": newStock,
		"note":  request.Note,
	})

}

func (h *productHandler) GetMovements(
	w http.ResponseWriter, r *http.Request,
) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	movements, err := h.repo.GetStockMovement(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}

		log.Printf("failed to get stock movements: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]any{
		"data": movements,
	}); err != nil {
		log.Printf("failed to encode stock movements: %v", err)
	}
}
