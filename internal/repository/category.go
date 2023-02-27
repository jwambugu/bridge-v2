package repository

import (
	"bridge/api/v1/pb"
	"bridge/internal/db"
	"bridge/internal/models"
	"bridge/internal/utils"
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"time"
)

type Category interface {
	All(ctx context.Context) ([]*pb.Category, error)
	Create(ctx context.Context, category *pb.Category) error
	FindByID(ctx context.Context, id string) (*pb.Category, error)
	FindBySlug(ctx context.Context, slug string) (*pb.Category, error)
	Update(ctx context.Context, category *pb.Category) error
}

type categoryRepo struct {
	db *sqlx.DB
	l  zerolog.Logger
}

const (
	_categoryBaseSelect = `SELECT id, name, slug, status, meta, created_at, updated_at FROM categories `
	_categoryAll        = _categoryBaseSelect + `WHERE deleted_at IS NULL`
	_categoryFindByID   = _categoryBaseSelect + `WHERE id = $1 AND deleted_at IS NULL`
	_categoryFindBySlug = _categoryBaseSelect + `WHERE slug = $1 AND deleted_at IS NULL`

	_categoryCreate = `INSERT INTO categories(name, slug, status, meta, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	_categoryUpdateCategory = `
	UPDATE categories
	SET name       = $1,
		slug       = $2,
		status     = $3,
		meta       = $4,
		updated_at = $5
	WHERE id = $6`
)

func (r *categoryRepo) scanRow(row scanner) (*pb.Category, error) {
	l := r.l.With().Str("action", "scan row").Logger()

	var (
		c                    = &pb.Category{}
		meta                 = &models.CategoryMeta{}
		createdAt, updatedAt time.Time
	)

	err := row.Scan(
		&c.ID,
		&c.Name,
		&c.Slug,
		&c.Status,
		&meta,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		l.Err(err).Msg("scan row")
		return nil, err
	}

	c.CreatedAt = timestamppb.New(createdAt)
	c.UpdatedAt = timestamppb.New(updatedAt)
	c.Meta = meta.CategoryMeta

	l.Info().Str("category_id", c.ID)
	return c, nil
}

func (r *categoryRepo) All(ctx context.Context) ([]*pb.Category, error) {
	l := r.l.With().Str("action", "get all categories").Str("query", _categoryAll).Logger()

	rows, err := r.db.QueryContext(ctx, _categoryAll)
	if err != nil {
		l.Err(err).Msg("query rows")
		return nil, err
	}

	var categories []*pb.Category
	for rows.Next() {
		c, err := r.scanRow(rows)
		if err != nil {
			l.Err(err).Msg("scan row")
			return nil, err
		}

		categories = append(categories, c)
	}

	return categories, nil
}

func (r *categoryRepo) Create(ctx context.Context, category *pb.Category) error {
	l := r.l.With().Str("action", "create").
		Interface("category", fmt.Sprintf("%+v", category)).
		Str("query", _categoryCreate).
		Logger()

	stmt, err := r.db.PreparexContext(ctx, _categoryCreate)
	if err != nil {
		l.Err(err).Msg("prepare statement")
		return err
	}

	var id string
	meta := &models.CategoryMeta{CategoryMeta: category.Meta}

	err = stmt.QueryRowxContext(ctx,
		category.Name,
		category.Slug,
		category.Status,
		meta,
		category.CreatedAt.AsTime(),
		category.UpdatedAt.AsTime(),
	).Scan(&id)

	if err != nil {
		l.Err(err).Msg("scan row")
		return err
	}

	l.Info().Str("id", id).Msg("completed successfully")
	category.ID = id
	return nil
}

func (r *categoryRepo) FindByID(ctx context.Context, id string) (*pb.Category, error) {
	l := r.l.With().Str("action", "find by id").Str("id", id).Str("query", _categoryFindByID).Logger()

	stmt, err := r.db.PrepareContext(ctx, _categoryFindByID)
	if err != nil {
		l.Err(err).Msg("prepare statement")
		return nil, err
	}

	category, err := r.scanRow(stmt.QueryRowContext(ctx, id))
	if err != nil {
		l.Err(err).Msg("scan row")
		return nil, err
	}
	return category, nil
}

func (r *categoryRepo) FindBySlug(ctx context.Context, slug string) (*pb.Category, error) {
	l := r.l.With().Str("action", "find by slug").Str("slug", slug).Str("query", _categoryFindByID).Logger()

	stmt, err := r.db.PrepareContext(ctx, _categoryFindBySlug)
	if err != nil {
		l.Err(err).Msg("prepare statement")
		return nil, err
	}

	category, err := r.scanRow(stmt.QueryRowContext(ctx, slug))
	if err != nil {
		l.Err(err).Msg("scan row")
		return nil, err
	}
	return category, nil
}

func (r *categoryRepo) Update(ctx context.Context, category *pb.Category) error {

	var (
		l = r.l.With().Str("action", "update").
			Interface("category", fmt.Sprintf("%+v", category)).
			Str("query", _categoryUpdateCategory).Logger()

		meta = &models.CategoryMeta{
			CategoryMeta: category.Meta,
		}
	)

	stmt, err := r.db.PrepareContext(ctx, _categoryUpdateCategory)
	if err != nil {
		l.Err(err).Msg("prepare statement")
		return err
	}

	category.UpdatedAt = timestamppb.New(time.Now())
	category.Slug = utils.Slugify(category.Name)

	_, err = stmt.ExecContext(
		ctx,
		category.Name,
		category.Slug,
		category.Status,
		meta,
		category.UpdatedAt.AsTime(),
		category.ID,
	)
	if err != nil {
		l.Err(err).Msg("exec query")
		return err
	}

	return nil
}

func NewTestCategoryRepo(l zerolog.Logger, categories ...*pb.Category) Category {
	conn, err := db.NewConnection()
	if err != nil {
		log.Fatalf("user repo test - %v", err)
	}

	var (
		ctx = context.Background()
		r   = NewCategoryRepo(conn, l)
	)

	for _, category := range categories {
		if err = r.Create(ctx, category); err != nil {
			log.Fatalf("failed to create category: %v", err)
		}
	}

	return r
}

func NewCategoryRepo(db *sqlx.DB, l zerolog.Logger) Category {
	return &categoryRepo{
		db: db,
		l:  l.With().Str("repo", "category_sqlx").Logger(),
	}
}
