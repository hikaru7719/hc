package models

import (
	"reflect"
	"testing"
)

func TestNewErrorResponse(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    *ErrorResponse
	}{
		{
			name:    "Simple error message",
			message: "Something went wrong",
			want: &ErrorResponse{
				Messages: []string{"Something went wrong"},
			},
		},
		{
			name:    "Empty message",
			message: "",
			want: &ErrorResponse{
				Messages: []string{""},
			},
		},
		{
			name:    "Message with special characters",
			message: "Error: Invalid input <script>alert('xss')</script>",
			want: &ErrorResponse{
				Messages: []string{"Error: Invalid input <script>alert('xss')</script>"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewErrorResponse(tt.message)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewErrorResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewErrorResponseWithMessages(t *testing.T) {
	tests := []struct {
		name     string
		messages []string
		want     *ErrorResponse
	}{
		{
			name:     "Single message",
			messages: []string{"Error 1"},
			want: &ErrorResponse{
				Messages: []string{"Error 1"},
			},
		},
		{
			name:     "Multiple messages",
			messages: []string{"Error 1", "Error 2", "Error 3"},
			want: &ErrorResponse{
				Messages: []string{"Error 1", "Error 2", "Error 3"},
			},
		},
		{
			name:     "Empty messages slice",
			messages: []string{},
			want: &ErrorResponse{
				Messages: []string{},
			},
		},
		{
			name:     "Nil messages",
			messages: nil,
			want: &ErrorResponse{
				Messages: nil,
			},
		},
		{
			name:     "Messages with empty strings",
			messages: []string{"", "Error", ""},
			want: &ErrorResponse{
				Messages: []string{"", "Error", ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewErrorResponseWithMessages(tt.messages)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewErrorResponseWithMessages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorResponseFields(t *testing.T) {
	// Test that the ErrorResponse struct has the correct fields
	err := &ErrorResponse{
		Messages: []string{"test"},
	}

	// Use reflection to check the JSON tags
	typ := reflect.TypeOf(*err)
	field, ok := typ.FieldByName("Messages")
	if !ok {
		t.Fatal("ErrorResponse should have Messages field")
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag != "messages" {
		t.Errorf("Messages field should have json tag 'messages', got '%s'", jsonTag)
	}
}
