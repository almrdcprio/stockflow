package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

type SalesOrder struct {
	ID          int              `json:"id"`
	Customer    string           `json:"customer"`
	Status      string           `json:"status"`
	Total       int              `json:"total"`
	CreatedAt   time.Time        `json:"created_at"`
	CancelledAt *time.Time       `json:"cancelled_at"`
	Items       []SalesOrderItem `json:"items,omitempty"`
}

type SalesOrderItem struct {
	ID        int    `json:"id"`
	OrderID   int    `json:"order_id"`
	SKU       string `json:"sku"`
	Qty       int    `json:"qty"`
	UnitPrice int    `json:"unit_price"`
}

type CreateOrderItem struct {
	SKU string
	Qty int
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

var (
	ErrInvalidOrder          = errors.New("invalid order")
	ErrInactiveProduct       = errors.New("inactive product")
	ErrDuplicateOrderItem    = errors.New("duplicate order item")
	ErrOrderNotFound         = errors.New("order not found")
	ErrOrderAlreadyCancelled = errors.New("order cancelled")
)

func (r *OrderRepository) CreateOrder(
	ctx context.Context,
	customer string,
	items []CreateOrderItem,
) (*SalesOrder, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	sort.Slice(items, func(i, j int) bool {
		return items[i].SKU < items[j].SKU
	})

	total := 0
	order := &SalesOrder{
		Customer: customer,
		Status:   "CREATED",
		Items:    make([]SalesOrderItem, 0, len(items)),
	}

	for i, item := range items {
		if i > 0 && items[i-1].SKU == item.SKU {
			return nil, ErrDuplicateOrderItem
		}

		var unitPrice int
		var stock int
		var isActive bool

		err := tx.QueryRow(ctx, `
            SELECT price, stock, is_active
            FROM products
            WHERE sku = $1
            FOR UPDATE
        `, item.SKU).Scan(&unitPrice, &stock, &isActive)

		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, ErrProductNotFound
			}
			return nil, fmt.Errorf("get product: %w", err)
		}

		if !isActive {
			return nil, ErrInactiveProduct
		}

		if stock < item.Qty {
			return nil, ErrInsufficientStock
		}

		_, err = tx.Exec(ctx, `
            UPDATE products
            SET stock = stock - $1
            WHERE sku = $2
              AND is_active = TRUE
              AND stock >= $1
        `, item.Qty, item.SKU)
		if err != nil {
			return nil, fmt.Errorf("decrease product stock: %w", err)
		}

		_, err = tx.Exec(ctx, `
            INSERT INTO stock_movements (product_id, type, qty, note)
            SELECT id, 'out', $1, $2
            FROM products
            WHERE sku = $3
        `, item.Qty, "Sales order", item.SKU)
		if err != nil {
			return nil, fmt.Errorf("create stock movement: %w", err)
		}

		subtotal := unitPrice * item.Qty
		total += subtotal
		order.Items = append(order.Items, SalesOrderItem{
			SKU:       item.SKU,
			Qty:       item.Qty,
			UnitPrice: unitPrice,
		})
	}

	err = tx.QueryRow(ctx, `
        INSERT INTO sales_orders (customer, status, total)
        VALUES ($1, 'CREATED', $2)
        RETURNING id, created_at
    `, customer, total).Scan(&order.ID, &order.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create sales order: %w", err)
	}

	for i := range order.Items {
		err := tx.QueryRow(ctx, `
            INSERT INTO sales_order_items (order_id, sku, qty, unit_price)
            VALUES ($1, $2, $3, $4)
            RETURNING id
        `, order.ID, order.Items[i].SKU,
			order.Items[i].Qty, order.Items[i].UnitPrice,
		).Scan(&order.Items[i].ID)

		if err != nil {
			return nil, fmt.Errorf("create sales order item: %w", err)
		}

		order.Items[i].OrderID = order.ID
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit order: %w", err)
	}

	return order, nil
}

func (r *OrderRepository) GetOrder(
	ctx context.Context, id int,
) (*SalesOrder, error) {
	order := &SalesOrder{
		Items: make([]SalesOrderItem, 0),
	}

	var cancelledAt *time.Time

	err := r.db.QueryRow(ctx, `
        SELECT id, customer, status, total, created_at, cancelled_at
        FROM sales_orders
        WHERE id = $1
		`,
		id).Scan(
		&order.ID,
		&order.Customer,
		&order.Status,
		&order.Total,
		&order.CreatedAt,
		&cancelledAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}

		return nil, fmt.Errorf("get order: %w", err)
	}

	order.CancelledAt = cancelledAt

	rows, err := r.db.Query(ctx, `
        SELECT id, order_id, sku, qty, unit_price
        FROM sales_order_items
        WHERE order_id = $1
        ORDER BY id
    `, id)

	if err != nil {
		return nil, fmt.Errorf("get order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item SalesOrderItem

		if err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.SKU,
			&item.Qty,
			&item.UnitPrice,
		); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}

		order.Items = append(order.Items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order items: %w", err)
	}

	return order, nil
}

func (r *OrderRepository) CancelOrder(
	ctx context.Context, id int,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var status string

	err = tx.QueryRow(ctx, `
        SELECT status
        FROM sales_orders
        WHERE id = $1
        FOR UPDATE
    `, id).Scan(&status)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrOrderNotFound
		}

		return fmt.Errorf("get order status: %w", err)
	}

	if status == "CANCELLED" {
		return ErrOrderAlreadyCancelled
	}

	rows, err := tx.Query(ctx, `
        SELECT sku, qty
        FROM sales_order_items
        WHERE order_id = $1
        ORDER BY sku
    `, id)

	if err != nil {
		return fmt.Errorf("get order items: %w", err)
	}

	items := make([]CreateOrderItem, 0)

	for rows.Next() {
		var item CreateOrderItem

		if err := rows.Scan(&item.SKU, &item.Qty); err != nil {
			rows.Close()
			return fmt.Errorf("scan order item: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("iterate order items: %w", err)
	}

	rows.Close()

	for _, item := range items {
		_, err := tx.Exec(ctx, `
            UPDATE products
            SET stock = stock + $1
            WHERE sku = $2
        `, item.Qty, item.SKU)

		if err != nil {
			return fmt.Errorf("restore product stock: %w", err)
		}

		_, err = tx.Exec(ctx, `
            INSERT INTO stock_movements (product_id, type, qty, note)
            SELECT id, 'in', $1, $2
            FROM products
            WHERE sku = $3
        `, item.Qty, "Sales order cancellation", item.SKU)

		if err != nil {
			return fmt.Errorf("create stock movement: %w", err)
		}
	}

	_, err = tx.Exec(ctx, `
        UPDATE sales_orders
        SET status = 'CANCELLED',
            cancelled_at = NOW()
        WHERE id = $1
    `, id)

	if err != nil {
		return fmt.Errorf("cancel sales order: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
