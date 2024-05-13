package shuttle

import (
	"context"
	"net/http"
)

type configuration struct {
	InputModel                  func() InputModel
	Processor                   func() Processor
	Deserializers               map[string]func() Deserializer
	Serializers                 map[string]func() Serializer
	VerifyAcceptHeader          bool
	ParseForm                   bool
	Bind                        bool
	Validate                    bool
	DefaultAcceptIfNotFound     bool
	LongLivedPoolCapacity       int
	MaxAcceptTypes              int
	MaxValidationErrors         int
	Readers                     []func() Reader
	Writer                      func() Writer
	NotAcceptableResult         *TextResult
	UnsupportedMediaTypeResult  any
	DeserializationFailedResult func() ResultContainer
	ParseFormFailedResult       any
	BindFailedResult            func() ResultContainer
	ValidationFailedResult      func() ResultContainer
	Monitor                     Monitor
}

func newConfig(options []option) configuration {
	this := configuration{
		Deserializers: map[string]func() Deserializer{},
		Serializers:   map[string]func() Serializer{},
	}
	Options.apply(options...)(&this)
	return this
}

// InputModel provides the callback that returns a unique instance of an InputModel on each invocation.
func (singleton) InputModel(value func() InputModel) option {
	return func(this *configuration) { this.InputModel = value }
}

// ProcessorSharedInstance is used when the instance provided has no shared, mutable state and the instance can be
// shared between all requests.
func (singleton) ProcessorSharedInstance(value Processor) option {
	return Options.Processor(func() Processor { return value })
}

// Processor is used when a long-lived, reusable, and stateful processor (and associated component tree) is created to
// service many requests, each request going through a pooled instance of that processor.
func (singleton) Processor(value func() Processor) option {
	return func(this *configuration) { this.Processor = value }
}

// DeserializeJSON indicates that the JSON decoder from the Go standard library should be used to deserialize HTTP
// request bodies which contain JSON.
func (singleton) DeserializeJSON(value bool) option {
	return func(this *configuration) {
		if value {
			Options.Deserializer(mimeTypeApplicationJSON, func() Deserializer { return newJSONDeserializer() })(this)
		} else {
			delete(this.Deserializers, mimeTypeApplicationJSON)
		}
	}
}

// DeserializeXML indicates that the XML decoder from the Go standard library should be used to deserialize HTTP
// request bodies which contain XML.
func (singleton) DeserializeXML(value bool) option {
	return func(this *configuration) {
		if value {
			Options.Deserializer(mimeTypeApplicationXML, func() Deserializer { return newXMLDeserializer() })(this)
		} else {
			delete(this.Deserializers, mimeTypeApplicationXML)
		}
	}
}

// Deserializer registers a callback which provides a unique instance of a deserializer per invocation and associates it
// with the content type value provided. If the deserializer contains any shared, mutable state, it must return a unique
// instance per invocation. If the deserializer only contains immutable state (or no state at all), then invocations of
// the callback provided may return the same instance.
func (singleton) Deserializer(contentType string, value func() Deserializer) option {
	return func(this *configuration) { this.Deserializers[contentType] = value }
}

// DefaultDeserializer registers a deserializer to be used for requests that do not provide any HTTP Accept request header
// or for those which contain the wildcard Accept header value of "*/*". If the deserializer callback yields an instance
// with mutable state, then each invocation must give back a unique instance. If the serializer doesn't contain any
// mutable state, then the same instance may be returned between each invocation of the callback provided.
func (singleton) DefaultDeserializer(value func() Deserializer) option {
	return func(this *configuration) { Options.Deserializer(emptyContentType, value)(this) }
}

// DefaultSerializer registers a serializer to be used for requests that do not provide any HTTP Accept request header
// or for those which contain the wildcard Accept header value of "*/*". If the serializer callback yields an instance
// with mutable state, then each invocation must give back a unique instance. If the serializer doesn't contain any
// mutable state, then the same instance may be returned between each invocation of the callback provided.
func (singleton) DefaultSerializer(value func() Serializer) option {
	return func(this *configuration) { Options.Serializer(emptyContentType, value)(this) }
}

// SerializeJSON indicates that the JSON encoder from the Go standard library should be used to serialize results into
// the HTTP response stream using JSON.
func (singleton) SerializeJSON(value bool) option {
	return func(this *configuration) {
		if value {
			Options.Serializer(mimeTypeApplicationJSON, func() Serializer { return newJSONSerializer() })(this)
		} else {
			delete(this.Serializers, mimeTypeApplicationJSON)
		}
	}
}

