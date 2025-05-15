package repository

import (
	"context"
	"database/sql"
	"fmt"

	model "github.com/omni-compos/digital-mono/services/seller/internal/domain"
)

// SellerRepository defines the interface for seller data operations.
type SellerRepository interface {
	CreateSeller(ctx context.Context, seller *model.Seller) error
	GetSellerByID(ctx context.Context, id string) (*model.Seller, error)
	UpdateSeller(ctx context.Context, seller *model.Seller) error
	DeleteSeller(ctx context.Context, id string) error
	ListSellers(ctx context.Context, limit, offset int) ([]*model.Seller, error)
}

// PGSellerRepository is a PostgreSQL implementation of SellerRepository.
type PGSellerRepository struct {
	db *sql.DB
}

// NewPGSellerRepository creates a new PGSellerRepository.
func NewPGSellerRepository(db *sql.DB) *PGSellerRepository {
	return &PGSellerRepository{db: db}
}

// CreateSeller inserts a new seller into the database.
func (r *PGSellerRepository) CreateSeller(ctx context.Context, seller *model.Seller) error {
	// In a real implementation, you would execute an SQL INSERT statement here.
	// Example placeholder:
	query := `INSERT INTO sellers (id, brand_id, status, address, city, state, country, postcode, email, phone_number, latitude, longitude, last_updated_by, last_update_time)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	_, err := r.db.ExecContext(ctx, query,
		seller.ID,
		seller.BrandID,
		seller.Status,
		seller.Address,
		seller.City,
		seller.State,
		seller.Country,
		seller.Postcode,
		seller.Email,
		seller.PhoneNumber,
		seller.Latitude,
		seller.Longitude,
		seller.LastUpdatedBy,
		seller.LastUpdateTime,
	)
	if err != nil {
		// Log or wrap the error appropriately
		return fmt.Errorf("failed to create seller: %w", err)
	}
	return nil
}

// GetSellerByID retrieves a seller by their ID.
func (r *PGSellerRepository) GetSellerByID(ctx context.Context, id string) (*model.Seller, error) {
	// In a real implementation, you would execute an SQL SELECT statement here.
	// Example placeholder:
	query := `SELECT id, brand_id, status, address, city, state, country, postcode, email, phone_number, latitude, longitude, last_updated_by, last_update_time
              FROM sellers WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	seller := &model.Seller{}
	err := row.Scan(
		&seller.ID,
		&seller.BrandID,
		&seller.Status,
		&seller.Address,
		&seller.City,
		&seller.State,
		&seller.Country,
		&seller.Postcode,
		&seller.Email,
		&seller.PhoneNumber,
		&seller.Latitude,
		&seller.Longitude,
		&seller.LastUpdatedBy,
		&seller.LastUpdateTime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Seller not found
		}
		// Log or wrap the error appropriately
		return nil, fmt.Errorf("failed to get seller by ID %s: %w", id, err)
	}
	return seller, nil
}

// UpdateSeller updates an existing seller in the database.
func (r *PGSellerRepository) UpdateSeller(ctx context.Context, seller *model.Seller) error {
	// In a real implementation, you would execute an SQL UPDATE statement here.
	// Example placeholder:
	query := `UPDATE sellers
              SET brand_id = $2, status = $3, address = $4, city = $5, state = $6, country = $7, postcode = $8, email = $9, phone_number = $10, latitude = $11, longitude = $12, last_updated_by = $13, last_update_time = $14
              WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query,
		seller.ID,
		seller.BrandID,
		seller.Status,
		seller.Address,
		seller.City,
		seller.State,
		seller.Country,
		seller.Postcode,
		seller.Email,
		seller.PhoneNumber,
		seller.Latitude,
		seller.Longitude,
		seller.LastUpdatedBy,
		seller.LastUpdateTime,
	)
	if err != nil {
		return fmt.Errorf("failed to update seller %s: %w", seller.ID, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after update for seller %s: %w", seller.ID, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("seller with ID %s not found for update", seller.ID) // Or a specific "not found" error
	}
	return nil
}

// DeleteSeller deletes a seller by their ID.
func (r *PGSellerRepository) DeleteSeller(ctx context.Context, id string) error {
	// In a real implementation, you would execute an SQL DELETE statement here.
	// Example placeholder:
	query := `DELETE FROM sellers WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete seller %s: %w", id, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after delete for seller %s: %w", id, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("seller with ID %s not found for delete", id) // Or a specific "not found" error
	}
	return nil
}

// ListSellers retrieves a list of sellers with pagination.
func (r *PGSellerRepository) ListSellers(ctx context.Context, limit, offset int) ([]*model.Seller, error) {
	// In a real implementation, you would execute an SQL SELECT statement here.
	// Example placeholder:
	query := `SELECT id, brand_id, status, address, city, state, country, postcode, email, phone_number, latitude, longitude, last_updated_by, last_update_time
              FROM sellers LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list sellers: %w", err)
	}
	defer rows.Close()

	var sellers []*model.Seller
	for rows.Next() {
		seller := &model.Seller{}
		err := rows.Scan(
			&seller.ID,
			&seller.BrandID,
			&seller.Status,
			&seller.Address,
			&seller.City,
			&seller.State,
			&seller.Country,
			&seller.Postcode,
			&seller.Email,
			&seller.PhoneNumber,
			&seller.Latitude,
			&seller.Longitude,
			&seller.LastUpdatedBy,
			&seller.LastUpdateTime,
		)
		if err != nil {
			// Log the scanning error but continue processing other rows if possible,
			// or return the error depending on desired behavior.
			fmt.Printf("Error scanning seller row: %v\n", err)
			continue
		}
		sellers = append(sellers, seller)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error after iterating through seller rows: %w", err)
	}

	return sellers, nil
}