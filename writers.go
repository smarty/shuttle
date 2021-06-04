package shuttle

import (
	"io"
	"net/http"
)

type defaultWriter struct {
	serializers       map[string]Serializer
	defaultSerializer Serializer
	monitor           Monitor
	bodyBuffer        []byte
	headerBuffer      []string
	serializeBuffer   *SerializeResult
}

func newWriter(serializerFactories map[string]func() Serializer, monitor Monitor) Writer {
	serializers := make(map[string]Serializer, len(serializerFactories))
	for acceptType, callback := range serializerFactories {
		serializers[acceptType] = callback()
	}

	return &defaultWriter{
		serializers:       serializers,
		defaultSerializer: serializers[defaultSerializerContentType],
		monitor:           monitor,
		bodyBuffer:        make([]byte, 1024*4),
		headerBuffer:      make([]string, 1),
		serializeBuffer:   &SerializeResult{},
	}
}

func (this *defaultWriter) Write(response http.ResponseWriter, request *http.Request, result interface{}) {
	if result == nil {
		return
	} else if handler, ok := result.(http.Handler); ok {
		handler.ServeHTTP(response, request)
	} else {
		this.write(response, request, result)
	}
}
func (this *defaultWriter) write(response http.ResponseWriter, request *http.Request, result interface{}) {
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
	this.writeHeader(response, typed.StatusCode, typed.ContentType, hasContent)
	if hasContent {
		_, err = io.WriteString(response, typed.Content)
	}

	return err
}
func (this *defaultWriter) writeBinaryResult(response http.ResponseWriter, typed *BinaryResult) (err error) {
	this.monitor.BinaryResult()
	hasContent := len(typed.Content) > 0
	this.writeHeader(response, typed.StatusCode, typed.ContentType, hasContent)
	if hasContent {
		_, err = response.Write(typed.Content)
	}

	return err
}
func (this *defaultWriter) writeStreamResult(response http.ResponseWriter, typed *StreamResult) (err error) {
	this.monitor.StreamResult()
	hasContent := typed.Content != nil
	this.writeHeader(response, typed.StatusCode, typed.ContentType, hasContent)
	if hasContent {
		_, err = io.CopyBuffer(response, typed.Content, this.bodyBuffer)
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

	this.writeHeader(response, typed.StatusCode, contentType, hasContent)
	if hasContent {
		return this.defaultSerializer.Serialize(response, typed.Content)
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

func (this *defaultWriter) writeHeader(response http.ResponseWriter, statusCode int, contentType string, hasContent bool) {
	if hasContent && len(contentType) > 0 {
		this.headerBuffer[0] = contentType
		response.Header()[headerContentType] = this.headerBuffer
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
