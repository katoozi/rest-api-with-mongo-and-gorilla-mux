package model

type (
	// Response is the http json response schema
	Response struct {
		Status  int         `json:"status"`
		Message string      `json:"message"`
		Content interface{} `json:"content"`
	}

	// PaginatedResponse is the paginated response json schema
	// we not use it yet
	PaginatedResponse struct {
		Count    int         `json:"count"`
		Next     string      `json:"next"`
		Previous string      `json:"previous"`
		Results  interface{} `json:"results"`
	}
)

// NewResponse is the Response struct factory function.
func NewResponse(status int, message string, content interface{}) *Response {
	return &Response{
		Status:  status,
		Message: message,
		Content: content,
	}
}
