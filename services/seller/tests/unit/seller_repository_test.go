package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/omni-compos/digital-mono/services/seller/internal/domain"
	"github.com/omni-compos/digital-mono/services/seller/internal/repository"
)

// Helper function to create a mock DB and repository
func newMockRepository(t *testing.T) (*sql.DB, sqlmock.Sqlmock, repository.SellerRepository) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	repo := repository.NewPGSellerRepository(db)
	return db, mock, repo
}

func TestPGSellerRepository_CreateSeller(t *testing.T) {
	db, mock, repo := newMockRepository(t)
	defer db.Close()

	seller := model.NewSeller()
	seller.ID = "test-id-123"
	seller.BrandID = model.BrandIDBrandA
	seller.Status = model.StatusActive
	seller.Address = "1 Test St"
	seller.City = "Testville"
	seller.State = "TS"
	seller.Country = "AUS"
	seller.Postcode = "1000"
	seller.Email = "test@example.com"
	seller.PhoneNumber = "1234567890"
	seller.Latitude = -34.0
	seller.Longitude = 151.0
	seller.LastUpdatedBy = "user-abc"
	seller.LastUpdateTime = sql.NullTime{Time: time.Now(), Valid: true}.Time // Use sql.NullTime for scanning if needed, but model uses time.Time

	// Expect an INSERT statement
	mock.ExpectExec(`INSERT INTO sellers`).
		WithArgs(
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
		).WillReturnResult(sqlmock.NewResult(1, 1)) // Assume 1 row affected

	err := repo.CreateSeller(context.Background(), seller)
	if err != nil {
		t.Errorf("CreateSeller() error = %v", err)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPGSellerRepository_GetSellerByID(t *testing.T) {
	db, mock, repo := newMockRepository(t)
	defer db.Close()

	sellerID := "test-id-123"
	// Define the columns expected in the SELECT query
	columns := []string{"id", "brand_id", "status", "address", "city", "state", "country", "postcode", "email", "phone_number", "latitude", "longitude", "last_updated_by", "last_update_time"}
	expectedTime := time.Now().UTC().Truncate(time.Second) // Match time precision if needed

	// Expect a SELECT statement
	mock.ExpectQuery(`SELECT id, brand_id, status, address, city, state, country, postcode, email, phone_number, latitude, longitude, last_updated_by, last_update_time FROM sellers WHERE id = \$1`).
		WithArgs(sellerID).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(sellerID, model.BrandIDBrandA, model.StatusActive, "1 Test St", "Testville", "TS", "AUS", "1000", "test@example.com", "1234567890", -34.0, 151.0, "user-abc", expectedTime))

	seller, err := repo.GetSellerByID(context.Background(), sellerID)
	if err != nil {
		t.Errorf("GetSellerByID() error = %v", err)
	}

	if seller == nil || seller.ID != sellerID {
		t.Errorf("GetSellerByID() got seller = %v, want seller with ID %s", seller, sellerID)
	}
	// Add more assertions to check other fields

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// Add tests for UpdateSeller, DeleteSeller, ListSellers following the same pattern
// using mock.ExpectExec for UPDATE/DELETE and mock.ExpectQuery for SELECT.
// Remember to handle sql.ErrNoRows for not found cases.

func TestPGSellerRepository_GetSellerByID_NotFound(t *testing.T) {
	db, mock, repo := newMockRepository(t)
	defer db.Close()

	sellerID := "non-existent-id"

	// Expect a SELECT statement and return no rows
	mock.ExpectQuery(`SELECT id, brand_id, status, address, city, state, country, postcode, email, phone_number, latitude, longitude, last_updated_by, last_update_time FROM sellers WHERE id = \$1`).
		WithArgs(sellerID).
		WillReturnError(sql.ErrNoRows) // Simulate not found

	seller, err := repo.GetSellerByID(context.Background(), sellerID)
	if err != nil && err != sql.ErrNoRows { // Expecting sql.ErrNoRows or nil if repo handles it
		t.Errorf("GetSellerByID() error = %v, want sql.ErrNoRows", err)
	}

	if seller != nil {
		t.Errorf("GetSellerByID() got seller = %v, want nil", seller)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}