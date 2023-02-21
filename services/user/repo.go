package user

import (
	"bridge/api/v1/pb"
	"bridge/internal/db"
	"bridge/internal/models"
	"bridge/internal/repository"
	"bridge/internal/rpc_error"
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"time"
)

type repo struct {
	db *sqlx.DB
}

const (
	_baseSelectQuery   = `SELECT id, name, email, phone_number, account_status, meta, created_at, updated_at FROM users `
	_findByID          = _baseSelectQuery + `WHERE id = $1 AND deleted_at IS NULL`
	_findByEmail       = _baseSelectQuery + `WHERE email = $1 AND deleted_at IS NULL`
	_findByPhoneNumber = _baseSelectQuery + `WHERE phone_number = $1 AND deleted_at IS NULL`

	_authenticateByEmail = `SELECT id, email, password FROM users WHERE email = $1 AND deleted_at IS NULL`

	_create = `
	INSERT INTO users (name, email, phone_number, password, account_status, meta, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	_updateUser = `
	UPDATE users
	SET name           = $1,
		email          = $2,
		phone_number   = $3,
		meta           = $4,
		account_status = $5,
		updated_at     = $6
	WHERE id = $7`
)

var existsQueriesMap = map[db.UserTblColumn]string{
	db.UserEmail:       `SELECT exists( SELECT 1 FROM users WHERE email = $1)`,
	db.UserPhoneNumber: `SELECT exists( SELECT 1 FROM users WHERE phone_number = $1)`,
}

func (r *repo) scanRow(row *sql.Row) (*pb.User, error) {
	var (
		u                    = &pb.User{}
		meta                 = &models.UserMeta{}
		createdAt, updatedAt time.Time
	)

	err := row.Scan(
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
		return nil, err
	}

	u.CreatedAt = timestamppb.New(createdAt)
	u.UpdatedAt = timestamppb.New(updatedAt)
	u.Meta = meta.UserMeta
	return u, nil
}

func (r *repo) Authenticate(ctx context.Context, email string) (*pb.User, error) {
	stmt, err := r.db.PrepareContext(ctx, _authenticateByEmail)
	if err != nil {
		return nil, err
	}

	user := &pb.User{}
	if err = stmt.QueryRowContext(ctx, email).Scan(&user.ID, &user.Email, &user.Password); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *repo) Create(ctx context.Context, user *pb.User) error {
	stmt, err := r.db.PreparexContext(ctx, _create)
	if err != nil {
		return err
	}

	var id string
	meta := &models.UserMeta{UserMeta: user.Meta}

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
		return err
	}

	user.ID = id
	return nil
}

func (r *repo) Exists(ctx context.Context, user *pb.User) error {
	checks := map[db.UserTblColumn]struct {
		field string
		err   error
	}{
		db.UserEmail:       {field: user.Email, err: rpc_error.ErrEmailExists},
		db.UserPhoneNumber: {field: user.PhoneNumber, err: rpc_error.ErrPhoneNumberExists},
	}

	var exists bool
	for column, st := range checks {
		q, ok := existsQueriesMap[column]
		if !ok {
			return rpc_error.ErrServerError
		}

		stmt, err := r.db.PreparexContext(ctx, q)
		if err != nil {
			return rpc_error.ErrServerError
		}

		if err = stmt.QueryRowxContext(ctx, st.field).Scan(&exists); err != nil {
			return rpc_error.ErrServerError
		}

		if exists {
			return st.err
		}
	}

	return nil
}

func (r *repo) FindByID(ctx context.Context, id string) (*pb.User, error) {
	stmt, err := r.db.PrepareContext(ctx, _findByID)
	if err != nil {
		return nil, err
	}

	u, err := r.scanRow(stmt.QueryRowContext(ctx, id))
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *repo) FindByEmail(ctx context.Context, email string) (*pb.User, error) {
	stmt, err := r.db.PrepareContext(ctx, _findByEmail)
	if err != nil {
		return nil, err
	}

	u, err := r.scanRow(stmt.QueryRowContext(ctx, email))
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *repo) FindByPhoneNumber(ctx context.Context, phoneNumber string) (*pb.User, error) {
	stmt, err := r.db.PrepareContext(ctx, _findByPhoneNumber)
	if err != nil {
		return nil, err
	}

	u, err := r.scanRow(stmt.QueryRowContext(ctx, phoneNumber))
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *repo) Update(ctx context.Context, user *pb.User) error {
	stmt, err := r.db.PrepareContext(ctx, _updateUser)
	if err != nil {
		return err
	}

	var (
		meta = &models.UserMeta{
			UserMeta: user.Meta,
		}
	)

	user.UpdatedAt = timestamppb.New(time.Now())

	if _, err = stmt.ExecContext(
		ctx,
		user.Name,
		user.Email,
		user.PhoneNumber,
		meta,
		user.AccountStatus,
		user.UpdatedAt.AsTime(),
		user.ID,
	); err != nil {
		return err
	}

	return nil
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

func NewRepo(db *sqlx.DB) repository.User {
	return &repo{
		db: db,
	}
}
