package model

const (
	DefaultPage    = 1
	DefaultPerPage = 10
)

type Pagination struct {
	Page         int
	PerPage      int
	TotalItems   int
	TotalPages   int
	HasPrevious  bool
	HasNext      bool
	PreviousPage int
	NextPage     int
	Offset       int
}

func NewPagination(page, perPage, totalItems int) Pagination {
	if page < 1 {
		page = DefaultPage
	}

	if perPage < 1 {
		perPage = DefaultPerPage
	}

	if totalItems < 0 {
		totalItems = 0
	}

	totalPages := (totalItems + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	if page > totalPages {
		page = totalPages
	}

	pagination := Pagination{
		Page:         page,
		PerPage:      perPage,
		TotalItems:   totalItems,
		TotalPages:   totalPages,
		HasPrevious:  page > 1,
		HasNext:      page < totalPages,
		PreviousPage: 1,
		NextPage:     totalPages,
		Offset:       (page - 1) * perPage,
	}

	if pagination.HasPrevious {
		pagination.PreviousPage = page - 1
	}

	if pagination.HasNext {
		pagination.NextPage = page + 1
	}

	return pagination
}
