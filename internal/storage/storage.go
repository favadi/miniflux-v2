// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"database/sql"
	"errors"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/querybuilder"
)

var ErrAPIKeyNotFound = errors.New("store: API Key not found")

type Storage interface {
	APIKeyExists(userID int64, description string) bool
	APIKeys(userID int64) (model.APIKeys, error)
	AddWebAuthnCredential(userID int64, handle []byte, credential *webauthn.Credential) error
	AnotherCategoryExists(userID int64, categoryID int64, title string) bool
	AnotherFeedURLExists(userID int64, feedID int64, feedURL string) bool
	AnotherUserExists(userID int64, username string) bool
	AnotherUserWithFieldExists(userID int64, field string, value string) bool
	AppSession(id string) (*model.Session, error)
	ArchiveEntries(status string, interval time.Duration, limit int) (int64, error)
	Categories(userID int64) (model.Categories, error)
	CategoriesWithFeedCount(userID int64) (model.Categories, error)
	Category(userID int64, categoryID int64) (*model.Category, error)
	CategoryByTitle(userID int64, title string) (*model.Category, error)
	CategoryFeedExists(userID int64, categoryID int64, feedID int64) bool
	CategoryIDExists(userID int64, categoryID int64) bool
	CategoryTitleExists(userID int64, title string) bool
	CheckPassword(username string, password string) error
	CheckedAt(userID int64, feedID int64) (time.Time, error)
	CleanOldSessions(interval time.Duration) int64
	CleanOldUserSessions(interval time.Duration) int64
	ClearRemovedEntriesContent(limit int) (int64, error)
	CountAllEntries() map[string]int64
	CountAllFeeds() map[string]int64
	CountAllFeedsWithErrors() int
	CountEntries(e *querybuilder.EntryQueryBuilder) (count int, err error)
	CountUnreadEntries(userID int64) int
	CountUserFeedsWithErrors(userID int64) int
	CountUsers() int
	CountWebAuthnCredentialsByUserID(userID int64) int
	CreateAPIKey(userID int64, description string) (*model.APIKey, error)
	CreateAppSession() (*model.Session, error)
	CreateAppSessionWithUserPrefs(userID int64) (*model.Session, error)
	CreateCategory(userID int64, request *model.CategoryCreationRequest) (*model.Category, error)
	CreateFeed(feed *model.Feed) error
	CreateUser(userCreationRequest *model.UserCreationRequest) (*model.User, error)
	CreateUserSessionFromUsername(username string, userAgent string, ip string) (sessionID string, userID int64, err error)
	DBSize() (string, error)
	DBStats() sql.DBStats
	DatabaseVersion() string
	DeleteAPIKey(userID int64, keyID int64) error
	DeleteAllWebAuthnCredentialsByUserID(userID int64) error
	DeleteCredentialByHandle(userID int64, handle []byte) error
	DeleteRemovedEntriesEnclosures() (int64, error)
	Entries(e *querybuilder.EntryPaginationBuilder) (*model.Entry, *model.Entry, error)
	EntryShareCode(userID int64, entryID int64) (shareCode string, err error)
	FeedByID(userID int64, feedID int64) (*model.Feed, error)
	FeedExists(userID int64, feedID int64) bool
	FeedURLExists(userID int64, feedURL string) bool
	Feeds(userID int64) (model.Feeds, error)
	FeedsByCategoryWithCounters(userID int64, categoryID int64) (model.Feeds, error)
	FeedsWithCounters(userID int64) (model.Feeds, error)
	FetchCounters(userID int64) (model.FeedCounters, error)
	FetchJobs(b *querybuilder.BatchBuilder) (model.JobList, error)
	FirstCategory(userID int64) (*model.Category, error)
	FlushAllSessions() (err error)
	FlushHistory(userID int64) error
	GetEnclosure(enclosureID int64) (*model.Enclosure, error)
	GetEnclosures(entryID int64) (model.EnclosureList, error)
	GetEnclosuresForEntries(entryIDs []int64) (map[int64]model.EnclosureList, error)
	GetEntries(e *querybuilder.EntryQueryBuilder) (model.Entries, error)
	GetEntry(e *querybuilder.EntryQueryBuilder) (*model.Entry, error)
	GetEntryIDs(e *querybuilder.EntryQueryBuilder) ([]int64, error)
	GetFeed(f *querybuilder.FeedQueryBuilder) (*model.Feed, error)
	GetFeeds(f *querybuilder.FeedQueryBuilder) (model.Feeds, error)
	GetReadTime(feedID int64, entryHash string) int
	GoogleReaderUserCheckPassword(username string, password string) error
	GoogleReaderUserGetIntegration(username string) (*model.Integration, error)
	HasDuplicateFeverUsername(userID int64, feverUsername string) bool
	HasDuplicateGoogleReaderUsername(userID int64, googleReaderUsername string) bool
	HasFeedIcon(feedID int64) bool
	HasPassword(userID int64) (bool, error)
	HasSaveEntry(userID int64) (result bool)
	IconByExternalID(externalIconID string) (*model.Icon, error)
	IconByFeedID(userID int64, feedID int64) (*model.Icon, error)
	IconByID(iconID int64) (*model.Icon, error)
	Icons(userID int64) (model.Icons, error)
	Integration(userID int64) (*model.Integration, error)
	IsNewEntry(feedID int64, entryHash string) bool
	MarkAllAsRead(userID int64) error
	MarkAllAsReadBeforeDate(userID int64, before time.Time) error
	MarkCategoryAsRead(userID int64, categoryID int64, before time.Time) error
	MarkFeedAsRead(userID int64, feedID int64, before time.Time) error
	MarkGloballyVisibleFeedsAsRead(userID int64) error
	Ping() error
	RefreshFeedEntries(userID int64, feedID int64, entries model.Entries, updateExistingEntries bool) (newEntries model.Entries, err error)
	RemoveAndReplaceCategoriesByName(userid int64, titles []string) error
	RemoveCategory(userID int64, categoryID int64) error
	RemoveFeed(userID int64, feedID int64) error
	RemoveUser(userID int64) error
	RemoveUserAsync(userID int64)
	RemoveUserSessionByID(userID int64, sessionID int64) error
	RemoveUserSessionByToken(userID int64, token string) error
	ResetFeedErrors() error
	ResetNextCheckAt() error
	SetAPIKeyUsedTimestamp(userID int64, token string) error
	SetEntriesStarredState(userID int64, entryIDs []int64, starred bool) error
	SetEntriesStatus(userID int64, entryIDs []int64, status string) error
	SetEntriesStatusCount(userID int64, entryIDs []int64, status string) (int, error)
	SetLastLogin(userID int64) error
	StoreFeedIcon(feedID int64, icon *model.Icon) error
	ToggleStarred(userID int64, entryID int64) error
	UnshareEntry(userID int64, entryID int64) (err error)
	UpdateAppSessionField(sessionID string, field string, value any) error
	UpdateAppSessionObjectField(sessionID string, field string, value any) error
	UpdateCategory(category *model.Category) error
	UpdateEnclosure(enclosure *model.Enclosure) error
	UpdateEntryTitleAndContent(entry *model.Entry) error
	UpdateFeed(feed *model.Feed) (err error)
	UpdateFeedError(feed *model.Feed) (err error)
	UpdateIntegration(integration *model.Integration) error
	UpdateUser(user *model.User) error
	UserByAPIKey(token string) (*model.User, error)
	UserByFeverToken(token string) (*model.User, error)
	UserByField(field string, value string) (*model.User, error)
	UserByID(userID int64) (*model.User, error)
	UserByUsername(username string) (*model.User, error)
	UserExists(username string) bool
	UserLanguage(userID int64) (language string)
	UserSessionByToken(token string) (*model.UserSession, error)
	UserSessions(userID int64) ([]model.UserSession, error)
	Users() (model.Users, error)
	WebAuthnCredentialByHandle(handle []byte) (int64, *model.WebAuthnCredential, error)
	WebAuthnCredentialsByUserID(userID int64) ([]model.WebAuthnCredential, error)
	WebAuthnSaveLogin(handle []byte) error
	WebAuthnUpdateName(handle []byte, name string) error
	WeeklyFeedEntryCount(userID int64, feedID int64) (int, error)
}

type DBMigrator interface {
	Migrate() error
	IsSchemaUpToDate() error
}
