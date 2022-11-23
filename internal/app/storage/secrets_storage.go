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

type SecretsRepository struct {
	pool *pgxpool.Pool
}

func NewSecretsRepository(ctx context.Context, dsn string) (*SecretsRepository, error) {
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return &SecretsRepository{
		pool: pool,
	}, nil
}

func (repo *SecretsRepository) CreateNew(encryptedData []byte, metadata models.SecretMetadata) error {
	conn, err := repo.pool.Acquire(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("couldnt acquire connection from pool")
		return err
	}

	err = conn.QueryRow(
		context.Background(),
		"insert into secrets (secret_data, user_id, secret_type, secret_name) values ($1, $2, $3, $4)",
		encryptedData,
		metadata.UserID,
		metadata.Type,
		metadata.Name,
	).Scan()

	conn.Release()

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return NewNotUniqueSecret(metadata, err)
		}
	}
	return nil
}

func (repo *SecretsRepository) FindSecretData(metadata models.SecretMetadata) ([]byte, error) {
	var data []byte

	conn, err := repo.pool.Acquire(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("couldnt acquire connection from pool")
		return nil, err
	}

	err = conn.QueryRow(
		context.Background(),
		"select secret_data from secrets where secret_type=$1 and user_id=$2 and secret_name=$3",
		metadata.Type,
		metadata.UserID,
		metadata.Name,
	).Scan(&data)

	conn.Release()

	if err == pgx.ErrNoRows {
		return nil, NewSecretNotFoundError(metadata, err)
	}

	return data, err
}
