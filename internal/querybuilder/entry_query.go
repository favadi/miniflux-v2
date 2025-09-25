// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package querybuilder // import "miniflux.app/v2/internal/querybuilder"

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

// EntryQueryBuilder builds a SQL query to fetch entries.
type EntryQueryBuilder struct {
	args            []any
	conditions      []string
	sortExpressions []string
	limit           int
	offset          int
	fetchEnclosures bool
}

// WithEnclosures fetches enclosures for each entry.
func (e *EntryQueryBuilder) WithEnclosures() *EntryQueryBuilder {
	e.fetchEnclosures = true
	return e
}

// WithSearchQuery adds full-text search query to the condition.
func (e *EntryQueryBuilder) WithSearchQuery(query string) *EntryQueryBuilder {
	if query != "" {
		nArgs := len(e.args) + 1
		e.conditions = append(e.conditions, fmt.Sprintf("e.document_vectors @@ plainto_tsquery($%d)", nArgs))
		e.args = append(e.args, query)

		// 0.0000001 = 0.1 / (seconds_in_a_day)
		e.WithSorting(
			fmt.Sprintf("ts_rank(document_vectors, plainto_tsquery($%d)) - extract (epoch from now() - published_at)::float * 0.0000001", nArgs),
			"DESC",
		)
	}
	return e
}

// WithStarred adds starred filter.
func (e *EntryQueryBuilder) WithStarred(starred bool) *EntryQueryBuilder {
	if starred {
		e.conditions = append(e.conditions, "e.starred is true")
	} else {
		e.conditions = append(e.conditions, "e.starred is false")
	}
	return e
}

// BeforeChangedDate adds a condition < changed_at
func (e *EntryQueryBuilder) BeforeChangedDate(date time.Time) *EntryQueryBuilder {
	e.conditions = append(e.conditions, "e.changed_at < $"+strconv.Itoa(len(e.args)+1))
	e.args = append(e.args, date)
	return e
}

// AfterChangedDate adds a condition > changed_at
func (e *EntryQueryBuilder) AfterChangedDate(date time.Time) *EntryQueryBuilder {
	e.conditions = append(e.conditions, "e.changed_at > $"+strconv.Itoa(len(e.args)+1))
	e.args = append(e.args, date)
	return e
}

// BeforePublishedDate adds a condition < published_at
func (e *EntryQueryBuilder) BeforePublishedDate(date time.Time) *EntryQueryBuilder {
	e.conditions = append(e.conditions, "e.published_at < $"+strconv.Itoa(len(e.args)+1))
	e.args = append(e.args, date)
	return e
}

// AfterPublishedDate adds a condition > published_at
func (e *EntryQueryBuilder) AfterPublishedDate(date time.Time) *EntryQueryBuilder {
	e.conditions = append(e.conditions, "e.published_at > $"+strconv.Itoa(len(e.args)+1))
	e.args = append(e.args, date)
	return e
}

// BeforeEntryID adds a condition < entryID.
func (e *EntryQueryBuilder) BeforeEntryID(entryID int64) *EntryQueryBuilder {
	if entryID != 0 {
		e.conditions = append(e.conditions, "e.id < $"+strconv.Itoa(len(e.args)+1))
		e.args = append(e.args, entryID)
	}
	return e
}

// AfterEntryID adds a condition > entryID.
func (e *EntryQueryBuilder) AfterEntryID(entryID int64) *EntryQueryBuilder {
	if entryID != 0 {
		e.conditions = append(e.conditions, "e.id > $"+strconv.Itoa(len(e.args)+1))
		e.args = append(e.args, entryID)
	}
	return e
}

// WithEntryIDs filter by entry IDs.
func (e *EntryQueryBuilder) WithEntryIDs(entryIDs []int64) *EntryQueryBuilder {
	e.conditions = append(e.conditions, fmt.Sprintf("e.id = ANY($%d)", len(e.args)+1))
	e.args = append(e.args, pq.Int64Array(entryIDs))
	return e
}

// WithEntryID filter by entry ID.
func (e *EntryQueryBuilder) WithEntryID(entryID int64) *EntryQueryBuilder {
	if entryID != 0 {
		e.conditions = append(e.conditions, "e.id = $"+strconv.Itoa(len(e.args)+1))
		e.args = append(e.args, entryID)
	}
	return e
}

// WithFeedID filter by feed ID.
func (e *EntryQueryBuilder) WithFeedID(feedID int64) *EntryQueryBuilder {
	if feedID > 0 {
		e.conditions = append(e.conditions, "e.feed_id = $"+strconv.Itoa(len(e.args)+1))
		e.args = append(e.args, feedID)
	}
	return e
}

