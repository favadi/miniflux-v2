// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/querybuilder"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) showSearchEntryPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	entryID := request.RouteInt64Param(r, "entryID")
	searchQuery := request.QueryStringParam(r, "q", "")
	builder := querybuilder.NewEntryQueryBuilder(user.ID)
	builder.WithSearchQuery(searchQuery)
	builder.WithEntryID(entryID)
	builder.WithoutStatus(model.EntryStatusRemoved)

	entry, err := h.store.GetEntry(builder)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if entry == nil {
		html.NotFound(w, r)
		return
	}

	if entry.ShouldMarkAsReadOnView(user) {
		err = h.store.SetEntriesStatus(user.ID, []int64{entry.ID}, model.EntryStatusRead)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}

		entry.Status = model.EntryStatusRead
	}

	if user.AlwaysOpenExternalLinks {
		html.Redirect(w, r, entry.URL)
		return
	}

	entryPaginationBuilder := querybuilder.NewEntryPaginationBuilder(user.ID, entry.ID, user.EntryOrder, user.EntryDirection)
	entryPaginationBuilder.WithSearchQuery(searchQuery)
	prevEntry, nextEntry, err := h.store.Entries(entryPaginationBuilder)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	nextEntryRoute := ""
	if nextEntry != nil {
		nextEntryRoute = route.Path(h.router, "searchEntry", "entryID", nextEntry.ID)
	}

	prevEntryRoute := ""
	if prevEntry != nil {
		prevEntryRoute = route.Path(h.router, "searchEntry", "entryID", prevEntry.ID)
	}

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("searchQuery", searchQuery)
	view.Set("entry", entry)
	view.Set("prevEntry", prevEntry)
	view.Set("nextEntry", nextEntry)
	view.Set("nextEntryRoute", nextEntryRoute)
	view.Set("prevEntryRoute", prevEntryRoute)
	view.Set("menu", "search")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	view.Set("hasSaveEntry", h.store.HasSaveEntry(user.ID))

	html.OK(w, r, view.Render("entry"))
}
