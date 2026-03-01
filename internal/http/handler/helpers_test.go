package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"elibrary/internal/repository"
)

func TestHealth(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "OK" {
		t.Fatalf("body = %q, want %q", rec.Body.String(), "OK")
	}
}

func TestParseBookFilter(t *testing.T) {
	t.Parallel()

	values := url.Values{
		"id":              []string{"550e8400-e29b-41d4-a716-446655440000"},
		"barcode":         []string{"1234567890123"},
		"factory_barcode": []string{"factory"},
		"q":               []string{" history "},
		"publisher_id":    []string{"550e8400-e29b-41d4-a716-446655440001"},
		"year_from":       []string{"1990"},
		"year_to":         []string{"2000"},
		"limit":           []string{"50"},
		"offset":          []string{"10"},
	}
	req := httptest.NewRequest(http.MethodGet, "/books?"+values.Encode(), nil)

	got, err := parseBookFilter(req)
	if err != nil {
		t.Fatalf("parseBookFilter() error = %v", err)
	}

	if got.ID == nil || got.Barcode == nil || got.FactoryBarcode == nil || got.Query == nil || got.PublisherID == nil {
		t.Fatalf("parseBookFilter() missing expected pointers: %+v", got)
	}
	if *got.Query != "history" {
		t.Fatalf("Query = %q, want %q", *got.Query, "history")
	}
	if *got.YearFrom != 1990 || *got.YearTo != 2000 {
		t.Fatalf("year range = %v..%v, want 1990..2000", *got.YearFrom, *got.YearTo)
	}
	if *got.Limit != 50 || *got.Offset != 10 {
		t.Fatalf("pagination = limit %d offset %d, want 50 and 10", *got.Limit, *got.Offset)
	}
}

func TestParseBookFilterInvalidValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		query string
		want  string
	}{
		{name: "bad id", query: "id=bad", want: "invalid id"},
		{name: "bad publisher", query: "publisher_id=bad", want: "invalid publisher_id"},
		{name: "bad year_from", query: "year_from=nope", want: "invalid year_from"},
		{name: "bad year_to", query: "year_to=nope", want: "invalid year_to"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/books?"+tt.query, nil)
			_, err := parseBookFilter(req)
			if err == nil || err.Error() != tt.want {
				t.Fatalf("parseBookFilter() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestParseBookFilterDefaults(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	got, err := parseBookFilter(req)
	if err != nil {
		t.Fatalf("parseBookFilter() error = %v", err)
	}

	if got.Limit == nil || *got.Limit != 20 {
		t.Fatalf("Limit = %v, want 20", got.Limit)
	}
	if got.Offset == nil || *got.Offset != 0 {
		t.Fatalf("Offset = %v, want 0", got.Offset)
	}
}

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	payload := map[string]any{"count": 2, "items": []string{"a", "b"}}

	writeJSON(rec, http.StatusCreated, payload)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", got, "application/json")
	}

	var decoded map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if !reflect.DeepEqual(decoded["items"], []any{"a", "b"}) {
		t.Fatalf("items = %#v, want %#v", decoded["items"], []any{"a", "b"})
	}
}

func TestParseIntDefaultAndIntPtr(t *testing.T) {
	t.Parallel()

	if got := parseIntDefault(" 15 ", 20); got != 15 {
		t.Fatalf("parseIntDefault() = %d, want %d", got, 15)
	}
	if got := parseIntDefault("bad", 20); got != 20 {
		t.Fatalf("parseIntDefault() invalid = %d, want %d", got, 20)
	}

	p := intPtr(7)
	if p == nil || *p != 7 {
		t.Fatalf("intPtr() = %v, want pointer to 7", p)
	}

	var _ repository.BookFilter
}
