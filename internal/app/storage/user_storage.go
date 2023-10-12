package storage

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v4"

	"github.com/belamov/ypgo-password-manager/internal/app/models"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"
)

type UsersRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(ctx context.Context, dsn string) (*UsersRepository, error) {
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return &UsersRepository{
		pool: pool,
	}, nil
}

func (repo *UsersRepository) CreateNew(username string, hashedPassword string) (*models.User, error) {
	user := &models.User{
		Username:       username,
		HashedPassword: hashedPassword,
	}

	conn, err := repo.pool.Acquire(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("couldnt acquire connection from pool")
		return nil, err
	}

	err = conn.QueryRow(
		context.Background(),
		"insert into users (username, hashed_password) values ($1, $2) returning id",
		user.Username,
		user.HashedPassword,
	).Scan(&user.Id)

	conn.Release()

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return nil, NewNotUniqueUsernameError(username, err)
		}
	}
	return user, err
}

func (repo *UsersRepository) Find(username string) (*models.User, error) {
	var user models.User

	conn, err := repo.pool.Acquire(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("couldnt acquire connection from pool")
		return nil, err
	}

	err = conn.QueryRow(
		context.Background(),
		"select id, username, hashed_password from users where username=$1",
		username,
	).Scan(&user.Id, &user.Username, &user.HashedPassword)

	conn.Release()

	if err == pgx.ErrNoRows {
		return nil, NewUserNotFoundError(username, err)
	}

	return &user, err
}
