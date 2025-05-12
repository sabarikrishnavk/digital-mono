package repository

import (
	"context"
	"database/sql"

	"github.com/omni-compos/digital-mono/services/user/internal/domain"
)

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
}

type pgUserRepository struct {
	db *sql.DB
}

// NewPGUserRepository creates a new PostgreSQL user repository.
func NewPGUserRepository(db *sql.DB) UserRepository {
	return &pgUserRepository{db: db}
}

func (r *pgUserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, name, email, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Name, user.Email, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *pgUserRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Or a custom domain.ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

// TODO: Add methods for UpdateUser, DeleteUser, ListUsers, etc.
// TODO: Ensure you have a `users` table in your PostgreSQL database.
// CREATE TABLE users (
//     id VARCHAR(36) PRIMARY KEY,
//     name VARCHAR(255) NOT NULL,
//     email VARCHAR(255) UNIQUE NOT NULL,
//     created_at TIMESTAMP WITH TIME ZONE NOT NULL,
//     updated_at TIMESTAMP WITH TIME ZONE NOT NULL
// );