// SerializeXML indicates that the XML encoder from the Go standard library should be used to serialize results into
// the HTTP response stream using XML.
func (singleton) SerializeXML(value bool) option {
	return func(this *configuration) {
		if value {
			Options.Serializer(mimeTypeApplicationXML, func() Serializer { return newXMLSerializer() })(this)
			Options.Serializer(mimeTypeApplicationTextXML, func() Serializer { return newXMLSerializer() })(this)
		} else {
			delete(this.Serializers, mimeTypeApplicationXML)
		}
	}
}

// Serializer registers a callback which provides a unique instance of a serializer per invocation and associates it
// with the content type value provided. If the serializer contains any mutable state, it must return a unique instance
// per invocation. If the serializer only contains immutable state (or no state at all), then invocations of the
// callback provided may return the same instance.
func (singleton) Serializer(contentType string, value func() Serializer) option {
	return func(this *configuration) { this.Serializers[contentType] = value }
}

// VerifyAcceptHeader indicates whether to inspect the Accept HTTP request header and to assert that it is both
// recognized and understood before continuing further.
func (singleton) VerifyAcceptHeader(value bool) option {
	return func(this *configuration) { this.VerifyAcceptHeader = value }
}

// ParseForm indicates whether to call ParseForm() on the incoming HTTP request.
func (singleton) ParseForm(value bool) option {
	return func(this *configuration) { this.ParseForm = value }
}

// Bind indicates whether to forward the raw HTTP request into the InputModel to bind parts of the request onto
// a pooled instanced of the InputModel configured for this route.
func (singleton) Bind(value bool) option {
	return func(this *configuration) { this.Bind = value }
}

// DefaultAcceptIfNotFound indicates whether to use the default serializer if no Accept types were acceptable.
func (singleton) DefaultAcceptIfNotFound(value bool) option {
	return func(this *configuration) { this.DefaultAcceptIfNotFound = value }
}

// MaxAcceptTypes defines the number of Accept sub values to consider before using the default serializer.
// A value of -1 means to consider all Accept sub values.
func (singleton) MaxAcceptTypes(value int) option {
	return func(this *configuration) { this.MaxAcceptTypes = value }
}

// Validate indicates whether to ask the pool instance of the InputModel associated with this request if it is in a
// valid state.
func (singleton) Validate(value bool) option {
	return func(this *configuration) { this.Validate = value }
}

// LongLivedPoolCapacity indicates that the handler should be managed by a long-lived pool rather than a short-term
// auto-garbage collected sync.Pool. This means that any pre-allocated resources will share their lifecycle scope with
// the http.Handler itself. Further, any HTTP requests against an empty pool will block.
func (singleton) LongLivedPoolCapacity(value uint16) option {
	return func(this *configuration) { this.LongLivedPoolCapacity = int(value) }
}

// MaxValidationErrors indicates the number of unique slots to pre-allocate to receive errors with the pooled input
// model associated with this route. It is suggested to set this value to a large enough number to be able to
// accommodate the maximum number of errors possible for the InputModel associated with this route.
func (singleton) MaxValidationErrors(value uint16) option {
	return func(this *configuration) { this.MaxValidationErrors = int(value) }
}

// Writer registers a callback the get an instance of a Writer used to render the actual HTTP response. If the instance
// of the Writer contains any mutable state, then each invocation of the callback must provide a unique instance. If the
// Writer is stateless or only contains shared, read-only state (along with all of all structures contained therein
// down the entire object graph), then the same instance can be returned each time.
func (singleton) Writer(value func() Writer) option {
	return func(this *configuration) { this.Writer = value }
}

// UnsupportedMediaTypeResult registers the result to be written to the underlying HTTP response stream to indicate when
// the values in the provided Content-Type HTTP request header are not recognized. A single, shared instance of this
// instance can be provided across all routes.
func (singleton) UnsupportedMediaTypeResult(value any) option {
	return func(this *configuration) { this.UnsupportedMediaTypeResult = value }
}

// DeserializationFailedResult registers the result to be written to the underlying HTTP response stream to indicate
// when the HTTP request body cannot be deserialized into the configured InputModel.
func (singleton) DeserializationFailedResult(value func() ResultContainer) option {
	return func(this *configuration) { this.DeserializationFailedResult = value }
}

// ParseFormFailedResult registers the result to be written to the underlying HTTP response stream to indicate when
// parsing the form and query fields of the HTTP request have failed.
func (singleton) ParseFormFailedResult(value any) option {
	return func(this *configuration) { this.ParseFormFailedResult = value }
}

// BindFailedResult registers the result to be written to the underlying HTTP response stream to indicate when the HTTP
// request cannot be properly bound or mapped onto the configured InputModel.
func (singleton) BindFailedResult(value func() ResultContainer) option {
	return func(this *configuration) { this.BindFailedResult = value }
}

