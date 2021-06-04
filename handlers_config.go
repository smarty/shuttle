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
	Bind                        bool
	Validate                    bool
	MaxValidationErrors         int
	Readers                     []func() Reader
	Writer                      func() Writer
	UnsupportedMediaTypeResult  interface{}
	DeserializationFailedResult func(error) interface{}
	BindFailedResult            func(error) interface{}
	ValidationFailedResult      func([]error) interface{}
	NotAcceptableResult         *TextResult
}

func newHandlerFromOptions(options []Option) http.Handler {
	config := newConfig(options)

	readers := make([]Reader, 0, len(config.Readers))
	for _, readerFactory := range config.Readers {
		readers = append(readers, readerFactory())
	}

	return newHandler(config.InputModel(), readers, config.Processor(), config.Writer())
}

func newConfig(options []Option) configuration {
	this := configuration{
		Deserializers: map[string]func() Deserializer{},
		Serializers:   map[string]func() Serializer{},
	}
	Options.apply(options...)(&this)
	return this
}

// InputModel provides the callback that returns a unique instance on each invocation.
func (singleton) InputModel(value func() InputModel) Option {
	return func(this *configuration) { this.InputModel = value }
}

// ProcessorSharedInstance is used when the instance provided has no shared, mutable state and the instance can be
// shared between all requests.
func (singleton) ProcessorSharedInstance(value Processor) Option {
	return Options.Processor(func() Processor { return value })
}

// Processor is used when an long-lived, reusable, and stateful processor (and associated component tree) is created to
// service many requests, each request going through a pooled instance of that processor.
func (singleton) Processor(value func() Processor) Option {
	return func(this *configuration) { this.Processor = value }
}

// DeserializeJSON indicates that the JSON decoder from the Go standard library should be used to deserialize HTTP
// request bodies which contain JSON.
func (singleton) DeserializeJSON(value bool) Option {
	return func(this *configuration) {
		if value {
			Options.Deserializer(mimeTypeApplicationJSON, func() Deserializer { return newJSONDeserializer() })(this)
		} else {
			delete(this.Deserializers, mimeTypeApplicationJSON)
		}
	}
}

// Deserializer registers a callback which provides a unique instance of a deserializer per invocation and associates it
// with the content type value provided. If the deserializer contains any shared, mutable state, it must return a unique
// instance per invocation. If the deserializer only contains immutable state (or no state at all), then invocations of
// the callback provided may return the same instance.
func (singleton) Deserializer(contentType string, value func() Deserializer) Option {
	return func(this *configuration) { this.Deserializers[contentType] = value }
}

// DefaultSerializer registers a serializer to be used for requests that do not provide any HTTP Accept request header
// or for those which contain the wildcard Accept header value of "*/*". If the serializer callback yields an instance
// with mutable state, then each invocation must given back a unique instance. If the serializer doesn't contain any
// mutable state, then the same instance may be returned between each invocation of the callback provided.
func (singleton) DefaultSerializer(value func() Serializer) Option {
	return func(this *configuration) {
		if value != nil {
			Options.Serializer("", value)(this)
		} else {
			delete(this.Serializers, "")
		}
	}
}

// SerializeJSON indicates that the JSON encoder from the Go standard library should be used to serialize results into
// the HTTP response stream using JSON.
func (singleton) SerializeJSON(value bool) Option {
	return func(this *configuration) {
		if value {
			Options.Serializer(mimeTypeApplicationJSON, func() Serializer { return newJSONSerializer() })(this)
		} else {
			delete(this.Serializers, mimeTypeApplicationJSON)
		}
	}
}

// Serializer registers a callback which provides a unique instance of a serializer per invocation and associates it
// with the content type value provided. If the serializer contains any mutable state, it must return a unique instance
// per invocation. If the serializer only contains immutable state (or no state at all), then invocations of the
// callback provided may return the same instance.
func (singleton) Serializer(contentType string, value func() Serializer) Option {
	return func(this *configuration) { this.Serializers[contentType] = value }
}

// VerifyAcceptHeader indicates whether or not to inspect the Accept HTTP request header and to assert that it is both
// recognized and understood before continuing further.
func (singleton) VerifyAcceptHeader(value bool) Option {
	return func(this *configuration) { this.VerifyAcceptHeader = value }
}

// Bind indicates whether or not to forward the raw HTTP request into the input model to bind parts of the request onto
// a pooled instanced of the input model configured for this route.
func (singleton) Bind(value bool) Option {
	return func(this *configuration) { this.Bind = value }
}

// Validate indicates whether or not ask the pool instance of the input model associated with this request if it is in a
// valid state.
func (singleton) Validate(value bool) Option {
	return func(this *configuration) { this.Validate = value }
}

// MaxValidationErrors indicates the number of unique slots to pre-allocate to receive errors with the pooled input
// model associated with this route. It is suggested to set this value to a large enough number to be able to
// accommodate the maximum number of errors possible for the input model associated with this route.
func (singleton) MaxValidationErrors(value uint16) Option {
	return func(this *configuration) { this.MaxValidationErrors = int(value) }
}

