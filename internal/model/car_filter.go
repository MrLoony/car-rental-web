package model

const (
	SortNewest    = "newest"
	SortPriceAsc  = "price_asc"
	SortPriceDesc = "price_desc"
)

type CarFilter struct {
	Search        string
	CategorySlug  string
	FuelType      string
	Transmission  string
	FavoritesOnly bool
	FavoriteSlugs []string
	Sort          string
	Page          int
	PerPage       int
}

func NormalizeCarSort(sort string) string {
	switch sort {
	case SortNewest, SortPriceAsc, SortPriceDesc:
		return sort
	default:
		return SortNewest
	}
}