// ValidationFailedResult registers the result to be written to the underlying HTTP response stream to indicate when the
// HTTP request contains validation errors according to the provided InputModel.
func (singleton) ValidationFailedResult(value func() ResultContainer) option {
	return func(this *configuration) { this.ValidationFailedResult = value }
}

// NotAcceptableResult registers the result to be written to the underlying HTTP response stream to indicate when all of
// the values in the provided Accept HTTP request header are not recognized. A single, shared instance of this
// instance can be provided across all routes. This must be a TextResult to ensure that it can be properly rendered.
func (singleton) NotAcceptableResult(value *TextResult) option {
	return func(this *configuration) { this.NotAcceptableResult = value }
}

// Monitor registers a mechanism to watch the internals of the library and to gather metrics when the various behaviors
// occur.
func (singleton) Monitor(value Monitor) option {
	return func(this *configuration) { this.Monitor = value }
}

func (singleton) apply(options ...option) option {
	return func(this *configuration) {
		for _, item := range Options.defaults(options...) {
			item(this)
		}

		if this.VerifyAcceptHeader {
			this.Readers = append(this.Readers, func() Reader {
				return newAcceptReader(this.Serializers, this.NotAcceptableResult, this.DefaultAcceptIfNotFound, this.MaxAcceptTypes, this.Monitor)
			})
		}

		if len(this.Deserializers) > 0 {
			this.Readers = append(this.Readers, func() Reader {
				return newDeserializeReader(this.Deserializers, this.UnsupportedMediaTypeResult, this.DeserializationFailedResult(), this.Monitor)
			})
		}

		if this.ParseForm {
			this.Readers = append(this.Readers, func() Reader { return newParseFormReader(this.ParseFormFailedResult, this.Monitor) })
		}

		if this.Bind {
			this.Readers = append(this.Readers, func() Reader { return newBindReader(this.BindFailedResult(), this.Monitor) })
		}

		if this.Validate {
			this.Readers = append(this.Readers, func() Reader {
				return newValidateReader(this.ValidationFailedResult(), this.MaxValidationErrors, this.Monitor)
			})
		}

		if this.Writer == nil {
			this.Writer = func() Writer { return newWriter(this.Serializers, this.Monitor) }
		}
	}
}
func (singleton) defaults(options ...option) []option {
	return append([]option{
		Options.InputModel(func() InputModel { return &nop{} }),
		Options.ProcessorSharedInstance(&nop{}),

		Options.VerifyAcceptHeader(true),
		Options.ParseForm(false),
		Options.Bind(true),
		Options.Validate(true),
		Options.MaxValidationErrors(32),
		Options.DefaultAcceptIfNotFound(false),
		Options.MaxAcceptTypes(-1),

		Options.SerializeJSON(true),
		Options.SerializeXML(false),
		Options.DefaultSerializer(func() Serializer { return newJSONSerializer() }),

		Options.Writer(nil),

		Options.NotAcceptableResult(notAcceptableResult),
		Options.UnsupportedMediaTypeResult(unsupportedMediaTypeResult),
		Options.ParseFormFailedResult(parseFormedFailedResult),
		Options.DeserializationFailedResult(deserializationResultFactory),
		Options.BindFailedResult(bindErrorResultFactory),
		Options.ValidationFailedResult(validationResultFactory),

		Options.Monitor(&nopMonitor{}),
	}, options...)
}

type singleton struct{}
type option func(*configuration)

var Options singleton

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type nop struct{}
type nopMonitor struct{}

func (*nop) Process(context.Context, any) any { return nil }

func (*nop) Reset()                   {}
func (*nop) Bind(*http.Request) error { return nil }
func (*nop) Validate([]error) int     { return 0 }

func (*nopMonitor) HandlerCreated()        {}
func (*nopMonitor) RequestReceived()       {}
func (*nopMonitor) NotAcceptable()         {}
func (*nopMonitor) UnsupportedMediaType()  {}
func (*nopMonitor) Deserialize()           {}
func (*nopMonitor) DeserializeFailed()     {}
func (*nopMonitor) ParseForm()             {}
func (*nopMonitor) ParseFormFailed(error)  {}
func (*nopMonitor) Bind()                  {}
func (*nopMonitor) BindFailed(error)       {}
func (*nopMonitor) Validate()              {}
func (*nopMonitor) ValidateFailed([]error) {}
func (*nopMonitor) TextResult()            {}
func (*nopMonitor) BinaryResult()          {}
func (*nopMonitor) StreamResult()          {}
func (*nopMonitor) SerializeResult()       {}
func (*nopMonitor) NativeResult()          {}
func (*nopMonitor) SerializeFailed()       {}
func (*nopMonitor) ResponseStatus(int)     {}
func (*nopMonitor) ResponseFailed(error)   {}
