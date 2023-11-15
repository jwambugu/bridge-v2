package repository

import (
	"bridge/api/v1/pb"
	"bridge/internal/db"
	"bridge/internal/logger"
	"bridge/internal/models"
	"bridge/internal/rpc_error"
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type User interface {
	Authenticate(ctx context.Context, email string) (*pb.User, error)
	Create(ctx context.Context, user *pb.User) error
	Exists(ctx context.Context, user *pb.User) error
	FindByEmail(ctx context.Context, email string) (*pb.User, error)
	FindByID(ctx context.Context, id string) (*pb.User, error)
	FindByPhoneNumber(ctx context.Context, phoneNumber string) (*pb.User, error)
	Update(ctx context.Context, user *pb.User) error
}

type userRepo struct {
	db *sqlx.DB
	l  zerolog.Logger
}

const (
	_userBaseSelect        = `SELECT id, name, email, phone_number, account_status, meta, created_at, updated_at FROM users `
	_userFindByID          = _userBaseSelect + `WHERE id = $1 AND deleted_at IS NULL`
	_userFindByEmail       = _userBaseSelect + `WHERE email = $1 AND deleted_at IS NULL`
	_userFindByPhoneNumber = _userBaseSelect + `WHERE phone_number = $1 AND deleted_at IS NULL`

	_userAuthenticateByEmail = `SELECT id, email, password FROM users WHERE email = $1 AND deleted_at IS NULL`

	_userCreate = `
	INSERT INTO users (name, email, phone_number, password, account_status, meta, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	_userUpdate = `
	UPDATE users
	SET name           = $1,
		email          = $2,
		phone_number   = $3,
		meta           = $4,
		account_status = $5,
		updated_at     = $6
	WHERE id = $7`
)

var userRepoExistsQueries = map[db.UserTblColumn]string{
	db.UserEmail:       `SELECT exists( SELECT 1 FROM users WHERE email = $1)`,
	db.UserPhoneNumber: `SELECT exists( SELECT 1 FROM users WHERE phone_number = $1)`,
}

func (r *userRepo) scanRow(row *sql.Row) (*pb.User, error) {
	l := r.l.With().Str("action", "scan row").Logger()

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
		l.Err(err).Msg("scan row")
		return nil, err
	}

	u.CreatedAt = timestamppb.New(createdAt)
	u.UpdatedAt = timestamppb.New(updatedAt)
	u.Meta = meta.UserMeta

	l.Info().Str("user_id", u.ID)
	return u, nil
}

func (r *userRepo) Authenticate(ctx context.Context, email string) (*pb.User, error) {
	l := r.l.With().Str("action", "authenticate").
		Str("email", email).
		Str("query", _userAuthenticateByEmail).
		Logger()

	stmt, err := r.db.PrepareContext(ctx, _userAuthenticateByEmail)
	if err != nil {
		l.Err(err).Msg("prepare statement")
		return nil, err
	}

	user := &pb.User{}
	if err = stmt.QueryRowContext(ctx, email).Scan(&user.ID, &user.Email, &user.Password); err != nil {
		l.Err(err).Msg("scan row")
		return nil, err
	}

	l.Info().Str("user_id", user.ID).Msg("completed successfully")
	return user, nil
}

func (r *userRepo) Create(ctx context.Context, user *pb.User) error {
	l := r.l.With().Str("action", "create").
		Interface("user", fmt.Sprintf("%+v", user)).
		Str("query", _userCreate).
		Logger()

	stmt, err := r.db.PreparexContext(ctx, _userCreate)
	if err != nil {
		l.Err(err).Msg("prepare statement")
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
		l.Err(err).Msg("exec and scan result")
		return err
	}

	user.ID = id
	return nil
}

func (r *userRepo) Exists(ctx context.Context, user *pb.User) error {
	l := r.l.With().Str("action", "exists").Interface("user", fmt.Sprintf("%+v", user)).Logger()

	checks := map[db.UserTblColumn]struct {
		field string
		err   error
	}{
		db.UserEmail:       {field: user.Email, err: rpc_error.ErrEmailExists},
		db.UserPhoneNumber: {field: user.PhoneNumber, err: rpc_error.ErrPhoneNumberExists},
	}

	var exists bool
	for column, st := range checks {
		q, ok := userRepoExistsQueries[column]
		if !ok {
			return rpc_error.ErrServerError
		}

		l = l.With().Str("query", q).Logger()

		stmt, err := r.db.PreparexContext(ctx, q)
		if err != nil {
			l.Err(err).Msg("prepare statement")
			return rpc_error.ErrServerError
		}

		if err = stmt.QueryRowxContext(ctx, st.field).Scan(&exists); err != nil {
			l.Err(err).Msg("scan row")
			return rpc_error.ErrServerError
		}

		if exists {
			l.Err(st.err).Msg("row exists error")
			return st.err
		}
	}

	l.Info().Msg("record doesn't exist")
	return nil
}

func (r *userRepo) FindByID(ctx context.Context, id string) (*pb.User, error) {
	l := r.l.With().Str("action", "find by id").
		Str("id", id).
		Str("query", _userFindByID).
		Logger()

	stmt, err := r.db.PrepareContext(ctx, _userFindByID)
	if err != nil {
		l.Err(err).Msg("prepare statement")
		return nil, err
	}

	u, err := r.scanRow(stmt.QueryRowContext(ctx, id))
	if err != nil {
		l.Err(err).Msg("scan row")
		return nil, err
	}

	l.Info().Str("id", id).Msg("completed successfully")
	return u, nil
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*pb.User, error) {
	l := r.l.With().Str("action", "find by email").
		Str("email", email).
		Str("query", _userFindByEmail).
		Logger()

	stmt, err := r.db.PrepareContext(ctx, _userFindByEmail)
	if err != nil {
		l.Err(err).Msg("prepare statement")
		return nil, err
	}

	u, err := r.scanRow(stmt.QueryRowContext(ctx, email))
	if err != nil {
		l.Err(err).Msg("scan row")
		return nil, err
	}

	l.Info().Str("id", u.ID).Msg("completed successfully")
	return u, nil
}

func (r *userRepo) FindByPhoneNumber(ctx context.Context, phoneNumber string) (*pb.User, error) {
	l := r.l.With().Str("action", "find by phone number").
		Str("phone_number", phoneNumber).
		Str("query", _userFindByPhoneNumber).
		Logger()

	stmt, err := r.db.PrepareContext(ctx, _userFindByPhoneNumber)
	if err != nil {
		l.Err(err).Msg("prepare statement")
		return nil, err
	}

	u, err := r.scanRow(stmt.QueryRowContext(ctx, phoneNumber))
	if err != nil {
		l.Err(err).Msg("scan row")
		return nil, err
	}

	l.Info().Str("id", u.ID).Msg("completed successfully")
	return u, nil
}

func (r *userRepo) Update(ctx context.Context, user *pb.User) error {
	l := r.l.With().Str("action", "user").
		Interface("user", fmt.Sprintf("%+v", user)).
		Str("query", _userUpdate).
		Logger()

	stmt, err := r.db.PrepareContext(ctx, _userUpdate)
	if err != nil {
		l.Err(err).Msg("prepare statement")
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
		l.Err(err).Msg("exec query")
		return err
	}

	l.Info().Str("id", user.ID).Msg("completed successfully")
	return nil
}

func NewTestUserRepo(ctx context.Context, db *sqlx.DB, users ...*pb.User) (User, error) {
	repo := NewUserRepo(db, logger.TestLogger)

	for _, user := range users {
		if err := repo.Create(ctx, user); err != nil {
			return nil, err
		}
	}

	return repo, nil
}

func NewUserRepo(db *sqlx.DB, l zerolog.Logger) User {
	return &userRepo{
		db: db,
		l:  l.With().Str("repo", "user_sqlx").Logger(),
	}
}
