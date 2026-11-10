package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"stockflow/internal/repository"

	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	repo *repository.OrderRepository
}

func NewOrderHandler(repo *repository.OrderRepository) *OrderHandler {
	return &OrderHandler{
		repo: repo,
	}
}

type CreateOrderRequest struct {
	Customer string                   `json:"customer"`
	Items    []CreateOrderItemRequest `json:"items"`
}

type CreateOrderItemRequest struct {
	SKU string `json:"sku"`
	Qty int    `json:"qty"`
}

func (h *OrderHandler) CreateOrder(
	w http.ResponseWriter, r *http.Request,
) {
	var req CreateOrderRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	req.Customer = strings.TrimSpace(req.Customer)

	if req.Customer == "" {
		http.Error(w, "customer are required", http.StatusBadRequest)
		return
	}

	if len(req.Items) == 0 {
		http.Error(w, "Items are required", http.StatusBadRequest)
		return
	}

	items := make([]repository.CreateOrderItem, 0, len(req.Items))

	for _, item := range req.Items {
		item.SKU = strings.TrimSpace(item.SKU)

		if item.SKU == "" || item.Qty <= 0 {
			http.Error(w, "invalid order item", http.StatusBadRequest)
			return
		}

		items = append(items, repository.CreateOrderItem{
			SKU: item.SKU,
			Qty: item.Qty,
		})
	}

	order, err := h.repo.CreateOrder(
		r.Context(),
		req.Customer,
		items,
	)

	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			respondError(w, http.StatusBadRequest, "PRODUCT_NOT_FOUND", "Product not found")
			return
		}

		if errors.Is(err, repository.ErrInsufficientStock) {
			respondError(w, http.StatusConflict, "INSUFFICIENT_STOCK", "Insufficient stock")
			return
		}

		if errors.Is(err, repository.ErrInactiveProduct) {
			respondError(w, http.StatusBadRequest, "PRODUCT_INACTIVE", "Product is inactive")
			return
		}

		if errors.Is(err, repository.ErrDuplicateOrderItem) {
			respondError(w, http.StatusBadRequest, "DUPLICATE_ORDER_ITEM", "Duplicate SKU in order")
			return
		}

		log.Printf("failed to create order: %v", err)
		respondError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error")
		return
	}

	respondJSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) GetOrder(
	w http.ResponseWriter, r *http.Request,
) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "INVALID_ORDER_ID", "Invalid order ID")
		return
	}

	order, err := h.repo.GetOrder(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			respondError(w, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found")
			return
		}

		log.Printf("failed to get order: %v", err)
		respondError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error")
		return
	}

	respondJSON(w, http.StatusOK, order)
}

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "INVALID_ORDER_ID", "Invalid order ID")
		return
	}

	err = h.repo.CancelOrder(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			respondError(w, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found")
			return
		}

		if errors.Is(err, repository.ErrOrderAlreadyCancelled) {
			respondError(w, http.StatusConflict, "ORDER_CANCELLED", "Order already cancelled")
			return
		}

		log.Printf("failed to cancel order: %v", err)
		respondError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
