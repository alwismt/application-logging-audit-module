package common

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

type Pagination struct {
	Page  int
	Limit int
}

func (p Pagination) Offset() int {
	page := p.Page
	if page < 1 {
		page = DefaultPage
	}
	limit := p.Limit
	if limit < 1 {
		limit = DefaultLimit
	}
	return (page - 1) * limit
}

func NormalizePagination(page, limit int) Pagination {
	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	return Pagination{Page: page, Limit: limit}
}
