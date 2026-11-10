package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository struct {
	db *pgxpool.Pool
}

var ErrDuplicateSKU = errors.New("Products with the same SKU already exists")
var ErrProductNotFound = errors.New("Product not found")
var ErrProductHasHistory = errors.New("Product has stock movement history and cannot be deleted")

type Product struct {
	ID             int        `json:"id"`
	SKU            string     `json:"sku"`
	Name           string     `json:"name"`
	Price          int        `json:"price"`
	Stock          int        `json:"stock"`
	IsActive       bool       `json:"is_active"`
	LastMovementAt *time.Time `json:"last_movement_at"`
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

func (r *ProductRepository) GetProducts(
	ctx context.Context,
	search string,
	page int,
	limit int,
) ([]Product, int, error) {
	offset := (page - 1) * limit

	query := `
	select
	p.id, 
	p.sku, 
	p.name, 
	p.price, 
	p.stock, 
	p.is_active, 
	MAX(sm.created_at) AS last_movement_at
	from products p
	left join stock_movements sm on p.id = sm.product_id
	where (
		$1 = ''
		or p.sku ilike '%' || $1 || '%'
		or p.name ilike '%' || $1 || '%'
	)
	group by p.id
	order by p.id asc
	limit $2 offset $3
	`
	rows, err := r.db.Query(ctx, query, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	products := make([]Product, 0)
	for rows.Next() {
		var p Product

		err := rows.Scan(
			&p.ID,
			&p.SKU,
			&p.Name,
			&p.Price,
			&p.Stock,
			&p.IsActive,
			&p.LastMovementAt,
		)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var totalCount int
	countQuery := `
		select count(*) from products p
		where (
			$1 = ''
			or p.sku ilike '%' || $1 || '%'
			or p.name ilike '%' || $1 || '%'
		)
	`

	err = r.db.QueryRow(ctx, countQuery, search).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	return products, totalCount, nil
}

func (r *ProductRepository) CreateProduct(
	ctx context.Context,
	sku string,
	name string,
	price int,
) (Product, error) {
	query := `
	insert into products (sku, name, price)
	values ($1, $2, $3)
	returning
		id, 
		sku, 
		name, 
		price, 
		stock, 
		is_active
	`

	var p Product
	err := r.db.QueryRow(
		ctx,
		query,
		sku,
		name,
		price,
	).Scan(
		&p.ID,
		&p.SKU,
		&p.Name,
		&p.Price,
		&p.Stock,
		&p.IsActive,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "products_sku_key" {
			return Product{}, ErrDuplicateSKU
		}

		return Product{}, fmt.Errorf("failed to create product: %w", err)
	}

	return p, nil

}

func (r *ProductRepository) UpdateProduct(
	ctx context.Context,
	id int,
	sku string,
	name string,
	price int,
) (Product, error) {
	query := `
		update products
		set
			sku = $1,
			name = $2,
			price = $3
		where id = $4
		returning
			id,
			sku,
			name,
			price,
			stock,
			is_active,
			(
				select max(sm.created_at)
				from stock_movements sm
				where sm.product_id = products.id
			) as last_movement_at
	`

	var p Product
	err := r.db.QueryRow(
		ctx,
		query,
		sku,
		name,
		price,
		id,
	).Scan(
		&p.ID,
		&p.SKU,
		&p.Name,
		&p.Price,
		&p.Stock,
		&p.IsActive,
		&p.LastMovementAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Product{}, ErrProductNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "products_sku_key" {
			return Product{}, ErrDuplicateSKU
		}
		return Product{}, fmt.Errorf("failed to update product: %w", err)
	}

	return p, nil
}

func (r *ProductRepository) DeleteProduct(
	ctx context.Context,
	id int,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var sku string
	err = tx.QueryRow(
		ctx,
		`
		select sku
		from products
		where id = $1
		for update
		`,
		id,
	).Scan(&sku)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProductNotFound
		}
		return fmt.Errorf("failed to check product existence: %w", err)
	}

	var hasHistory bool
	err = tx.QueryRow(
		ctx,
		`
		select exists (
			select 1
			from stock_movements
			where product_id = $1
		)
		or exists (
			select 1
			from sales_order_items
			where sku = $2
		)
		`,
		id,
		sku,
	).Scan(&hasHistory)

	if err != nil {
		return fmt.Errorf("failed to check product history: %w", err)
	}

	if hasHistory {
		return ErrProductHasHistory
	}

	_, err = tx.Exec(
		ctx,
		`
		delete from products
		where id = $1
		`,
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