// WithCategoryID filter by category ID.
func (e *EntryQueryBuilder) WithCategoryID(categoryID int64) *EntryQueryBuilder {
	if categoryID > 0 {
		e.conditions = append(e.conditions, "f.category_id = $"+strconv.Itoa(len(e.args)+1))
		e.args = append(e.args, categoryID)
	}
	return e
}

// WithStatus filter by entry status.
func (e *EntryQueryBuilder) WithStatus(status string) *EntryQueryBuilder {
	if status != "" {
		e.conditions = append(e.conditions, "e.status = $"+strconv.Itoa(len(e.args)+1))
		e.args = append(e.args, status)
	}
	return e
}

// WithStatuses filter by a list of entry statuses.
func (e *EntryQueryBuilder) WithStatuses(statuses []string) *EntryQueryBuilder {
	if len(statuses) > 0 {
		e.conditions = append(e.conditions, fmt.Sprintf("e.status = ANY($%d)", len(e.args)+1))
		e.args = append(e.args, pq.StringArray(statuses))
	}
	return e
}

// WithTags filter by a list of entry tags.
func (e *EntryQueryBuilder) WithTags(tags []string) *EntryQueryBuilder {
	if len(tags) > 0 {
		for _, cat := range tags {
			e.conditions = append(e.conditions, fmt.Sprintf("LOWER($%d) = ANY(LOWER(e.tags::text)::text[])", len(e.args)+1))
			e.args = append(e.args, cat)
		}
	}
	return e
}

// WithoutStatus set the entry status that should not be returned.
func (e *EntryQueryBuilder) WithoutStatus(status string) *EntryQueryBuilder {
	if status != "" {
		e.conditions = append(e.conditions, "e.status <> $"+strconv.Itoa(len(e.args)+1))
		e.args = append(e.args, status)
	}
	return e
}

// WithShareCode set the entry share code.
func (e *EntryQueryBuilder) WithShareCode(shareCode string) *EntryQueryBuilder {
	e.conditions = append(e.conditions, "e.share_code = $"+strconv.Itoa(len(e.args)+1))
	e.args = append(e.args, shareCode)
	return e
}

// WithShareCodeNotEmpty adds a filter for non-empty share code.
func (e *EntryQueryBuilder) WithShareCodeNotEmpty() *EntryQueryBuilder {
	e.conditions = append(e.conditions, "e.share_code <> ''")
	return e
}

// WithSorting add a sort expression.
func (e *EntryQueryBuilder) WithSorting(column, direction string) *EntryQueryBuilder {
	e.sortExpressions = append(e.sortExpressions, column+" "+direction)
	return e
}

// WithLimit set the limit.
func (e *EntryQueryBuilder) WithLimit(limit int) *EntryQueryBuilder {
	if limit > 0 {
		e.limit = limit
	}
	return e
}

// WithOffset set the offset.
func (e *EntryQueryBuilder) WithOffset(offset int) *EntryQueryBuilder {
	if offset > 0 {
		e.offset = offset
	}
	return e
}

// WithGloballyVisible set to show all entries globally.
func (e *EntryQueryBuilder) WithGloballyVisible() *EntryQueryBuilder {
	e.conditions = append(e.conditions, "c.hide_globally IS FALSE")
	e.conditions = append(e.conditions, "f.hide_globally IS FALSE")
	return e
}

// Args returns the configured args attribute.
func (e *EntryQueryBuilder) Args() []any {
	return e.args
}

// Args returns the configured fetchEnclosures attribute.
func (e *EntryQueryBuilder) FetchEnclosures() bool {
	return e.fetchEnclosures
}

// BuildCondition joins the configured conditions.
func (e *EntryQueryBuilder) BuildCondition() string {
	return strings.Join(e.conditions, " AND ")
}

// BuildSorting adds the sorting expression to query statement.
func (e *EntryQueryBuilder) BuildSorting() string {
	var parts string

	if len(e.sortExpressions) > 0 {
		parts += " ORDER BY " + strings.Join(e.sortExpressions, ", ")
	}

	if e.limit > 0 {
		parts += " LIMIT " + strconv.Itoa(e.limit)
	}

	if e.offset > 0 {
		parts += " OFFSET " + strconv.Itoa(e.offset)
	}

	return parts
}

// NewEntryQueryBuilder returns a new EntryQueryBuilder.
func NewEntryQueryBuilder(userID int64) *EntryQueryBuilder {
	return &EntryQueryBuilder{
		args:       []any{userID},
		conditions: []string{"e.user_id = $1"},
	}
}

// NewAnonymousQueryBuilder returns a new EntryQueryBuilder suitable for anonymous users.
func NewAnonymousQueryBuilder() *EntryQueryBuilder {
	return &EntryQueryBuilder{}
}
