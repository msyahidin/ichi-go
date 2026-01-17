package dto

// PaginationRequest represents common pagination parameters
type PaginationRequest struct {
	Page     int    `json:"page" query:"page" validate:"min=1"`
	PageSize int    `json:"page_size" query:"page_size" validate:"min=1,max=100"`
	SortBy   string `json:"sort_by,omitempty" query:"sort_by"`
	SortDir  string `json:"sort_dir,omitempty" query:"sort_dir" validate:"omitempty,oneof=asc desc"`
}

// PaginationMetadata represents pagination metadata in responses
type PaginationMetadata struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalPages int `json:"total_pages"`
	TotalCount int `json:"total_count"`
}

// NewPaginationMetadata creates pagination metadata
func NewPaginationMetadata(page, pageSize, totalCount int) PaginationMetadata {
	totalPages := totalCount / pageSize
	if totalCount%pageSize != 0 {
		totalPages++
	}

	return PaginationMetadata{
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		TotalCount: totalCount,
	}
}

// GetOffset calculates the offset for database queries
func (p *PaginationRequest) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetLimit returns the page size limit
func (p *PaginationRequest) GetLimit() int {
	return p.PageSize
}

// Validate sets default values if not provided
func (p *PaginationRequest) Validate() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}
	if p.SortDir != "asc" && p.SortDir != "desc" {
		p.SortDir = "desc"
	}
}
