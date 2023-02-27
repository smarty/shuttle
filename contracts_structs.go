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

	// Headers, if provided, are added to the response
	Headers map[string][]string

	// Content, if provided, use this value, otherwise no content will be written to the response stream.
	Content string
}

// BinaryResult provides the ability render a result which contains binary data.
type BinaryResult struct {

	// StatusCode, if provided, use this value, otherwise HTTP 200.
	StatusCode int

	// ContentType, if provided, use this value.
	ContentType string

	// ContentDisposition, if provided, use this value.
	ContentDisposition string

	// Headers, if provided, are added to the response
	Headers map[string][]string

	// Content, if provided, use this value, otherwise no content will be written to the response stream.
	Content []byte
}

// StreamResult provides the ability render a result which is streamed from another source.
type StreamResult struct {

	// StatusCode, if provided, use this value, otherwise HTTP 200.
	StatusCode int

	// ContentType, if provided, use this value.
	ContentType string

	// ContentDisposition, if provided, use this value.
	ContentDisposition string

	// Headers, if provided, are added to the response
	Headers map[string][]string

	// Content, if provided, use this value, otherwise no content will be written to the response stream.
	Content io.Reader
}

// SerializeResult provides the ability render a serialized result.
type SerializeResult struct {

	// StatusCode, if provided, use this value, otherwise HTTP 200.
	StatusCode int

	// ContentType, if provided, use this value.
	ContentType string

	// Headers, if provided, are added to the response
	Headers map[string][]string

	// Content, if provided, use this value, otherwise no content will be written to the response stream.
	Content interface{}
}

func (this *SerializeResult) SetContent(value interface{}) { this.Content = value }
func (this *SerializeResult) Result() interface{}          { return this }

// JSONPResult provides the ability render a JSON-P result to the response.
type JSONPResult struct {
	// StatusCode, if provided, use this value, otherwise HTTP 200.
	StatusCode int

	// ContentType, if provided, use this value.
	ContentType string

	// Headers, if provided, are added to the response
	Headers map[string][]string

	// Content, if provided, use this value, otherwise no content will be written to the response stream.
	Content interface{}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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

	// Context represents a error-specific value in the form a serializable instance that the client can use to get more understanding about the error itself.
	Context any `json:"context,omitempty"`
}

// InputErrors represents a set of problems with the calling HTTP request.
type InputErrors struct {
	Errors []error `json:"errors,omitempty"`
}

func (this InputError) Error() string { return this.Message }

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// BaseInputModel allows enables struct embedding such that other InputModels don't necessarily need to re-implement each method.
type BaseInputModel struct{}

func (*BaseInputModel) Reset()                   {}
func (*BaseInputModel) Bind(*http.Request) error { return nil }
func (*BaseInputModel) Validate([]error) int     { return 0 }

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type fixedResultContainer struct{ ResultContainer }
type bindErrorContainer struct{ *SerializeResult }
type validationErrorContainer struct{ *SerializeResult }

func (this *fixedResultContainer) SetContent(interface{}) {} // no-op
func (this *fixedResultContainer) Result() interface{}    { return this.ResultContainer }

func (this *bindErrorContainer) SetContent(value interface{}) {
	inputErrors := this.SerializeResult.Content.(*InputErrors)
	inputErrors.Errors = inputErrors.Errors[0:0]
	inputErrors.Errors = append(inputErrors.Errors, value.(error))
}
func (this *bindErrorContainer) Result() interface{} { return this.SerializeResult }

func (this *validationErrorContainer) SetContent(value interface{}) {
	inputErrors := this.SerializeResult.Content.(*InputErrors)
	inputErrors.Errors = inputErrors.Errors[0:0]
	inputErrors.Errors = value.([]error)
}
func (this *validationErrorContainer) Result() interface{} { return this.SerializeResult }

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	notAcceptableResult = &TextResult{
		StatusCode:  http.StatusNotAcceptable,
		ContentType: mimeTypeApplicationJSONUTF8,
		Content: _serializeJSON(InputErrors{
			Errors: []error{
				InputError{
					Fields:  []string{"header:Accept"},
					Name:    "invalid-accept-header",
					Message: "Unable to represent the application results using the Accept type.",
				},
			},
		}),
	}
	unsupportedMediaTypeResult = &SerializeResult{
		StatusCode: http.StatusUnsupportedMediaType,
		Content: InputErrors{
			Errors: []error{
				InputError{
					Fields:  []string{"header:Content-Type"},
					Name:    "invalid-content-type-header",
					Message: "The content type specified, if any, was not recognized.",
				},
			},
		},
	}
	deserializationResultFactory = func() ResultContainer {
		return &fixedResultContainer{
			ResultContainer: &SerializeResult{
				StatusCode: http.StatusBadRequest,
				Content: InputErrors{
					Errors: []error{
						InputError{
							Fields:  []string{"body"},
							Name:    "malformed-request-payload",
							Message: "The body did not contain well-formed data and could not be properly deserialized.",
						},
					},
				},
			},
		}
	}
	parseFormedFailedResult = &SerializeResult{
		StatusCode: http.StatusBadRequest,
		Content: InputErrors{
			Errors: []error{
				InputError{
					Fields:  []string{"form"},
					Name:    "invalid-form-data",
					Message: "The form data provided was not valid and could not be parsed.",
				},
			},
		},
	}
	bindErrorResultFactory = func() ResultContainer {
		return &bindErrorContainer{
			SerializeResult: &SerializeResult{
				StatusCode: http.StatusBadRequest,
				Content:    &InputErrors{},
			},
		}
	}

	validationResultFactory = func() ResultContainer {
		return &validationErrorContainer{
			SerializeResult: &SerializeResult{
				StatusCode: http.StatusUnprocessableEntity,
				Content:    &InputErrors{},
			},
		}
	}
)

func _serializeJSON(instance interface{}) string {
	raw, _ := json.Marshal(instance)
	return string(raw)
}
