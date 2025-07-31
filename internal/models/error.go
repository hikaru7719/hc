package models

type ErrorResponse struct {
	Messages []string `json:"messages"`
}

func NewErrorResponse(message string) *ErrorResponse {
	return &ErrorResponse{
		Messages: []string{message},
	}
}
func NewErrorResponseWithMessages(messages []string) *ErrorResponse {
	return &ErrorResponse{
		Messages: messages,
	}
}
