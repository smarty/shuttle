package shuttle

import (
	"net/http"
	"sync"
)

type poolHandler struct {
	pool    *sync.Pool
	monitor Monitor
}

func NewHandler(options ...Option) http.Handler {
	pool := &sync.Pool{New: func() interface{} {
		// The config is a "shared nothing" style wherein each handler gets its own configuration values which include
		// callbacks to stateful error writers and stateful serializers.
		return newHandlerFromOptions(options)
	}}

	return &poolHandler{pool: pool}
}

func (this *poolHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	handler := this.pool.Get().(http.Handler)
	defer this.pool.Put(handler)
	handler.ServeHTTP(response, request)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type defaultHandler struct {
	input     InputModel
	readers   []Reader
	processor Processor
	writer    Writer
	monitor   Monitor
}

func newHandler(input InputModel, readers []Reader, processor Processor, writer Writer, monitor Monitor) http.Handler {
	monitor.HandlerCreated()
	return &defaultHandler{
		input:     input,
		readers:   readers,
		processor: processor,
		writer:    writer,
		monitor:   monitor,
	}
}

func (this *defaultHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	this.monitor.RequestReceived()
	result := this.process(request)
	this.writer.Write(response, request, result)
}
func (this *defaultHandler) process(request *http.Request) interface{} {
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
