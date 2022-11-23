// The response package defines the success and error response type.
package response

// ResponsePagination definition.
type ResponsePagination struct {
	Next string `json:"next"`
}

// Response represent a success API response.
type Response[T any] struct {
	Data       T                  `json:"data"`
	Pagination ResponsePagination `json:"pagination"`
}
