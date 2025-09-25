// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package querybuilder // import "miniflux.app/v2/internal/querybuilder"

import (
	"strconv"
	"strings"

	"miniflux.app/v2/internal/model"
)

// FeedQueryBuilder builds a SQL query to fetch feeds.
type FeedQueryBuilder struct {
	args              []any
	conditions        []string
	sortExpressions   []string
	limit             int
	offset            int
	withCounters      bool
	counterJoinFeeds  bool
	counterArgs       []any
	counterConditions []string
}

// NewFeedQueryBuilder returns a new FeedQueryBuilder.
func NewFeedQueryBuilder(userID int64) *FeedQueryBuilder {
	return &FeedQueryBuilder{
		args:              []any{userID},
		conditions:        []string{"f.user_id = $1"},
		counterArgs:       []any{userID, model.EntryStatusRead, model.EntryStatusUnread},
		counterConditions: []string{"e.user_id = $1", "e.status IN ($2, $3)"},
	}
}

// WithCategoryID filter by category ID.
func (f *FeedQueryBuilder) WithCategoryID(categoryID int64) *FeedQueryBuilder {
	if categoryID > 0 {
		f.conditions = append(f.conditions, "f.category_id = $"+strconv.Itoa(len(f.args)+1))
		f.args = append(f.args, categoryID)
		f.counterConditions = append(f.counterConditions, "f.category_id = $"+strconv.Itoa(len(f.counterArgs)+1))
		f.counterArgs = append(f.counterArgs, categoryID)
		f.counterJoinFeeds = true
	}
	return f
}

// WithFeedID filter by feed ID.
func (f *FeedQueryBuilder) WithFeedID(feedID int64) *FeedQueryBuilder {
	if feedID > 0 {
		f.conditions = append(f.conditions, "f.id = $"+strconv.Itoa(len(f.args)+1))
		f.args = append(f.args, feedID)
	}
	return f
}

// WithCounters let the builder return feeds with counters of statuses of entries.
func (f *FeedQueryBuilder) WithCounters() *FeedQueryBuilder {
	f.withCounters = true
	return f
}

// WithSorting add a sort expression.
func (f *FeedQueryBuilder) WithSorting(column, direction string) *FeedQueryBuilder {
	f.sortExpressions = append(f.sortExpressions, column+" "+direction)
	return f
}

// WithLimit set the limit.
func (f *FeedQueryBuilder) WithLimit(limit int) *FeedQueryBuilder {
	f.limit = limit
	return f
}

// WithOffset set the offset.
func (f *FeedQueryBuilder) WithOffset(offset int) *FeedQueryBuilder {
	f.offset = offset
	return f
}

// Args returns the configured args attribute.
func (f *FeedQueryBuilder) Args() []any {
	return f.args
}

// Args returns true if withCounters is configured.
func (f *FeedQueryBuilder) IsWithCounters() bool {
	return f.withCounters
}

// IsCounterJoinFeeds returns true if counterJoinFeeds is configured.
func (f *FeedQueryBuilder) IsCounterJoinFeeds() bool {
	return f.counterJoinFeeds
}

// CounterArgs returns the configured counterArgs attribute.
func (f *FeedQueryBuilder) CounterArgs() []any {
	return f.counterArgs
}

// BuildCondition joins configured conditions.
func (f *FeedQueryBuilder) BuildCondition() string {
	return strings.Join(f.conditions, " AND ")
}

// BuildCondition joins configured counter conditions.
func (f *FeedQueryBuilder) BuildCounterCondition() string {
	return strings.Join(f.counterConditions, " AND ")
}

// BuildSorting adds sorting expression to query statement.
func (f *FeedQueryBuilder) BuildSorting() string {
	var parts string

	if len(f.sortExpressions) > 0 {
		parts += " ORDER BY " + strings.Join(f.sortExpressions, ", ")
	}

	if len(parts) > 0 {
		parts += ", lower(f.title) ASC"
	}

	if f.limit > 0 {
		parts += " LIMIT " + strconv.Itoa(f.limit)
	}

	if f.offset > 0 {
		parts += " OFFSET " + strconv.Itoa(f.offset)
	}

	return parts
}
