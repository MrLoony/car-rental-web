package repository

import (
	"fmt"
	"strings"
	"testing"

	"github.com/MrLoony/car-rental-web/internal/model"
)

func TestGetCarImagesByCarIDQueryOrdersPrimaryThenSortOrder(t *testing.T) {
	want := "ORDER BY is_primary DESC, sort_order ASC, id ASC"
	if !strings.Contains(getCarImagesByCarIDQuery, want) {
		t.Fatalf("gallery query ordering = %q, want to contain %q", getCarImagesByCarIDQuery, want)
	}
}

func TestGetCatalogImageURLsByCarIDsQueryUsesPrimaryThenFirstGalleryImage(t *testing.T) {
	expectedParts := []string{
		"SELECT DISTINCT ON (car_id)",
		"WHERE car_id = ANY($1)",
		"ORDER BY car_id, is_primary DESC, sort_order ASC, id ASC",
	}

	for _, want := range expectedParts {
		if !strings.Contains(getCatalogImageURLsByCarIDsQuery, want) {
			t.Fatalf("catalog image query = %q, want to contain %q", getCatalogImageURLsByCarIDsQuery, want)
		}
	}
}

func TestAppendPublicCarFiltersUsesFavoriteSlugs(t *testing.T) {
	var query strings.Builder
	args := make([]any, 0)

	appendPublicCarFilters(&query, &args, model.CarFilter{
		FavoritesOnly: true,
		FavoriteSlugs: []string{"nissan-sunny-white", "kia-pegas-gold"},
	})

	if !strings.Contains(query.String(), "c.slug = ANY($1)") {
		t.Fatalf("query = %q, want favorite slug filter", query.String())
	}
	if len(args) != 1 {
		t.Fatalf("args length = %d, want 1", len(args))
	}
	slugs, ok := args[0].([]string)
	if !ok {
		t.Fatalf("arg type = %T, want []string", args[0])
	}
	if fmt.Sprint(slugs) != "[nissan-sunny-white kia-pegas-gold]" {
		t.Fatalf("favorite slugs = %v", slugs)
	}
}

func TestAppendPublicCarFiltersUsesFalseForEmptyFavoritesMode(t *testing.T) {
	var query strings.Builder
	args := make([]any, 0)

	appendPublicCarFilters(&query, &args, model.CarFilter{FavoritesOnly: true})

	if !strings.Contains(query.String(), "AND FALSE") {
		t.Fatalf("query = %q, want empty favorites filter to return no rows", query.String())
	}
	if len(args) != 0 {
		t.Fatalf("args length = %d, want 0", len(args))
	}
}
