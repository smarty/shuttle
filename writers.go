package shuttle

import (
	"cmp"
	"io"
	"net/http"
	"strings"
)

type defaultWriter struct {
	serializers              map[string]Serializer
	defaultSerializer        Serializer
	monitor                  Monitor
	bodyBuffer               []byte
	contentTypeBuffer        []string
	contentDispositionBuffer []string
	serializeBuffer          *SerializeResult
}

func newWriter(serializerFactories map[string]func() Serializer, monitor Monitor) Writer {
	serializers := make(map[string]Serializer, len(serializerFactories))
	for acceptType, callback := range serializerFactories {
		serializers[acceptType] = callback()
	}

	return &defaultWriter{
		serializers:              serializers,
		defaultSerializer:        serializers[emptyContentType],
		monitor:                  monitor,
		bodyBuffer:               make([]byte, 1024*4),
		contentTypeBuffer:        make([]string, 1),
		contentDispositionBuffer: make([]string, 1),
		serializeBuffer:          &SerializeResult{},
	}
}

func (this *defaultWriter) Write(response http.ResponseWriter, request *http.Request, result any) {
	response.Header()["Date"] = nil // remove Date header from HTTP response

	if result == nil {
		response.WriteHeader(http.StatusNoContent)
	} else if handler, ok := result.(http.Handler); ok {
		handler.ServeHTTP(response, request)
	} else {
		this.write(response, request, result)
	}
}
func (this *defaultWriter) write(response http.ResponseWriter, request *http.Request, result any) {
	switch typed := result.(type) {
	case *TextResult:
		this.responseStatus(this.writeTextResult(response, typed))
	case TextResult:
		this.responseStatus(this.writeTextResult(response, &typed))

	case *BinaryResult:
		this.responseStatus(this.writeBinaryResult(response, typed))
	case BinaryResult:
		this.responseStatus(this.writeBinaryResult(response, &typed))

	case *StreamResult:
		this.responseStatus(this.writeStreamResult(response, typed))
	case StreamResult:
		this.responseStatus(this.writeStreamResult(response, &typed))

	case *SerializeResult:
		this.responseStatus(this.writeSerializeResult(response, request, typed))
	case SerializeResult:
		this.responseStatus(this.writeSerializeResult(response, request, &typed))

	case string:
		this.responseStatus(this.writeStringResult(response, typed))
	case []byte:
		this.responseStatus(this.writeByteResult(response, typed))
	case bool:
		this.responseStatus(this.writeBoolResult(response, typed))

	default:
		this.serializeBuffer.Content = result
		this.responseStatus(this.writeSerializeResult(response, request, this.serializeBuffer))
	}
}

func (this *defaultWriter) writeTextResult(response http.ResponseWriter, typed *TextResult) (err error) {
	this.monitor.TextResult()
	hasContent := len(typed.Content) > 0

	headers := response.Header()
	for key, values := range typed.Headers {
		headers[key] = values
	}

	this.writeHeader(response, typed.StatusCode, typed.ContentType, "", hasContent)
	if hasContent {
		_, err = io.WriteString(response, typed.Content)
	}

	return err
}
func (this *defaultWriter) writeBinaryResult(response http.ResponseWriter, typed *BinaryResult) (err error) {
	this.monitor.BinaryResult()
	hasContent := len(typed.Content) > 0

	headers := response.Header()
	for key, values := range typed.Headers {
		headers[key] = values
	}

	this.writeHeader(response, typed.StatusCode, typed.ContentType, typed.ContentDisposition, hasContent)
	if hasContent {
		_, err = response.Write(typed.Content)
	}

	return err
}
func (this *defaultWriter) writeStreamResult(response http.ResponseWriter, typed *StreamResult) (err error) {
	this.monitor.StreamResult()
	hasContent := typed.Content != nil

	headers := response.Header()
	for key, values := range typed.Headers {
		headers[key] = values
	}

	this.writeHeader(response, typed.StatusCode, typed.ContentType, typed.ContentDisposition, hasContent)
	if hasContent {
		_, err = io.CopyBuffer(response, typed.Content, this.bodyBuffer)
	}

	if closer, ok := typed.Content.(io.Closer); ok {
		err = cmp.Or(err, closer.Close())
	}

	return err
}
func (this *defaultWriter) writeSerializeResult(response http.ResponseWriter, request *http.Request, typed *SerializeResult) error {
	this.monitor.SerializeResult()
	hasContent := typed.Content != nil

	serializer := this.loadSerializer(request.Header[headerAccept])
	contentType := typed.ContentType
	if len(contentType) == 0 {
		contentType = serializer.ContentType()
	}

	headers := response.Header()
	for key, values := range typed.Headers {
		headers[key] = values
	}

	this.writeHeader(response, typed.StatusCode, contentType, "", hasContent)
	if hasContent {
		if strings.Contains(request.Header.Get("Accept"), "/xml") {
			this.write(response, request, xmlPrefix)
		}
		return serializer.Serialize(response, typed.Content)
	}

	return nil
}
func (this *defaultWriter) loadSerializer(acceptTypes []string) Serializer {
	for _, acceptType := range acceptTypes {
		if serializer, contains := this.serializers[normalizeMediaType(acceptType)]; contains {
			return serializer
		}
	}

	return this.defaultSerializer
}

func (this *defaultWriter) writeStringResult(response http.ResponseWriter, typed string) (err error) {
	this.monitor.NativeResult()

	if len(typed) > 0 {
		_, err = io.WriteString(response, typed)
	}

	return err
}
func (this *defaultWriter) writeByteResult(response http.ResponseWriter, typed []byte) (err error) {
	this.monitor.NativeResult()

	if len(typed) > 0 {
		_, err = response.Write(typed)
	}

	return err
}
func (this *defaultWriter) writeBoolResult(response http.ResponseWriter, typed bool) (err error) {
	this.monitor.NativeResult()

	if typed {
		_, err = io.WriteString(response, "true")
	} else {
		_, err = io.WriteString(response, "false")
	}

	return err
}

func (this *defaultWriter) writeHeader(response http.ResponseWriter, statusCode int, contentType, contentDisposition string, hasContent bool) {
	if hasContent && len(contentType) > 0 {
		this.contentTypeBuffer[0] = contentType
		response.Header()[headerContentType] = this.contentTypeBuffer
	}
	if hasContent && len(contentDisposition) > 0 {
		this.contentDispositionBuffer[0] = contentDisposition
		response.Header()[headerContentDisposition] = this.contentDispositionBuffer
	}

	if statusCode > 0 {
		this.monitor.ResponseStatus(statusCode)
		response.WriteHeader(statusCode)
	} else {
		this.monitor.ResponseStatus(http.StatusOK)
	}
}
func (this *defaultWriter) responseStatus(err error) {
	if err != nil {
		this.monitor.ResponseFailed(err)
	}
}
