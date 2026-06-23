package repository

import (
	"strings"
	"testing"
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
