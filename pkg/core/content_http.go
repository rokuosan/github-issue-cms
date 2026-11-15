package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// NewContentHandler serves a previously loaded snapshot without contacting GitHub.
func NewContentHandler(snapshot *ContentSnapshot) http.Handler {
	return &contentHandler{snapshot: snapshot}
}

type contentHandler struct {
	snapshot *ContentSnapshot
}

func (h *contentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		writeContentError(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	switch {
	case r.URL.Path == "/data/articles":
		query, err := parseContentQuery(r.URL.RawQuery)
		if err != nil {
			writeContentError(w, r, http.StatusBadRequest, err.Error())
			return
		}
		page, err := h.snapshot.Query(query)
		if err != nil {
			writeContentError(w, r, http.StatusBadRequest, err.Error())
			return
		}
		writeContentJSON(w, r, http.StatusOK, page)
	case r.URL.Path == "/data/metadata.json":
		writeContentJSON(w, r, http.StatusOK, h.snapshot.Metadata())
	case r.URL.Path == "/data/search/index.json":
		writeContentJSON(w, r, http.StatusOK, h.snapshot.SearchIndex())
	case strings.HasPrefix(r.URL.Path, "/data/articles/page/"):
		h.servePage(w, r)
	case strings.HasPrefix(r.URL.Path, "/data/article/"):
		h.serveArticle(w, r)
	case strings.HasPrefix(r.URL.Path, "/images/"):
		asset, ok := h.snapshot.Assets[r.URL.Path]
		if !ok {
			writeContentError(w, r, http.StatusNotFound, "image not found")
			return
		}
		w.Header().Set("Content-Type", asset.ContentType)
		w.Header().Set("Content-Length", strconv.Itoa(len(asset.Data)))
		if r.Method != http.MethodHead {
			_, _ = w.Write(asset.Data)
		}
	default:
		writeContentError(w, r, http.StatusNotFound, "not found")
	}
}

func (h *contentHandler) serveArticle(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimPrefix(r.URL.Path, "/data/article/")
	raw = strings.TrimSuffix(raw, ".json")
	id, err := positiveContentInteger(raw)
	if err != nil {
		writeContentError(w, r, http.StatusBadRequest, "invalid article ID")
		return
	}
	article, ok := h.snapshot.Find(id)
	if !ok {
		writeContentError(w, r, http.StatusNotFound, "article not found")
		return
	}
	writeContentJSON(w, r, http.StatusOK, article)
}

func (h *contentHandler) servePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.RawQuery != "" {
		writeContentError(w, r, http.StatusBadRequest, "static pages do not accept query parameters; use /data/articles")
		return
	}
	raw := strings.TrimPrefix(r.URL.Path, "/data/articles/page/")
	if !strings.HasSuffix(raw, ".json") {
		writeContentError(w, r, http.StatusNotFound, "page not found")
		return
	}
	pageNumber, err := positiveContentInteger(strings.TrimSuffix(raw, ".json"))
	if err != nil {
		writeContentError(w, r, http.StatusBadRequest, "invalid page number")
		return
	}
	page, err := h.snapshot.Query(ContentQuery{Page: pageNumber})
	if err != nil {
		writeContentError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if pageNumber > page.Pagination.TotalPages {
		writeContentError(w, r, http.StatusNotFound, "page not found")
		return
	}
	writeContentJSON(w, r, http.StatusOK, page)
}

func parseContentQuery(raw string) (ContentQuery, error) {
	values, err := url.ParseQuery(raw)
	if err != nil {
		return ContentQuery{}, fmt.Errorf("invalid query encoding")
	}
	for key, entries := range values {
		switch key {
		case "tag":
		case "page", "per_page", "q", "category", "sort", "order":
			if len(entries) != 1 {
				return ContentQuery{}, fmt.Errorf("%s must be specified once", key)
			}
		default:
			return ContentQuery{}, fmt.Errorf("unknown query parameter: %s", key)
		}
	}
	query := ContentQuery{
		Page: 1, Q: values.Get("q"), Tags: values["tag"],
		Category: values.Get("category"), Sort: values.Get("sort"), Order: values.Get("order"),
	}
	for key, dest := range map[string]*int{"page": &query.Page, "per_page": &query.PerPage} {
		if entries, ok := values[key]; ok {
			value, err := positiveContentInteger(entries[0])
			if err != nil {
				return ContentQuery{}, fmt.Errorf("%s must be a positive integer", key)
			}
			*dest = value
		}
	}
	for _, tag := range query.Tags {
		if tag == "" {
			return ContentQuery{}, fmt.Errorf("tag must not be empty")
		}
	}
	return query, nil
}

func positiveContentInteger(raw string) (int, error) {
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || strconv.Itoa(value) != raw {
		return 0, fmt.Errorf("invalid positive integer")
	}
	return value, nil
}

func writeContentError(w http.ResponseWriter, r *http.Request, status int, message string) {
	writeContentJSON(w, r, status, map[string]string{"error": message})
}

func writeContentJSON(w http.ResponseWriter, r *http.Request, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		status = http.StatusInternalServerError
		data = []byte(`{"error":"failed to encode response"}`)
	}
	data = append(data, '\n')
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(status)
	if r.Method != http.MethodHead {
		_, _ = w.Write(data)
	}
}
