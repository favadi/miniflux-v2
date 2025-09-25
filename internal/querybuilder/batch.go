// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package querybuilder // import "miniflux.app/v2/internal/querybuilder"

import (
	"strconv"
)

// BatchBuilder builds a SQL query to create batch jobs.
type BatchBuilder struct {
	args         []any
	conditions   []string
	batchSize    int
	limitPerHost int
}

// NewBatchBuilder returns a new BatchBuilder.
func NewBatchBuilder() *BatchBuilder {
	return &BatchBuilder{}
}

// WithBatchSize sets size of the batch job.
func (b *BatchBuilder) WithBatchSize(batchSize int) *BatchBuilder {
	b.batchSize = batchSize
	return b
}

// WithUserID sets the user ID of the batch job.
func (b *BatchBuilder) WithUserID(userID int64) *BatchBuilder {
	b.conditions = append(b.conditions, "user_id = $"+strconv.Itoa(len(b.args)+1))
	b.args = append(b.args, userID)
	return b
}

// WithCategoryID sets the category of the batch job.
func (b *BatchBuilder) WithCategoryID(categoryID int64) *BatchBuilder {
	b.conditions = append(b.conditions, "category_id = $"+strconv.Itoa(len(b.args)+1))
	b.args = append(b.args, categoryID)
	return b
}

// WithErrorLimit sets the error limit of the batch job.
func (b *BatchBuilder) WithErrorLimit(limit int) *BatchBuilder {
	if limit > 0 {
		b.conditions = append(b.conditions, "parsing_error_count < $"+strconv.Itoa(len(b.args)+1))
		b.args = append(b.args, limit)
	}
	return b
}

// WithNextCheckExpired sets next check expiry of the batch job.
func (b *BatchBuilder) WithNextCheckExpired() *BatchBuilder {
	b.conditions = append(b.conditions, "next_check_at < now()")
	return b
}

// WithoutDisabledFeeds sets the query to not included disabled feeds.
func (b *BatchBuilder) WithoutDisabledFeeds() *BatchBuilder {
	b.conditions = append(b.conditions, "disabled IS false")
	return b
}

// WithLimitPerHost sets the limit per host of the batch job.
func (b *BatchBuilder) WithLimitPerHost(limit int) *BatchBuilder {
	if limit > 0 {
		b.limitPerHost = limit
	}
	return b
}

// Args returns the configured args attribute.
func (b *BatchBuilder) Args() []any {
	return b.args
}

// Conditions returns the configured conditions attribute.
func (b *BatchBuilder) Conditions() []string {
	return b.conditions
}

// BatchSize returns the configured batch size.
func (b *BatchBuilder) BatchSize() int {
	return b.batchSize
}

// LimitPerHost returns the configured limit per host.
func (b *BatchBuilder) LimitPerHost() int {
	return b.limitPerHost
}
