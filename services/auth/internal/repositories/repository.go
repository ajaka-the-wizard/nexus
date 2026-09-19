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
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	pool *pgxpool.Pool
}

const passwordResetCooldown = 10 * 24 * time.Hour

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

func (r *Repository) VerifyUser(ctx context.Context, email string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.pool.Exec(ctx, `
		UPDATE users
		SET verified = TRUE, active = TRUE
		WHERE email = $1 AND verified = FALSE AND active = FALSE
	`, email)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errs.ERR_EMAIL_NO_EXISTS
	}

	return nil
}

func (r *Repository) ResetPassword(ctx context.Context, email string, request *models.ResetPasswordRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var user models.User
	query := `
	SELECT id, password, password_updated_at
	FROM users
	WHERE email = $1
	FOR UPDATE
	`
	if err := tx.QueryRow(ctx, query, email).Scan(
		&user.Id,
		&user.Password,
		&user.PasswordUpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ERR_EMAIL_NO_EXISTS
		}
		return err
	}
	if time.Since(user.PasswordUpdatedAt) < passwordResetCooldown {
		return errs.ERR_PASSWORD_RESET_COOLDOWN
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.NewPassword)); err == nil {
		return errs.ERR_PASSWORD_REUSED
	} else if err != bcrypt.ErrMismatchedHashAndPassword {
		return err
	}

	historyRows, err := tx.Query(ctx, `
		SELECT password
		FROM password_history
		WHERE user_id = $1
	`, user.Id)
	if err != nil {
		return err
	}
	defer historyRows.Close()

	for historyRows.Next() {
		var passwordHash string
		if err := historyRows.Scan(&passwordHash); err != nil {
			return err
		}
		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(request.NewPassword)); err == nil {
			return errs.ERR_PASSWORD_REUSED
		} else if err != bcrypt.ErrMismatchedHashAndPassword {
			return err
		}
	}
	if err := historyRows.Err(); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO password_history (user_id, password, created_at)
		VALUES ($1, $2, $3)
	`, user.Id, user.Password, user.PasswordUpdatedAt)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		UPDATE users
		SET password = $1, password_updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, string(hashedPassword), user.Id)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
