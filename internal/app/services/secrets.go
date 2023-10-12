package services

import (
	"bytes"
	"encoding/gob"
	"github.com/belamov/ypgo-password-manager/internal/app/models"
	"github.com/belamov/ypgo-password-manager/internal/app/storage"
	"github.com/rs/zerolog/log"
)

type SecretsManagerInterface interface {
	SaveSecret(encodedSecret []byte, metadata models.SecretMetadata) error
	GetPassword(metadata models.SecretMetadata) (*models.PasswordSecret, error)
	GetCard(metadata models.SecretMetadata) (*models.CardSecret, error)
	GetText(metadata models.SecretMetadata) (*models.TextSecret, error)
}

type SecretsManager struct {
	secretsRepo   storage.SecretsStorage
	cryptographer Cryptographer
}

func (s *SecretsManager) GetPassword(metadata models.SecretMetadata) (*models.PasswordSecret, error) {
	secret := &models.PasswordSecret{}
	encodedSecret, err := s.GetSecret(metadata)
	dec := gob.NewDecoder(bytes.NewReader(encodedSecret))
	err = dec.Decode(&secret)
	if err != nil {
		log.Error().Err(err).Msg("cant decode secret")
		return nil, err
	}
	return secret, err
}

func (s *SecretsManager) GetCard(metadata models.SecretMetadata) (*models.CardSecret, error) {
	secret := &models.CardSecret{}
	encodedSecret, err := s.GetSecret(metadata)
	dec := gob.NewDecoder(bytes.NewReader(encodedSecret))
	err = dec.Decode(&secret)
	if err != nil {
		log.Error().Err(err).Msg("cant decode secret")
		return nil, err
	}
	return secret, err
}

func (s *SecretsManager) GetText(metadata models.SecretMetadata) (*models.TextSecret, error) {
	secret := &models.TextSecret{}
	encodedSecret, err := s.GetSecret(metadata)
	dec := gob.NewDecoder(bytes.NewReader(encodedSecret))
	err = dec.Decode(&secret)
	if err != nil {
		log.Error().Err(err).Msg("cant decode secret")
		return nil, err
	}
	return secret, err
}

func NewSecretsService(secretsStorage storage.SecretsStorage, cryptographer Cryptographer) *SecretsManager {
	return &SecretsManager{
		secretsRepo:   secretsStorage,
		cryptographer: cryptographer,
	}
}

func (s *SecretsManager) SaveSecret(encodedSecret []byte, metadata models.SecretMetadata) error {
	encryptedData, err := s.cryptographer.Encrypt(encodedSecret)
	if err != nil {
		log.Error().Err(err).Msg("cant encrypt secret")
		return err
	}

	err = s.secretsRepo.CreateNew(encryptedData, metadata)
	if err != nil {
		log.Error().Err(err).Msg("cant save secret")
		return err
	}

	return nil
}

func (s *SecretsManager) GetSecret(metadata models.SecretMetadata) ([]byte, error) {
	encryptedData, err := s.secretsRepo.FindSecretData(metadata)
	if err != nil {
		log.Error().Err(err).Msg("cant get secret from storage")
		return nil, err
	}

	decryptedData, err := s.cryptographer.Decrypt(encryptedData)
	if err != nil {
		log.Error().Err(err).Msg("cant decrypt data")
		return nil, err
	}

	return decryptedData, nil
}
