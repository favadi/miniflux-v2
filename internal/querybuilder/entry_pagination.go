// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package querybuilder // import "miniflux.app/v2/internal/querybuilder"

import (
	"fmt"
	"strconv"
)

// EntryPaginationBuilder is a builder for entry prev/next queries.
type EntryPaginationBuilder struct {
	conditions []string
	args       []any
	entryID    int64
	order      string
	direction  string
}

// WithArg appends the given argument to the builder.
func (e *EntryPaginationBuilder) WithArg(arg any) {
	e.args = append(e.args, arg)
}

// WithSearchQuery adds full-text search query to the condition.
func (e *EntryPaginationBuilder) WithSearchQuery(query string) {
	if query != "" {
		e.conditions = append(e.conditions, fmt.Sprintf("e.document_vectors @@ plainto_tsquery($%d)", len(e.args)+1))
		e.args = append(e.args, query)
	}
}

// WithStarred adds starred to the condition.
func (e *EntryPaginationBuilder) WithStarred() {
	e.conditions = append(e.conditions, "e.starred is true")
}

// WithFeedID adds feed_id to the condition.
func (e *EntryPaginationBuilder) WithFeedID(feedID int64) {
	if feedID != 0 {
		e.conditions = append(e.conditions, "e.feed_id = $"+strconv.Itoa(len(e.args)+1))
		e.args = append(e.args, feedID)
	}
}

// WithCategoryID adds category_id to the condition.
func (e *EntryPaginationBuilder) WithCategoryID(categoryID int64) {
	if categoryID != 0 {
		e.conditions = append(e.conditions, "f.category_id = $"+strconv.Itoa(len(e.args)+1))
		e.args = append(e.args, categoryID)
	}
}

// WithStatus adds status to the condition.
func (e *EntryPaginationBuilder) WithStatus(status string) {
	if status != "" {
		e.conditions = append(e.conditions, "e.status = $"+strconv.Itoa(len(e.args)+1))
		e.args = append(e.args, status)
	}
}

// WithTags adds given tags to the condition.
func (e *EntryPaginationBuilder) WithTags(tags []string) {
	if len(tags) > 0 {
		for _, tag := range tags {
			e.conditions = append(e.conditions, fmt.Sprintf("LOWER($%d) = ANY(LOWER(e.tags::text)::text[])", len(e.args)+1))
			e.args = append(e.args, tag)
		}
	}
}

// WithGloballyVisible adds global visibility to the condition.
func (e *EntryPaginationBuilder) WithGloballyVisible() {
	e.conditions = append(e.conditions, "not c.hide_globally")
	e.conditions = append(e.conditions, "not f.hide_globally")
}

// Conditions returns the configured conditions attribute.
func (e *EntryPaginationBuilder) Conditions() []string {
	return e.conditions
}

// Conditions returns the configured args attribute.
func (e *EntryPaginationBuilder) Args() []any {
	return e.args
}

// Conditions returns the configured entryID attribute.
func (e *EntryPaginationBuilder) EntryID() int64 {
	return e.entryID
}

// Conditions returns the configured order attribute.
func (e *EntryPaginationBuilder) Order() string {
	return e.order
}

// Conditions returns the configured direction attribute.
func (e *EntryPaginationBuilder) Direction() string {
	return e.direction
}

// NewEntryPaginationBuilder returns a new EntryPaginationBuilder.
func NewEntryPaginationBuilder(userID, entryID int64, order, direction string) *EntryPaginationBuilder {
	return &EntryPaginationBuilder{
		args:       []any{userID, "removed"},
		conditions: []string{"e.user_id = $1", "e.status <> $2"},
		entryID:    entryID,
		order:      order,
		direction:  direction,
	}
}
