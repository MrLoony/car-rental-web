package handler

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/MrLoony/car-rental-web/internal/model"
)

func TestParseFavoriteSlugs(t *testing.T) {
	slugs, active := parseFavoriteSlugs(url.Values{
		"favorites": {"Nissan-Sunny-White, kia-pegas-gold,,nissan-sunny-white"},
	})

	if !active {
		t.Fatal("favorites active = false, want true")
	}
	if got := strings.Join(slugs, ","); got != "nissan-sunny-white,kia-pegas-gold" {
		t.Fatalf("favorite slugs = %q", got)
	}
}

func TestParseFavoriteSlugsEmptyParamIsActive(t *testing.T) {
	slugs, active := parseFavoriteSlugs(url.Values{"favorites": {""}})

	if !active {
		t.Fatal("favorites active = false, want true")
	}
	if len(slugs) != 0 {
		t.Fatalf("favorite slugs length = %d, want 0", len(slugs))
	}
}

func TestPaginationURLPreservesFavorites(t *testing.T) {
	request, err := http.NewRequest(http.MethodGet, "/cars?favorites=nissan-sunny-white,kia-pegas-gold&sort=price_asc&page=1", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}

	got := paginationURL(request, 2)

	if !strings.Contains(got, "favorites=nissan-sunny-white%2Ckia-pegas-gold") {
		t.Fatalf("paginationURL() = %q, want favorites query preserved", got)
	}
	if !strings.Contains(got, "page=2") {
		t.Fatalf("paginationURL() = %q, want page=2", got)
	}
}

func TestCatalogFilterURLPreservesFavorites(t *testing.T) {
	got := catalogFilterURL(model.CarFilter{
		Search:        "nissan",
		FavoritesOnly: true,
		FavoriteSlugs: []string{"nissan-sunny-white", "kia-pegas-gold"},
	}, "search")

	if got != "/cars?favorites=nissan-sunny-white%2Ckia-pegas-gold" {
		t.Fatalf("catalogFilterURL() = %q", got)
	}
}
