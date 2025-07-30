package models

// ErrorResponse represents an error response with multiple messages
type ErrorResponse struct {
	Messages []string `json:"messages"`
}

// NewErrorResponse creates a new error response with a single message
func NewErrorResponse(message string) *ErrorResponse {
	return &ErrorResponse{
		Messages: []string{message},
	}
}

// NewErrorResponseWithMessages creates a new error response with multiple messages
func NewErrorResponseWithMessages(messages []string) *ErrorResponse {
	return &ErrorResponse{
		Messages: messages,
	}
}