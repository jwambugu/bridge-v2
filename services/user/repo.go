package user

import (
	"bridge/api/v1/pb"
	"bridge/core/db"
	"bridge/core/models"
	"bridge/core/repository"
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"time"
)

type repo struct {
	db *sqlx.DB
}

const (
	_findByID = `
	SELECT id,
		   name,
		   email,
		   phone_number,
		   account_status,
		   meta,
		   created_at,
		   updated_at
	FROM users
	WHERE id = $1
	  AND deleted_at IS NULL`

	_findByEmail = `SELECT id, email, password FROM users WHERE email = $1 AND deleted_at IS NULL`

	_create = `
	INSERT INTO users (name, email, phone_number, password, account_status, meta, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
)

func (r *repo) Authenticate(ctx context.Context, email string) (*pb.User, error) {
	stmt, err := r.db.PrepareContext(ctx, _findByEmail)
	if err != nil {
		return nil, fmt.Errorf("authenticate prepare: %w", err)
	}

	user := &pb.User{}
	if err = stmt.QueryRowContext(ctx, email).Scan(&user.ID, &user.Email, &user.Password); err != nil {
		return nil, fmt.Errorf("authenticate query: %w", err)
	}

	return user, nil
}

func (r *repo) Create(ctx context.Context, user *pb.User) error {
	var id uint64

	meta := &models.UserMeta{UserMeta: user.Meta}
	stmt, err := r.db.PreparexContext(ctx, _create)
	if err != nil {
		return fmt.Errorf("create user prepare: %w", err)
	}

	err = stmt.QueryRowxContext(
		ctx,
		user.Name,
		user.Email,
		user.PhoneNumber,
		user.Password,
		user.AccountStatus,
		meta,
		user.CreatedAt.AsTime(),
		user.UpdatedAt.AsTime(),
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("create user query: %w", err)
	}

	user.ID = id
	return nil
}

func (r *repo) Find(ctx context.Context, id uint64) (*pb.User, error) {
	if id == 0 {
		return nil, fmt.Errorf("id cannot be empty")
	}

	stmt, err := r.db.PrepareContext(ctx, _findByID)
	if err != nil {
		return nil, fmt.Errorf("find user prepare: %w", err)
	}

	var (
		u                    = &pb.User{}
		meta                 = &models.UserMeta{}
		createdAt, updatedAt time.Time
	)

	err = stmt.QueryRowContext(ctx, id).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PhoneNumber,
		&u.AccountStatus,
		&meta,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("find user query: %w", err)
	}

	u.CreatedAt = timestamppb.New(createdAt)
	u.UpdatedAt = timestamppb.New(updatedAt)
	u.Meta = meta.UserMeta
	return u, nil
}

func NewTestRepo(users ...*pb.User) repository.User {
	conn, err := db.NewConnection()
	if err != nil {
		log.Fatalf("user repo test - %v", err)
	}

	var (
		ctx = context.Background()
		r   = NewRepo(conn)
	)

	for _, user := range users {
		if err = r.Create(ctx, user); err != nil {
			log.Fatalf("failed to create user: %v", err)
		}
	}

	return r
}

func NewRepo(db *sqlx.DB) *repo {
	return &repo{db: db}
}
