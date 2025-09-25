// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package postgres // import "miniflux.app/v2/internal/storage/postgres"

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/querybuilder"
	"miniflux.app/v2/internal/urllib"
)

// FetchJobs retrieves a batch of jobs based on the conditions set in the builder.
// When limitPerHost is set, it limits the number of jobs per feed hostname to prevent overwhelming a single host.
func (p *Postgres) FetchJobs(b *querybuilder.BatchBuilder) (model.JobList, error) {
	query := `SELECT id, user_id, feed_url FROM feeds`

	if len(b.Conditions()) > 0 {
		query += " WHERE " + strings.Join(b.Conditions(), " AND ")
	}

	query += " ORDER BY next_check_at ASC"

	if b.BatchSize() > 0 {
		query += " LIMIT " + strconv.Itoa(b.BatchSize())
	}

	rows, err := p.db.Query(query, b.Args()...)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch batch of jobs: %v`, err)
	}
	defer rows.Close()

	jobs := make(model.JobList, 0, b.BatchSize())
	hosts := make(map[string]int)
	nbRows := 0
	nbSkippedFeeds := 0

	for rows.Next() {
		var job model.Job
		if err := rows.Scan(&job.FeedID, &job.UserID, &job.FeedURL); err != nil {
			return nil, fmt.Errorf(`store: unable to fetch job record: %v`, err)
		}

		nbRows++

		if b.LimitPerHost() > 0 {
			feedHostname := urllib.Domain(job.FeedURL)
			if hosts[feedHostname] >= b.LimitPerHost() {
				slog.Debug("Feed host limit reached for this batch",
					slog.String("feed_url", job.FeedURL),
					slog.String("feed_hostname", feedHostname),
					slog.Int("limit_per_host", b.LimitPerHost()),
					slog.Int("current", hosts[feedHostname]),
				)
				nbSkippedFeeds++
				continue
			}
			hosts[feedHostname]++
		}

		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(`store: error iterating on job records: %v`, err)
	}

	slog.Info("Created a batch of feeds",
		slog.Int("batch_size", b.BatchSize()),
		slog.Int("rows_count", nbRows),
		slog.Int("skipped_feeds_count", nbSkippedFeeds),
		slog.Int("jobs_count", len(jobs)),
	)

	return jobs, nil
}
