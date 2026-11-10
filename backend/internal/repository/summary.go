package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Summary struct {
	TotalProducts    int   `json:"total_products"`
	ActiveProducts   int   `json:"active_products"`
	StockValue       int64 `json:"stock_value"`
	LowStockProducts int   `json:"low_stock_products"`
}

type SummaryRepository struct {
	db *pgxpool.Pool
}

func NewSummaryRepository(db *pgxpool.Pool) *SummaryRepository {
	return &SummaryRepository{db: db}
}

func (r *SummaryRepository) GetSummary(
	ctx context.Context,
) (Summary, error) {
	var summary Summary

	err := r.db.QueryRow(ctx, `
        SELECT
            COUNT(*),
            COUNT(*) FILTER (WHERE is_active = TRUE),
            COALESCE(SUM(price::BIGINT * stock::BIGINT), 0),
            COUNT(*) FILTER (WHERE stock <= 25)
        FROM products
    `).Scan(
		&summary.TotalProducts,
		&summary.ActiveProducts,
		&summary.StockValue,
		&summary.LowStockProducts,
	)

	if err != nil {
		return Summary{}, fmt.Errorf("get dashboard summary: %w", err)
	}

	return summary, nil
}
