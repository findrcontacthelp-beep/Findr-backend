package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/findr-app/findr-backend/internal/model"
	authmodel "github.com/findr-app/findr-backend/internals/model"
)

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrEmailAlreadyExists = errors.New("email already exists")
var ErrInvalidCredentials = errors.New("invalid credentials")

type UserRepository interface {
	CreateUser(ctx context.Context, req authmodel.RegisterRequest) (*model.User, error)
	AuthenticateUser(ctx context.Context, email, password string) (*model.User, error)
}

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, req authmodel.RegisterRequest) (*model.User, error) {
	email := normalizeEmail(req.Email)
	if emailExists, err := r.emailExists(ctx, email); err != nil {
		return nil, err
	} else if emailExists {
		return nil, ErrEmailAlreadyExists
	}

	var passwordHash *string
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		hash := string(hashed)
		passwordHash = &hash
	}

	userUUID := uuid.New().String()

	var u model.User
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (user_uuid, name, email, password_hash, college_name, college_stream, college_year, interests)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, user_uuid, name, email, college_name, college_stream, college_year, interests, created_at, updated_at`,
		userUUID, req.Name, email, passwordHash, req.CollegeName, req.Branch, req.GraduationYear, req.Interests,
	).Scan(&u.ID, &u.UserUUID, &u.Name, &u.Email, &u.CollegeName, &u.CollegeStream, &u.CollegeYear, &u.Interests, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	return &u, nil
}

func (r *PostgresUserRepository) AuthenticateUser(ctx context.Context, email, password string) (*model.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_uid, name, email, college_name, college_stream, college_year, interests, created_at, updated_at, password_hash
		 FROM users
		 WHERE lower(email) = $1
		 ORDER BY created_at ASC
		 LIMIT 2`,
		normalizeEmail(email),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matchedUsers int
	var u model.User
	var passwordHash *string
	for rows.Next() {
		matchedUsers++
		if matchedUsers > 1 {
			return nil, ErrEmailAlreadyExists
		}
		if err := rows.Scan(&u.ID, &u.UserUUID, &u.Name, &u.Email, &u.CollegeName, &u.CollegeStream, &u.CollegeYear, &u.Interests, &u.CreatedAt, &u.UpdatedAt, &passwordHash); err != nil {
			return nil, err
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if matchedUsers == 0 || passwordHash == nil || *passwordHash == "" {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*passwordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return &u, nil
}

func (r *PostgresUserRepository) emailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1
			FROM users
			WHERE lower(email) = $1
		)`,
		email,
	).Scan(&exists)
	return exists, err
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
