package storage

import (
	"errors"
	"os"

	"github.com/belamov/ypgo-password-manager/internal/app/models"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
)

type UserStorage interface {
	CreateNew(username string, hashedPassword string) (*models.User, error)
	Find(username string) (*models.User, error)
}

type SecretsStorage interface {
	CreateNew(encryptedData []byte, metadata models.SecretMetadata) error
	FindSecretData(metadata models.SecretMetadata) ([]byte, error)
}

func RunMigrations(dsn string) error {
	m, err := migrate.New(getMigrationsPath(), dsn+"&x-migrations-table=migrations")
	if err != nil {
		return err
	}

	log.Info().Msg("Migrating...")

	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		log.Info().Msg("Nothing to migrate")
		return nil
	}
	if err != nil {
		log.Error().Err(err).Msg("Migration failed!")
		return err
	}
	log.Info().Msg("Migrated successfully")
	return nil
}

func getMigrationsPath() string {
	path := os.Getenv("MIGRATIONS_PATH")
	if path == "" {
		path = "file://internal/app/storage/migrations"
	}
	return path
}
