package shuttle

import (
	"net/http"
	"sync"
)

func NewHandler(options ...Option) http.Handler {
	config := newConfig(options)
	if config.LongLivedPoolCapacity == 0 {
		return newSemiPersistentHandler(options)
	}

	return newPersistentHandler(config)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type persistentHandler struct {
	buffer chan http.Handler
}

func newPersistentHandler(config configuration) http.Handler {
	buffer := make(chan http.Handler, config.LongLivedPoolCapacity)
	for i := 0; i < config.LongLivedPoolCapacity; i++ {
		buffer <- newTransientHandlerFromConfig(config)
	}

	return &persistentHandler{buffer: buffer}
}

func (this *persistentHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	handler := <-this.buffer
	defer func() { this.buffer <- handler }()
	handler.ServeHTTP(response, request)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type semiPersistentHandler struct {
	buffer *sync.Pool
}

func newSemiPersistentHandler(options []Option) http.Handler {
	buffer := &sync.Pool{New: func() interface{} {
		// The config is a "shared nothing" style wherein each handler gets its own configuration values which include
		// callbacks to stateful error writers and stateful serializers.
		config := newConfig(options)
		return newTransientHandlerFromConfig(config)
	}}

	return &semiPersistentHandler{buffer: buffer}
}

func (this *semiPersistentHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	handler := this.buffer.Get().(http.Handler)
	defer this.buffer.Put(handler)
	handler.ServeHTTP(response, request)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type transientHandler struct {
	input     InputModel
	readers   []Reader
	processor Processor
	writer    Writer
	monitor   Monitor
}

func newTransientHandlerFromConfig(config configuration) http.Handler {
	readers := make([]Reader, 0, len(config.Readers))
	for _, readerFactory := range config.Readers {
		readers = append(readers, readerFactory())
	}

	return newTransientHandler(config.InputModel(), readers, config.Processor(), config.Writer(), config.Monitor)
}
func newTransientHandler(input InputModel, readers []Reader, processor Processor, writer Writer, monitor Monitor) http.Handler {
	monitor.HandlerCreated()
	return &transientHandler{
		input:     input,
		readers:   readers,
		processor: processor,
		writer:    writer,
		monitor:   monitor,
	}
}

func (this *transientHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	this.monitor.RequestReceived()
	result := this.process(request)
	this.writer.Write(response, request, result)
}
func (this *transientHandler) process(request *http.Request) interface{} {
	this.input.Reset()

	for _, reader := range this.readers {
		if result := reader.Read(this.input, request); result != nil {
			return result
		}
	}

	// FUTURE: if the context is cancelled, don't bother rendering a response
	return this.processor.Process(request.Context(), this.input)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// TODO:
// 404 Not Found (JSON/XML/etc. output)
// 405 Method Not Allowed (JSON/XML/etc. output)
