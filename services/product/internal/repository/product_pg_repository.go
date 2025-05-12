package repository

import (
	"context"
	"database/sql"

	"github.com/omni-compos/digital-mono/services/product/internal/domain"
)

// ProductRepository defines the interface for product data operations.
type ProductRepository interface {
	CreateProduct(ctx context.Context, product *domain.Product) error
	GetProductByID(ctx context.Context, id string) (*domain.Product, error)
}

type pgProductRepository struct {
	db *sql.DB
}

// NewPGProductRepository creates a new PostgreSQL product repository.
func NewPGProductRepository(db *sql.DB) ProductRepository {
	return &pgProductRepository{db: db}
}

func (r *pgProductRepository) CreateProduct(ctx context.Context, product *domain.Product) error {
	query := `INSERT INTO products (id, name, description, sku, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, product.ID, product.Name, product.Description, product.SKU, product.CreatedAt, product.UpdatedAt)
	return err
}

func (r *pgProductRepository) GetProductByID(ctx context.Context, id string) (*domain.Product, error) {
	product := &domain.Product{}
	query := `SELECT id, name, description, sku, created_at, updated_at FROM products WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(&product.ID, &product.Name, &product.Description, &product.SKU, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Or a custom domain.ErrNotFound
		}
		return nil, err
	}
	return product, nil
}

// TODO: Add methods for UpdateProduct, DeleteProduct, ListProducts, etc.
// TODO: Ensure you have a `products` table in your PostgreSQL database.
// CREATE TABLE products (
//     id VARCHAR(36) PRIMARY KEY,
//     name VARCHAR(255) NOT NULL,
//     description TEXT,
//     sku VARCHAR(100) UNIQUE NOT NULL,
//     created_at TIMESTAMP WITH TIME ZONE NOT NULL,
//     updated_at TIMESTAMP WITH TIME ZONE NOT NULL
// );