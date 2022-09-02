package user

import (
	"bridge/api/v1/pb"
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type repo struct {
	db *sqlx.DB
}

const (
	_findByID          = `SELECT * FROM users WHERE id = '' AND deleted_at IS NULL`
	_findByEmail       = `SELECT * FROM users WHERE email = '' AND deleted_at IS NULL`
	_findByPhoneNumber = `SELECT * FROM users WHERE phone_number = '' AND deleted_at IS NULL`
	_create            = `
	INSERT INTO users (name, email, phone_number, password, meta, account_status, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`
)

func (r *repo) Create(ctx context.Context, user *pb.User) error {
	var id uint64

	err := r.db.QueryRowxContext(
		ctx,
		_create,
		user.Name,
		user.Email,
		user.PhoneNumber,
		user.Password,
		user.Meta,
		user.AccountStatus,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("user: create - %w", err)
	}

	user.ID = id
	return nil
}

func (r *repo) Find(ctx context.Context, user *pb.User) error {
	var (
		query string
		args  any
	)
	if user.ID != 0 {
		query, args = _findByID, user.ID
	}
	if user.Email != "" {
		query, args = _findByEmail, user.Email
	}
	if user.PhoneNumber != "" {
		query, args = _findByPhoneNumber, user.PhoneNumber
	}

	if query == "" {
		return fmt.Errorf("user: invalid search parameter provided")
	}

	if err := r.db.GetContext(ctx, &user, query, args); err != nil {
		return fmt.Errorf("user: find - %w", err)
	}

	return nil
}

func NewRepo(db *sqlx.DB) *repo {
	return &repo{db: db}
}
