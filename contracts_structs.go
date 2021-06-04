package shuttle

import (
	"encoding/json"
	"io"
	"net/http"
)

// TextResult provides the ability render a result which contains text.
type TextResult struct {

	// StatusCode, if provided, use this value, otherwise HTTP 200.
	StatusCode int

	// ContentType, if provided, use this value.
	ContentType string

	// Content, if provided, use this value, otherwise no content will be written to the response stream.
	Content string
}

// BinaryResult provides the ability render a result which contains binary data.
type BinaryResult struct {

	// StatusCode, if provided, use this value, otherwise HTTP 200.
	StatusCode int

	// ContentType, if provided, use this value.
	ContentType string

	// Content, if provided, use this value, otherwise no content will be written to the response stream.
	Content []byte
}

type StreamResult struct {

	// StatusCode, if provided, use this value, otherwise HTTP 200.
	StatusCode int

	// ContentType, if provided, use this value.
	ContentType string

	// Content, if provided, use this value, otherwise no content will be written to the response stream.
	Content io.Reader
}

type SerializeResult struct {

	// StatusCode, if provided, use this value, otherwise HTTP 200.
	StatusCode int

	// ContentType, if provided, use this value.
	ContentType string

	// Content, if provided, use this value, otherwise no content will be written to the response stream.
	Content interface{}
}

// InputError represents some kind of problem with the calling HTTP request.
type InputError struct {

	// Fields indicates the exact location(s) of the errors including the part of the HTTP request itself this is
	// invalid. Valid field prefixes include "path", "query", "header", "form", and "body".
	Fields []string `json:"fields,omitempty"`

	// ID represents the unique, numeric contractual identifier that can be used to associate this error with a particular front-end error message, if any.
	ID int `json:"id,omitempty"`

	// Name represents the unique string-based, contractual value that can be used to associate this error with a particular front-end error message, if any.
	Name string `json:"name,omitempty"`

	// Message represents a friendly, user-facing message to indicate why there was a problem with the input.
	Message string `json:"message,omitempty"`
}

// InputErrors represents a set of problems with the calling HTTP request.
type InputErrors struct {
	Errors []InputError `json:"errors"`
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (this InputError) Error() string                   { return this.Message }
func (this *TextResult) SetContent(value interface{})   { this.Content = value.(string) }
func (this *BinaryResult) SetContent(value interface{}) { this.Content = value.([]byte) }
func (this *StreamResult) SetContent(value interface{}) { this.Content = value.(io.Reader) }
func (this *SerializeResult) SetContent(value interface{}) {
	switch typed := value.(type) {
	case []InputError:
		if content, ok := this.Content.(InputErrors); ok {
			content.Errors = typed
		} else {
			this.Content = InputErrors{Errors: typed}
		}

	case InputError:
		if content, ok := this.Content.(InputErrors); ok {
			content.Errors = content.Errors[0:0]
			content.Errors = append(content.Errors, typed)
		} else {
			this.Content = InputErrors{Errors: []InputError{typed}}
		}
	default:
		this.Content = value
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	notAcceptableResult = &TextResult{
		StatusCode:  http.StatusNotAcceptable,
		ContentType: mimeTypeApplicationJSONUTF8,
		Content: _serializeJSON(InputErrors{Errors: []InputError{{
			Fields:  []string{"header:Accept"},
			Name:    "invalid-accept-header",
			Message: "Unable to represent the application results using the Accept type.",
		}}}),
	}
	unsupportedMediaTypeResult = &SerializeResult{
		StatusCode: http.StatusUnsupportedMediaType,
		Content: InputErrors{Errors: []InputError{{
			Fields:  []string{"header:Content-Type"},
			Name:    "invalid-content-type-header",
			Message: "The content type specified, if any, was not recognized.",
		}}},
	}
	deserializationResult = &SerializeResult{
		StatusCode: http.StatusBadRequest,
		Content: InputErrors{Errors: []InputError{{
			Fields:  []string{"body"},
			Name:    "malformed-request-payload",
			Message: "The body did not contain well-formed data and could not be properly deserialized.",
		}}},
	}
	bindFailedResult = &SerializeResult{
		StatusCode: http.StatusBadRequest,
		Content: InputErrors{Errors: []InputError{{
			Fields:  []string{"body"},
			Name:    "malformed-request-payload",
			Message: "Unable to bind the HTTP request values onto the appropriate data structure.",
		}}},
	}
)

func _serializeJSON(instance interface{}) string {
	raw, _ := json.Marshal(instance)
	return string(raw)
}
