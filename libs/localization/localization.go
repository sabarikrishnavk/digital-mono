package localization

import (
	"context"
	"fmt"
)

// LocationalisationService defines the interface for getting location data.
type LocationalisationService interface {
	GetLatLngFromAddress(ctx context.Context, address, city, state, country, postcode string) (latitude, longitude float64, err error)
}

// DummyLocationalisationService is a placeholder implementation.
type DummyLocationalisationService struct{}

// NewDummyLocationalisationService creates a new dummy service.
func NewDummyLocationalisationService() *DummyLocationalisationService {
	return &DummyLocationalisationService{}
}

// GetLatLngFromAddress simulates getting latitude and longitude.
// In a real application, this would call an external geocoding API.
func (s *DummyLocationalisationService) GetLatLngFromAddress(ctx context.Context, address, city, state, country, postcode string) (latitude, longitude float64, err error) {
	// Simulate some logic based on input, or just return fixed values/errors
	fullAddress := fmt.Sprintf("%s, %s, %s, %s %s", address, city, state, postcode, country)
	fmt.Printf("Simulating geocoding for: %s\n", fullAddress) // Log the address being processed

	// Simple dummy logic: return different coords based on city
	switch city {
	case "Sydney":
		return -33.8688, 151.2093, nil // Sydney coordinates
	case "Melbourne":
		return -37.8136, 144.9631, nil // Melbourne coordinates
	case "Brisbane":
		return -27.4698, 153.0251, nil // Brisbane coordinates
	default:
		// Return some default or error
		return -34.0, 151.0, nil // Default dummy coordinates
	}
}