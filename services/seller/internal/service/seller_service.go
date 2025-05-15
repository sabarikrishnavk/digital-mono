package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/omni-compos/digital-mono/libs/localization"
	"github.com/omni-compos/digital-mono/libs/logger"

	model "github.com/omni-compos/digital-mono/services/seller/internal/domain"
	"github.com/omni-compos/digital-mono/services/seller/internal/repository"
)

// SellerService defines the interface for seller business logic.
type SellerService interface {
	CreateSeller(ctx context.Context, seller *model.Seller, userID string) (*model.Seller, error)
	GetSellerByID(ctx context.Context, id string) (*model.Seller, error)
	UpdateSeller(ctx context.Context, id string, updates *model.Seller, userID string) (*model.Seller, error)
	DeleteSeller(ctx context.Context, id string) error
	ListSellers(ctx context.Context, limit, offset int) ([]*model.Seller, error)
}

// DefaultSellerService is the default implementation of SellerService.
type DefaultSellerService struct {
	repo              repository.SellerRepository
	localization	  localization.LocationalisationService
	logger            logger.Logger
}

// NewProductService creates a new DefaultSellerService.
func NewSellerService(repo repository.SellerRepository, localization localization.LocationalisationService, logger logger.Logger) *DefaultSellerService {
	return &DefaultSellerService{
		repo:              repo,
		localization: 	   localization,
		logger:            logger,
	}
}

// CreateSeller handles the creation of a new seller.
func (s *DefaultSellerService) CreateSeller(ctx context.Context, seller *model.Seller, userID string) (*model.Seller, error) {
	// Validate static lists
	if !isValidBrandID(seller.BrandID) {
		return nil, fmt.Errorf("invalid brand ID: %s", seller.BrandID)
	}
	if !isValidStatus(seller.Status) {
		return nil, fmt.Errorf("invalid status: %s", seller.Status)
	}

	// Get Lat/Lng from address using locationalisation service
	lat, lng, err := s.localization.GetLatLngFromAddress(ctx, seller.Address, seller.City, seller.State, seller.Country, seller.Postcode)
	if err != nil {
		s.logger.Error(err, "Failed to get lat/lng for seller", "address", seller.Address)
		// Decide if this should be a hard error or if you proceed without coords
		// For now, let's return an error
		return nil, fmt.Errorf("failed to geocode address: %w", err)
	}
	seller.Latitude = lat
	seller.Longitude = lng

	// Set audit fields
	seller.ID = uuid.New().String() // Generate a new UUID for the seller
	seller.LastUpdatedBy = userID
	seller.LastUpdateTime = time.Now()

	// Save to repository
	err = s.repo.CreateSeller(ctx, seller)
	if err != nil {
		s.logger.Error(err, "Failed to create seller in repository")
		return nil, fmt.Errorf("failed to save seller: %w", err)
	}

	s.logger.Info("Seller created successfully", "seller_id", seller.ID, "updated_by", userID)
	return seller, nil
}

// GetSellerByID retrieves a seller by ID.
func (s *DefaultSellerService) GetSellerByID(ctx context.Context, id string) (*model.Seller, error) {
	seller, err := s.repo.GetSellerByID(ctx, id)
	if err != nil {
		s.logger.Error(err, "Failed to get seller by ID from repository", "seller_id", id)
		return nil, fmt.Errorf("failed to retrieve seller: %w", err)
	}
	if seller == nil {
		return nil, nil // Seller not found
	}
	return seller, nil
}

// UpdateSeller handles updating an existing seller.
func (s *DefaultSellerService) UpdateSeller(ctx context.Context, id string, updates *model.Seller, userID string) (*model.Seller, error) {
	existingSeller, err := s.repo.GetSellerByID(ctx, id)
	if err != nil {
		s.logger.Error(err, "Failed to get existing seller for update", "seller_id", id)
		return nil, fmt.Errorf("failed to retrieve seller for update: %w", err)
	}
	if existingSeller == nil {
		return nil, fmt.Errorf("seller with ID %s not found", id)
	}

	// Apply updates (only fields that are allowed to be updated)
	// This is a simplified approach; a real implementation might merge fields carefully
	if updates.BrandID != "" {
		if !isValidBrandID(updates.BrandID) {
			return nil, fmt.Errorf("invalid brand ID: %s", updates.BrandID)
		}
		existingSeller.BrandID = updates.BrandID
	}
	if updates.Status != "" {
		if !isValidStatus(updates.Status) {
			return nil, fmt.Errorf("invalid status: %s", updates.Status)
		}
		existingSeller.Status = updates.Status
	}
	if updates.Address != "" {
		existingSeller.Address = updates.Address
	}
	if updates.City != "" {
		existingSeller.City = updates.City
	}
	if updates.State != "" {
		existingSeller.State = updates.State
	}
	if updates.Country != "" {
		existingSeller.Country = updates.Country
	}
	if updates.Postcode != "" {
		existingSeller.Postcode = updates.Postcode
	}
	if updates.Email != "" {
		existingSeller.Email = updates.Email
	}
	if updates.PhoneNumber != "" {
		existingSeller.PhoneNumber = updates.PhoneNumber
	}

	// Re-geocode if address fields changed
	// A more robust check would compare old vs new address fields
	lat, lng, err := s.localization.GetLatLngFromAddress(ctx, existingSeller.Address, existingSeller.City, existingSeller.State, existingSeller.Country, existingSeller.Postcode)
	if err != nil {
		s.logger.Error(err, "Failed to re-geocode address for seller update", "seller_id", id)
		// Decide if this should block the update or just skip updating coords
		// For now, let's return an error
		return nil, fmt.Errorf("failed to geocode address for update: %w", err)
	}
	existingSeller.Latitude = lat
	existingSeller.Longitude = lng

	// Update audit fields
	existingSeller.LastUpdatedBy = userID
	existingSeller.LastUpdateTime = time.Now()

	// Save updates to repository
	err = s.repo.UpdateSeller(ctx, existingSeller)
	if err != nil {
		s.logger.Error(err, "Failed to update seller in repository", "seller_id", id)
		return nil, fmt.Errorf("failed to save seller updates: %w", err)
	}

	s.logger.Info("Seller updated successfully", "seller_id", id, "updated_by", userID)
	return existingSeller, nil
}

// DeleteSeller handles deleting a seller.
func (s *DefaultSellerService) DeleteSeller(ctx context.Context, id string) error {
	err := s.repo.DeleteSeller(ctx, id)
	if err != nil {
		s.logger.Error(err, "Failed to delete seller from repository", "seller_id", id)
		return fmt.Errorf("failed to delete seller: %w", err)
	}
	s.logger.Info("Seller deleted successfully", "seller_id", id)
	return nil
}

// ListSellers handles retrieving a list of sellers.
func (s *DefaultSellerService) ListSellers(ctx context.Context, limit, offset int) ([]*model.Seller, error) {
	sellers, err := s.repo.ListSellers(ctx, limit, offset)
	if err != nil {
		s.logger.Error(err, "Failed to list sellers from repository")
		return nil, fmt.Errorf("failed to list sellers: %w", err)
	}
	return sellers, nil
}

// Helper functions for validation
func isValidBrandID(brandID string) bool {
	for _, b := range model.ValidBrandIDs {
		if b == brandID {
			return true
		}
	}
	return false
}

func isValidStatus(status string) bool {
	for _, s := range model.ValidStatuses {
		if s == status {
			return true
		}
	}
	return false
}