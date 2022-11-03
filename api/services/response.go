package services

type ResponsePagination struct {
	Next string `json:"next"`
}

type Response[T any] struct {
	Data       T                  `json:"data"`
	Error      error              `json:"error"`
	Pagination ResponsePagination `json:"pagination"`
}
