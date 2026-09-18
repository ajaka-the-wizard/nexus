package repositories

import (
	"auth/internal/errs"
	"auth/internal/models"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func InitRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) CreateUser(ctx context.Context, user *models.RegisterRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	query := `
	INSERT INTO users (full_name,email,password)
	VALUES ($1, $2, $3)
	`
	_, err := r.pool.Exec(ctx, query, user.FullName, user.Email, user.Password)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return errs.ERR_DUPLICATE_EMAIL
		}
		return err
	}
	return nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	query := `
	SELECT id, email, password, verified
	FROM users
	WHERE email = $1
	`
	rows, err := r.pool.Query(ctx, query, email)
	if err != nil {
		return nil, err
	}
	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[models.User])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ERR_EMAIL_NO_EXISTS
		}
		return nil, err
	}

	return &user, nil
}
