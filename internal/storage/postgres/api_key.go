// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package postgres // import "miniflux.app/v2/internal/storage/postgres"

import (
	"errors"
	"fmt"

	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
)

var ErrAPIKeyNotFound = errors.New("store: API Key not found")

// APIKeyExists checks if an API Key with the same description exists.
func (p *Postgres) APIKeyExists(userID int64, description string) bool {
	var result bool
	query := `SELECT true FROM api_keys WHERE user_id=$1 AND lower(description)=lower($2) LIMIT 1`
	p.db.QueryRow(query, userID, description).Scan(&result)
	return result
}

// SetAPIKeyUsedTimestamp updates the last used date of an API Key.
func (p *Postgres) SetAPIKeyUsedTimestamp(userID int64, token string) error {
	query := `UPDATE api_keys SET last_used_at=now() WHERE user_id=$1 and token=$2`
	_, err := p.db.Exec(query, userID, token)
	if err != nil {
		return fmt.Errorf(`store: unable to update last used date for API key: %v`, err)
	}

	return nil
}

// APIKeys returns all API Keys that belongs to the given user.
func (p *Postgres) APIKeys(userID int64) (model.APIKeys, error) {
	query := `
		SELECT
			id, user_id, token, description, last_used_at, created_at
		FROM
			api_keys
		WHERE
			user_id=$1
		ORDER BY description ASC
	`
	rows, err := p.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch API Keys: %v`, err)
	}
	defer rows.Close()

	apiKeys := make(model.APIKeys, 0)
	for rows.Next() {
		var apiKey model.APIKey
		if err := rows.Scan(
			&apiKey.ID,
			&apiKey.UserID,
			&apiKey.Token,
			&apiKey.Description,
			&apiKey.LastUsedAt,
			&apiKey.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf(`store: unable to fetch API Key row: %v`, err)
		}

		apiKeys = append(apiKeys, apiKey)
	}

	return apiKeys, nil
}

// CreateAPIKey inserts a new API key.
func (p *Postgres) CreateAPIKey(userID int64, description string) (*model.APIKey, error) {
	query := `
		INSERT INTO api_keys
			(user_id, token, description)
		VALUES
			($1, $2, $3)
		RETURNING
			id, user_id, token, description, last_used_at, created_at
	`
	var apiKey model.APIKey
	err := p.db.QueryRow(
		query,
		userID,
		crypto.GenerateRandomStringHex(32),
		description,
	).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.Token,
		&apiKey.Description,
		&apiKey.LastUsedAt,
		&apiKey.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to create API Key: %v`, err)
	}

	return &apiKey, nil
}

// DeleteAPIKey deletes an API Key.
func (p *Postgres) DeleteAPIKey(userID, keyID int64) error {
	result, err := p.db.Exec(`DELETE FROM api_keys WHERE id = $1 AND user_id = $2`, keyID, userID)
	if err != nil {
		return fmt.Errorf(`store: unable to delete this API Key: %v`, err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(`store: unable to delete this API Key: %v`, err)
	}

	if count == 0 {
		return storage.ErrAPIKeyNotFound
	}

	return nil
}
