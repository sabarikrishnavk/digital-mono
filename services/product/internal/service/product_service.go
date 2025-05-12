package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/omni-compos/digital-mono/libs/logger"
	"github.com/omni-compos/digital-mono/services/product/internal/domain"
	"github.com/omni-compos/digital-mono/services/product/internal/repository"
)

// ProductService defines the interface for product business logic.
type ProductService interface {
	CreateProduct(ctx context.Context, name, description, sku string) (*domain.Product, error)
	GetProduct(ctx context.Context, id string) (*domain.Product, error)
}

type productService struct {
	repo   repository.ProductRepository
	logger logger.Logger
}

// NewProductService creates a new product service.
func NewProductService(repo repository.ProductRepository, log logger.Logger) ProductService {
	return &productService{repo: repo, logger: log}
}

func (s *productService) CreateProduct(ctx context.Context, name, description, sku string) (*domain.Product, error) {
	s.logger.Info("Creating product", "name", name, "sku", sku)
	now := time.Now()
	product := &domain.Product{
		ID:          uuid.NewString(),
		Name:        name,
		Description: description,
		SKU:         sku,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return product, s.repo.CreateProduct(ctx, product)
}

func (s *productService) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	s.logger.Info("Getting product", "id", id)
	return s.repo.GetProductByID(ctx, id)
}