// Writer registers a callback the get an instance of a Writer used to render the actual HTTP response. If the instance
// of the Writer contains any mutable state, then each invocation of the callback must provide a unique instance. If the
// Writer is stateless or only contains shared, read-only state (along with all of all structures contained therein
// down the entire object graph), then the same instance can be returned each time.
func (singleton) Writer(value func() Writer) Option {
	return func(this *configuration) { this.Writer = value }
}

// UnsupportedMediaTypeResult registers the result to be written to the underlying HTTP response stream to indicate when
// the values in the provided Content-Type HTTP request header are not recognized. A single, shared instance of this
// instance can be provided across all routes.
func (singleton) UnsupportedMediaTypeResult(value interface{}) Option {
	return func(this *configuration) { this.UnsupportedMediaTypeResult = value }
}

// DeserializationFailedResult registers the result to be written to the underlying HTTP response stream to indicate
// when the HTTP request body cannot be deserialized into the configured InputModel. If the callback receives the error
// provided and attaches it an instance of the result, then a unique instance must be provided per invocation.
func (singleton) DeserializationFailedResult(value func(error) interface{}) Option {
	return func(this *configuration) { this.DeserializationFailedResult = value }
}

// BindFailedResult registers the result to be written to the underlying HTTP response stream to indicate when the HTTP
// request cannot be properly bound or mapped onto the configured InputModel. If the callback receives the error
// provided and attaches it an instance of the result, then a unique instance must be provided per invocation.
func (singleton) BindFailedResult(value func(error) interface{}) Option {
	return func(this *configuration) { this.BindFailedResult = value }
}

// ValidationFailedResult registers the result to be written to the underlying HTTP response stream to indicate when the
// HTTP request contains validation errors according to the provided InputModel. If the callback receives the error
// slice provided and attaches it an instance of the result, then a unique instance must be provided per invocation.
func (singleton) ValidationFailedResult(value func([]error) interface{}) Option {
	return func(this *configuration) { this.ValidationFailedResult = value }
}

// NotAcceptableResult registers the result to be written to the underlying HTTP response stream to indicate when all of
// the values in the provided Accept HTTP request header are not recognized. A single, shared instance of this
// instance can be provided across all routes. This must be a TextResult to ensure that it can be properly rendered.
func (singleton) NotAcceptableResult(value *TextResult) Option {
	return func(this *configuration) { this.NotAcceptableResult = value }
}

func (singleton) apply(options ...Option) Option {
	return func(this *configuration) {
		for _, item := range Options.defaults(options...) {
			item(this)
		}

		if this.VerifyAcceptHeader {
			this.Readers = append(this.Readers, func() Reader { return newAcceptReader(this.Serializers, this.NotAcceptableResult) })
		}

		if len(this.Deserializers) > 0 {
			this.Readers = append(this.Readers, func() Reader {
				return newDeserializeReader(this.Deserializers, this.UnsupportedMediaTypeResult, this.DeserializationFailedResult)
			})
		}

		if this.Bind {
			this.Readers = append(this.Readers, func() Reader { return newBindReader(this.BindFailedResult) })
		}

		if this.Validate {
			this.Readers = append(this.Readers, func() Reader { return newValidateReader(this.ValidationFailedResult, this.MaxValidationErrors) })
		}

		if this.Writer == nil {
			this.Writer = func() Writer { return newWriter(this.Serializers) }
		}
	}
}
func (singleton) defaults(options ...Option) []Option {
	// TODO: if this isn't a pointer, does the ValidationFailedResult option below make a copy of this
	// if not, we have a concurrency error
	validationFailedResult := SerializeResult{StatusCode: http.StatusUnprocessableEntity}

	return append([]Option{
		Options.InputModel(func() InputModel { return &nop{} }),
		Options.ProcessorSharedInstance(&nop{}),

		Options.VerifyAcceptHeader(true),
		Options.Bind(true),
		Options.Validate(true),
		Options.MaxValidationErrors(32),

		Options.SerializeJSON(true),
		Options.DefaultSerializer(func() Serializer { return newJSONSerializer() }),

		Options.Writer(nil),

		Options.UnsupportedMediaTypeResult(unsupportedMediaTypeResult),
		Options.DeserializationFailedResult(func(error) interface{} { return deserializationResult }),
		Options.BindFailedResult(func(error) interface{} { return bindFailedResult }),
		Options.ValidationFailedResult(func(errs []error) interface{} { validationFailedResult.Content = errs; return validationFailedResult }),
		Options.NotAcceptableResult(notAcceptableResult),
	}, options...)
}

type singleton struct{}
type Option func(*configuration)

var Options singleton

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type nop struct{}

func (*nop) Process(context.Context, interface{}) interface{} { return nil }
func (*nop) Reset()                                           {}
func (*nop) Bind(*http.Request) error                         { return nil }
func (*nop) Validate([]error) int                             { return 0 }
