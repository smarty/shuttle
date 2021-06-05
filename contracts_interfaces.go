package shuttle

import (
	"context"
	"errors"
	"io"
	"net/http"
)

// InputModel represents user input from the HTTP request. Generally speaking, each type represents an operation,
// intention, or instruction. Each intention or operation should be a separate, well-named structure. By design in this
// library, each instance will be reusable and long lived.
//
// As a best practice or design note and to assert application correctness, the fields of any given InputModel should be
// explicitly initialized and populated using garbage or junk values to ensure they are properly Reset.
type InputModel interface {
	// Reset clears the contents of the instance and prepares it for the next use.
	Reset()
	// Bind maps the values on the HTTP request provided to fields on the instance. If there are any problems and error
	// is returned which can then be rendered on the response using the configured callback.
	Bind(*http.Request) error
	// Validate ensures that the instance has all the necessary fields appropriately populated. The slice of errors
	// provided is a pre-allocated buffer in which to place any errors encountered during validation. It returns the
	// number of errors placed into the buffer provided.
	Validate([]error) int
}

// Reader provides the ability to read values from the incoming HTTP request and to either manipulate the associated
// InputModel in some fashion or to otherwise short-circuit the request pipeline by returning a result to be rendered
// the caller's HTTP response stream. If a nil (meaning successful) result is returned, then processing continues.
type Reader interface {
	Read(InputModel, *http.Request) interface{}
}

// ContentResult sets the content on the underlying instance.
type ContentResult interface {
	SetContent(interface{})
	Result() interface{}
}

// Processor represents the mechanism used to carry out the desired instruction or user-provided intention. The second
// value is deliberately left opaque to reduce library dependencies and to encourage proper type discovery of the
// InputModel provided.
//
// Depending upon how the processor is implemented (e.g. stateless and shared vs stateful between unique requests), the
// result returned can be allocated in the Processor instance and returned or it can be a stateful field somewhere in
// the Processor's object graph.
//
// For Processors that are shared and where each response pathway is stored as a field in the Processor's object graph,
// it may be helpful to follow the InputModel pattern of initializing with garbage or junk values. This will help to
// ensure that all fields are appropriately cleared and overwritten between requests.
//
// The value returned by the processor may be a primitive type, a TextResult, BinaryResult, StreamResult,
// SerializeResult or the aforementioned types using a pointer. If the value returned implements the http.Handler
// interface, that method will be invoked to render the result directly using the underlying http.Request
// and http.ResponseWriter. If the value returned is not one of the aforementioned types, it will be serialized using
// either the requested HTTP Accept type or it will use the default serializer configured, if any.
type Processor interface {
	Process(context.Context, interface{}) interface{}
}

// Writer is responsible to render to result provided to the associated response stream.
type Writer interface {
	Write(http.ResponseWriter, *http.Request, interface{})
}

// Deserializer instances provide the ability to transform an opaque byte stream into an instance of a structure.
type Deserializer interface {
	// Deserialize renders the decodes the source stream into the instance provided. If there are any problems, an error is returned.
	Deserialize(interface{}, io.Reader) error
}

// Serializer instances provide the ability to transform an instance of a structure into a byte stream.
type Serializer interface {
	// Serialize renders the instance provided to the io.Writer. If there are any problems, an error is returned.
	Serialize(io.Writer, interface{}) error
	// ContentType returns HTTP Content-Type header that will be used when writing to the HTTP response.
	ContentType() string
}

type Monitor interface {
	HandlerCreated()
	RequestReceived()
	NotAcceptable()
	UnsupportedMediaType()
	Deserialize()
	DeserializeFailed()
	ParseForm()
	ParseFormFailed(error)
	Bind()
	BindFailed(error)
	Validate()
	ValidateFailed([]error)
	TextResult()
	BinaryResult()
	StreamResult()
	SerializeResult()
	NativeResult()
	SerializeFailed()
	ResponseStatus(int)
	ResponseFailed(error)
}

var (
	// ErrDeserializationFailure indicates that there was some kind of problem deserializing the request stream.
	ErrDeserializationFailure = errors.New("failed to deserialize the stream into the instance provided")

	// ErrSerializationFailure indicates that there was some kind of problem serializing the structure to the response stream.
	ErrSerializationFailure = errors.New("failed to serialize the instance into the stream provided")
)

const (
	mimeTypeApplicationJSON     = "application/json"
	mimeTypeApplicationJSONUTF8 = mimeTypeApplicationJSON + characterSetUTF8

	mimeTypeApplicationJavascript     = "application/javascript"
	mimeTypeApplicationJavascriptUTF8 = mimeTypeApplicationJavascript + characterSetUTF8

	characterSetUTF8 = "; charset=utf-8"

	headerContentType    = "Content-Type"
	headerAccept         = "Accept"
	headerAcceptAnyValue = "*/*"

	defaultSerializerContentType  = ""
	defaultJSONPCallbackParameter = "callback"
	defaultJSONPCallbackName      = "callback"
)

var (
	headerAcceptTypeJavascription = []string{mimeTypeApplicationJavascript}
)
