// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package postgres // import "miniflux.app/v2/internal/storage/postgres"

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/querybuilder"
)

// Entries returns previous and next entries.
func (p *Postgres) Entries(e *querybuilder.EntryPaginationBuilder) (*model.Entry, *model.Entry, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return nil, nil, fmt.Errorf("begin transaction for entry pagination: %v", err)
	}

	prevID, nextID, err := p.getPrevNextID(tx, e)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	prevEntry, err := p.getEntry(tx, prevID)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	nextEntry, err := p.getEntry(tx, nextID)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	tx.Commit()

	if e.Direction() == "desc" {
		return nextEntry, prevEntry, nil
	}

	return prevEntry, nextEntry, nil
}

func (p *Postgres) getPrevNextID(tx *sql.Tx, e *querybuilder.EntryPaginationBuilder) (prevID int64, nextID int64, err error) {
	cte := `
		WITH entry_pagination AS (
			SELECT
				e.id,
				lag(e.id) over (order by e.%[1]s asc, e.created_at asc, e.id desc) as prev_id,
				lead(e.id) over (order by e.%[1]s asc, e.created_at asc, e.id desc) as next_id
			FROM entries AS e
			JOIN feeds AS f ON f.id=e.feed_id
			JOIN categories c ON c.id = f.category_id
			WHERE %[2]s
			ORDER BY e.%[1]s asc, e.created_at asc, e.id desc
		)
		SELECT prev_id, next_id FROM entry_pagination AS ep WHERE %[3]s;
	`

	subCondition := strings.Join(e.Conditions(), " AND ")
	finalCondition := "ep.id = $" + strconv.Itoa(len(e.Args())+1)
	query := fmt.Sprintf(cte, e.Order(), subCondition, finalCondition)
	e.WithArg(e.EntryID())

	var pID, nID sql.NullInt64
	err = tx.QueryRow(query, e.Args()...).Scan(&pID, &nID)
	switch {
	case err == sql.ErrNoRows:
		return 0, 0, nil
	case err != nil:
		return 0, 0, fmt.Errorf("entry pagination: %v", err)
	}

	if pID.Valid {
		prevID = pID.Int64
	}

	if nID.Valid {
		nextID = nID.Int64
	}

	return prevID, nextID, nil
}

func (p *Postgres) getEntry(tx *sql.Tx, entryID int64) (*model.Entry, error) {
	var entry model.Entry

	err := tx.QueryRow(`SELECT id, title FROM entries WHERE id = $1`, entryID).Scan(
		&entry.ID,
		&entry.Title,
	)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("fetching sibling entry: %v", err)
	}

	return &entry, nil
}
