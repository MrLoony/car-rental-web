package model

import "testing"

func TestNewPaginationNormal(t *testing.T) {
	pagination := NewPagination(2, 10, 35)

	assertPagination(t, pagination, Pagination{
		Page:         2,
		PerPage:      10,
		TotalItems:   35,
		TotalPages:   4,
		HasPrevious:  true,
		HasNext:      true,
		PreviousPage: 1,
		NextPage:     3,
		Offset:       10,
	})
}

func TestNewPaginationFirstPage(t *testing.T) {
	pagination := NewPagination(1, 10, 35)

	if pagination.HasPrevious {
		t.Fatal("NewPagination().HasPrevious = true, want false")
	}

	if pagination.PreviousPage != 1 {
		t.Fatalf("NewPagination().PreviousPage = %d, want 1", pagination.PreviousPage)
	}
}

func TestNewPaginationLastPage(t *testing.T) {
	pagination := NewPagination(4, 10, 35)

	if pagination.HasNext {
		t.Fatal("NewPagination().HasNext = true, want false")
	}

	if pagination.NextPage != 4 {
		t.Fatalf("NewPagination().NextPage = %d, want 4", pagination.NextPage)
	}
}

func TestNewPaginationClampsPageBelowOne(t *testing.T) {
	pagination := NewPagination(0, 10, 35)

	if pagination.Page != 1 {
		t.Fatalf("NewPagination().Page = %d, want 1", pagination.Page)
	}
}

func TestNewPaginationClampsPageAboveTotalPages(t *testing.T) {
	pagination := NewPagination(99, 10, 35)

	if pagination.Page != 4 {
		t.Fatalf("NewPagination().Page = %d, want 4", pagination.Page)
	}

	if pagination.Offset != 30 {
		t.Fatalf("NewPagination().Offset = %d, want 30", pagination.Offset)
	}
}

func TestNewPaginationZeroTotalItems(t *testing.T) {
	pagination := NewPagination(1, 10, 0)

	if pagination.TotalPages != 1 {
		t.Fatalf("NewPagination().TotalPages = %d, want 1", pagination.TotalPages)
	}

	if pagination.TotalItems != 0 {
		t.Fatalf("NewPagination().TotalItems = %d, want 0", pagination.TotalItems)
	}
}

func TestNewPaginationInvalidPerPageDefaultsToTen(t *testing.T) {
	pagination := NewPagination(1, 0, 35)

	if pagination.PerPage != DefaultPerPage {
		t.Fatalf("NewPagination().PerPage = %d, want %d", pagination.PerPage, DefaultPerPage)
	}
}

func TestNewPaginationOffsetCalculation(t *testing.T) {
	pagination := NewPagination(3, 25, 100)

	if pagination.Offset != 50 {
		t.Fatalf("NewPagination().Offset = %d, want 50", pagination.Offset)
	}
}

func assertPagination(t *testing.T, got, want Pagination) {
	t.Helper()

	if got != want {
		t.Fatalf("NewPagination() = %+v, want %+v", got, want)
	}
}
