package service

// ListParams holds pagination parameters for record listing.
type ListParams struct {
	Page    int
	PerPage int
}

// RecordListResult holds a paginated list of records.
type RecordListResult struct {
	Data       []map[string]any `json:"data"`
	Pagination PaginationMeta   `json:"pagination"`
}

// PaginationMeta holds pagination metadata.
type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// CreateResult holds the result of a record creation.
type CreateResult struct {
	ID string `json:"id"`
}
