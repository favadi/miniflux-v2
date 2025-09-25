// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package postgres // import "miniflux.app/v2/internal/storage/postgres"

import (
	"database/sql"
	"fmt"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/querybuilder"
	"miniflux.app/v2/internal/timezone"
)

// GetFeeds returns a list of feeds that match the condition.
func (p *Postgres) GetFeeds(f *querybuilder.FeedQueryBuilder) (model.Feeds, error) {
	var query = `
		SELECT
			f.id,
			f.feed_url,
			f.site_url,
			f.title,
			f.description,
			f.etag_header,
			f.last_modified_header,
			f.user_id,
			f.checked_at at time zone u.timezone,
			f.next_check_at at time zone u.timezone,
			f.parsing_error_count,
			f.parsing_error_msg,
			f.scraper_rules,
			f.rewrite_rules,
			f.url_rewrite_rules,
			f.blocklist_rules,
			f.keeplist_rules,
			f.block_filter_entry_rules,
			f.keep_filter_entry_rules,
			f.crawler,
			f.user_agent,
			f.cookie,
			f.username,
			f.password,
			f.ignore_http_cache,
			f.allow_self_signed_certificates,
			f.fetch_via_proxy,
			f.disabled,
			f.no_media_player,
			f.hide_globally,
			f.category_id,
			c.title as category_title,
			c.hide_globally as category_hidden,
			fi.icon_id,
			i.external_id,
			u.timezone,
			f.apprise_service_urls,
			f.webhook_url,
			f.disable_http2,
			f.ntfy_enabled,
			f.ntfy_priority,
			f.ntfy_topic,
			f.pushover_enabled,
			f.pushover_priority,
			f.proxy_url
		FROM
			feeds f
		LEFT JOIN
			categories c ON c.id=f.category_id
		LEFT JOIN
			feed_icons fi ON fi.feed_id=f.id
		LEFT JOIN
			icons i ON i.id=fi.icon_id
		LEFT JOIN
			users u ON u.id=f.user_id
		WHERE %s
		%s
	`

	query = fmt.Sprintf(query, f.BuildCondition(), f.BuildSorting())

	rows, err := p.db.Query(query, f.Args()...)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch feeds: %w`, err)
	}
	defer rows.Close()

	readCounters, unreadCounters, err := p.fetchFeedCounter(f)
	if err != nil {
		return nil, err
	}

	feeds := make(model.Feeds, 0)
	for rows.Next() {
		var feed model.Feed
		var iconID sql.NullInt64
		var externalIconID sql.NullString
		var tz string
		feed.Category = &model.Category{}

		err := rows.Scan(
			&feed.ID,
			&feed.FeedURL,
			&feed.SiteURL,
			&feed.Title,
			&feed.Description,
			&feed.EtagHeader,
			&feed.LastModifiedHeader,
			&feed.UserID,
			&feed.CheckedAt,
			&feed.NextCheckAt,
			&feed.ParsingErrorCount,
			&feed.ParsingErrorMsg,
			&feed.ScraperRules,
			&feed.RewriteRules,
			&feed.UrlRewriteRules,
			&feed.BlocklistRules,
			&feed.KeeplistRules,
			&feed.BlockFilterEntryRules,
			&feed.KeepFilterEntryRules,
			&feed.Crawler,
			&feed.UserAgent,
			&feed.Cookie,
			&feed.Username,
			&feed.Password,
			&feed.IgnoreHTTPCache,
			&feed.AllowSelfSignedCertificates,
			&feed.FetchViaProxy,
			&feed.Disabled,
			&feed.NoMediaPlayer,
			&feed.HideGlobally,
			&feed.Category.ID,
			&feed.Category.Title,
			&feed.Category.HideGlobally,
			&iconID,
			&externalIconID,
			&tz,
			&feed.AppriseServiceURLs,
			&feed.WebhookURL,
			&feed.DisableHTTP2,
			&feed.NtfyEnabled,
			&feed.NtfyPriority,
			&feed.NtfyTopic,
			&feed.PushoverEnabled,
			&feed.PushoverPriority,
			&feed.ProxyURL,
		)

		if err != nil {
			return nil, fmt.Errorf(`store: unable to fetch feeds row: %w`, err)
		}

		if iconID.Valid && externalIconID.Valid {
			feed.Icon = &model.FeedIcon{FeedID: feed.ID, IconID: iconID.Int64, ExternalIconID: externalIconID.String}
		} else {
			feed.Icon = &model.FeedIcon{FeedID: feed.ID, IconID: 0, ExternalIconID: ""}
		}

		if readCounters != nil {
			if count, found := readCounters[feed.ID]; found {
				feed.ReadCount = count
			}
		}
		if unreadCounters != nil {
			if count, found := unreadCounters[feed.ID]; found {
				feed.UnreadCount = count
			}
		}

		feed.NumberOfVisibleEntries = feed.ReadCount + feed.UnreadCount
		feed.CheckedAt = timezone.Convert(tz, feed.CheckedAt)
		feed.NextCheckAt = timezone.Convert(tz, feed.NextCheckAt)
		feed.Category.UserID = feed.UserID
		feeds = append(feeds, &feed)
	}

	return feeds, nil
}

// GetFeed returns a single feed that match the condition.
func (p *Postgres) GetFeed(f *querybuilder.FeedQueryBuilder) (*model.Feed, error) {
	f.WithLimit(1)
	feeds, err := p.GetFeeds(f)
	if err != nil {
		return nil, err
	}

	if len(feeds) != 1 {
		return nil, nil
	}

	return feeds[0], nil
}

func (p *Postgres) fetchFeedCounter(f *querybuilder.FeedQueryBuilder) (unreadCounters map[int64]int, readCounters map[int64]int, err error) {
	if !f.IsWithCounters() {
		return nil, nil, nil
	}
	query := `
		SELECT
			e.feed_id,
			e.status,
			count(*)
		FROM
			entries e
		%s
		WHERE
			%s
		GROUP BY
			e.feed_id, e.status
	`
	join := ""
	if f.IsCounterJoinFeeds() {
		join = "LEFT JOIN feeds f ON f.id=e.feed_id"
	}
	query = fmt.Sprintf(query, join, f.BuildCounterCondition())

	rows, err := p.db.Query(query, f.CounterArgs()...)
	if err != nil {
		return nil, nil, fmt.Errorf(`store: unable to fetch feed counts: %w`, err)
	}
	defer rows.Close()

	readCounters = make(map[int64]int)
	unreadCounters = make(map[int64]int)
	for rows.Next() {
		var feedID int64
		var status string
		var count int
		if err := rows.Scan(&feedID, &status, &count); err != nil {
			return nil, nil, fmt.Errorf(`store: unable to fetch feed counter row: %w`, err)
		}

		switch status {
		case model.EntryStatusRead:
			readCounters[feedID] = count
		case model.EntryStatusUnread:
			unreadCounters[feedID] = count
		}
	}

	return readCounters, unreadCounters, nil
}
