package model_test

import (
	"testing"

	"github.com/omni-compos/digital-mono/services/seller/internal/domain"
)

func TestNewSeller(t *testing.T) {
	// Test that NewSeller initializes with default values
	seller := model.NewSeller()

	if seller.Country != "AUS" {
		t.Errorf("Expected default Country 'AUS', but got '%s'", seller.Country)
	}

	// Add more assertions for other default values if any
}

// Add tests for validating BrandID and Status if validation logic was in the model
// (Currently validation is in the service, which is fine)