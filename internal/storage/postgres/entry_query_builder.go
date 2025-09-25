// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package postgres // import "miniflux.app/v2/internal/storage/postgres"

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/querybuilder"
	"miniflux.app/v2/internal/timezone"
)

// CountEntries count the number of entries that match the condition.
func (p *Postgres) CountEntries(e *querybuilder.EntryQueryBuilder) (count int, err error) {
	query := `
		SELECT count(*)
		FROM entries e
			JOIN feeds f ON f.id = e.feed_id
			JOIN categories c ON c.id = f.category_id
		WHERE ` + e.BuildCondition()

	err = p.db.QueryRow(query, e.Args()...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("store: unable to count entries: %v", err)
	}

	return count, nil
}

// GetEntries returns a list of entries that match the condition.
func (p *Postgres) GetEntries(e *querybuilder.EntryQueryBuilder) (model.Entries, error) {
	query := `
		SELECT
			e.id,
			e.user_id,
			e.feed_id,
			e.hash,
			e.published_at at time zone u.timezone,
			e.title,
			e.url,
			e.comments_url,
			e.author,
			e.share_code,
			e.content,
			e.status,
			e.starred,
			e.reading_time,
			e.created_at,
			e.changed_at,
			e.tags,
			f.title as feed_title,
			f.feed_url,
			f.site_url,
			f.description,
			f.checked_at,
			f.category_id,
			c.title as category_title,
			c.hide_globally as category_hidden,
			f.scraper_rules,
			f.rewrite_rules,
			f.crawler,
			f.user_agent,
			f.cookie,
			f.hide_globally,
			f.no_media_player,
			f.webhook_url,
			fi.icon_id,
			i.external_id AS icon_external_id,
			u.timezone
		FROM
			entries e
		LEFT JOIN
			feeds f ON f.id=e.feed_id
		LEFT JOIN
			categories c ON c.id=f.category_id
		LEFT JOIN
			feed_icons fi ON fi.feed_id=f.id
		LEFT JOIN
			icons i ON i.id=fi.icon_id
		LEFT JOIN
			users u ON u.id=e.user_id
		WHERE ` + e.BuildCondition() + " " + e.BuildSorting()

	rows, err := p.db.Query(query, e.Args()...)
	if err != nil {
		return nil, fmt.Errorf("store: unable to get entries: %v", err)
	}
	defer rows.Close()

	entries := make(model.Entries, 0)
	entryMap := make(map[int64]*model.Entry)
	var entryIDs []int64

	for rows.Next() {
		var iconID sql.NullInt64
		var externalIconID sql.NullString
		var tz string

		entry := model.NewEntry()

		err := rows.Scan(
			&entry.ID,
			&entry.UserID,
			&entry.FeedID,
			&entry.Hash,
			&entry.Date,
			&entry.Title,
			&entry.URL,
			&entry.CommentsURL,
			&entry.Author,
			&entry.ShareCode,
			&entry.Content,
			&entry.Status,
			&entry.Starred,
			&entry.ReadingTime,
			&entry.CreatedAt,
			&entry.ChangedAt,
			pq.Array(&entry.Tags),
			&entry.Feed.Title,
			&entry.Feed.FeedURL,
			&entry.Feed.SiteURL,
			&entry.Feed.Description,
			&entry.Feed.CheckedAt,
			&entry.Feed.Category.ID,
			&entry.Feed.Category.Title,
			&entry.Feed.Category.HideGlobally,
			&entry.Feed.ScraperRules,
			&entry.Feed.RewriteRules,
			&entry.Feed.Crawler,
			&entry.Feed.UserAgent,
			&entry.Feed.Cookie,
			&entry.Feed.HideGlobally,
			&entry.Feed.NoMediaPlayer,
			&entry.Feed.WebhookURL,
			&iconID,
			&externalIconID,
			&tz,
		)

		if err != nil {
			return nil, fmt.Errorf("store: unable to fetch entry row: %v", err)
		}

		if iconID.Valid && externalIconID.Valid && externalIconID.String != "" {
			entry.Feed.Icon.FeedID = entry.FeedID
			entry.Feed.Icon.IconID = iconID.Int64
			entry.Feed.Icon.ExternalIconID = externalIconID.String
		} else {
			entry.Feed.Icon.IconID = 0
		}

		// Make sure that timestamp fields contain timezone information (API)
		entry.Date = timezone.Convert(tz, entry.Date)
		entry.CreatedAt = timezone.Convert(tz, entry.CreatedAt)
		entry.ChangedAt = timezone.Convert(tz, entry.ChangedAt)
		entry.Feed.CheckedAt = timezone.Convert(tz, entry.Feed.CheckedAt)

		entry.Feed.ID = entry.FeedID
		entry.Feed.UserID = entry.UserID
		entry.Feed.Icon.FeedID = entry.FeedID
		entry.Feed.Category.UserID = entry.UserID

		entries = append(entries, entry)
		entryMap[entry.ID] = entry
		entryIDs = append(entryIDs, entry.ID)
	}

	if e.FetchEnclosures() && len(entryIDs) > 0 {
		enclosures, err := p.GetEnclosuresForEntries(entryIDs)
		if err != nil {
			return nil, fmt.Errorf("store: unable to fetch enclosures: %w", err)
		}

		for entryID, entryEnclosures := range enclosures {
			if entry, exists := entryMap[entryID]; exists {
				entry.Enclosures = entryEnclosures
			}
		}
	}

	return entries, nil
}

// GetEntry returns a single entry that match the condition.
func (p *Postgres) GetEntry(e *querybuilder.EntryQueryBuilder) (*model.Entry, error) {
	e.WithLimit(1)
	entries, err := p.GetEntries(e)
	if err != nil {
		return nil, err
	}

	if len(entries) != 1 {
		return nil, nil
	}

	entries[0].Enclosures, err = p.GetEnclosures(entries[0].ID)
	if err != nil {
		return nil, err
	}

	return entries[0], nil
}

// GetEntryIDs returns a list of entry IDs that match the condition.
func (p *Postgres) GetEntryIDs(e *querybuilder.EntryQueryBuilder) ([]int64, error) {
	query := `
		SELECT
			e.id
		FROM
			entries e
		LEFT JOIN
			feeds f
		ON
			f.id=e.feed_id
		WHERE ` + e.BuildCondition() + " " + e.BuildSorting()

	rows, err := p.db.Query(query, e.Args()...)
	if err != nil {
		return nil, fmt.Errorf("store: unable to get entries: %v", err)
	}
	defer rows.Close()

	var entryIDs []int64
	for rows.Next() {
		var entryID int64

		err := rows.Scan(&entryID)
		if err != nil {
			return nil, fmt.Errorf("store: unable to fetch entry row: %v", err)
		}

		entryIDs = append(entryIDs, entryID)
	}

	return entryIDs, nil
}
