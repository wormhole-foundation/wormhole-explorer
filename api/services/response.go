package services

// ResponsePagination definition.
type ResponsePagination struct {
	Next string `json:"next"`
}

// Response represent a success api response.
type Response[T any] struct {
	Data       T                  `json:"data"`
	Pagination ResponsePagination `json:"pagination"`
}
