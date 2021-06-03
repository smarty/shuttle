package shuttle

import "net/http"

type defaultHandler struct {
	input     InputModel
	readers   []Reader
	processor Processor
	writer    Writer
}

func newHandler(input InputModel, readers []Reader, processor Processor, writer Writer) http.Handler {
	return &defaultHandler{
		input:     input,
		readers:   readers,
		processor: processor,
		writer:    writer,
	}
}

func (this *defaultHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
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

	return this.processor.Process(request.Context(), this.input)
}
