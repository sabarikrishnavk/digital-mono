package domain

import (
	"time"
)

// Static lists for BrandID and Status (example values)
const (
	BrandIDBrandA = "BRAND_A"
	BrandIDBrandB = "BRAND_B"
	BrandIDBrandC = "BRAND_C"
)

var ValidBrandIDs = []string{BrandIDBrandA, BrandIDBrandB, BrandIDBrandC}

const (
	StatusActive   = "ACTIVE"
	StatusInactive = "INACTIVE"
	StatusPending  = "PENDING"
)

var ValidStatuses = []string{StatusActive, StatusInactive, StatusPending}

// Seller represents the seller entity.
type Seller struct {
	ID            string     `json:"id"` // Assuming a unique ID, maybe UUID
	BrandID       string     `json:"brandId"`
	Status        string     `json:"status"`
	Address       string     `json:"address"`
	City          string     `json:"city"`
	State         string     `json:"state"`
	Country       string     `json:"country"` // Default to AUS
	Postcode      string     `json:"postcode"`
	Email         string     `json:"email"`
	PhoneNumber   string     `json:"phoneNumber"`
	Latitude      float64    `json:"latitude"`
	Longitude     float64    `json:"longitude"`
	LastUpdatedBy string     `json:"lastUpdatedBy"` // User ID from JWT
	LastUpdateTime time.Time `json:"lastUpdateTime"`
}

// NewSeller creates a new Seller instance with default values.
func NewSeller() *Seller {
	return &Seller{
		Country: "AUS", // Default country
	}
}