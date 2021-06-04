package shuttle

import (
	"io"
	"net/http"
)

type defaultWriter struct {
	serializers       map[string]Serializer
	defaultSerializer Serializer
	bodyBuffer        []byte
	headerBuffer      []string
	serializeBuffer   *SerializeResult
}

func newWriter(serializerFactories map[string]func() Serializer) Writer {
	serializers := make(map[string]Serializer, len(serializerFactories))
	for acceptType, callback := range serializerFactories {
		serializers[acceptType] = callback()
	}

	return &defaultWriter{
		serializers:       serializers,
		defaultSerializer: serializers[defaultSerializerContentType],
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
		this.writeTextResult(response, typed)
	case TextResult:
		this.writeTextResult(response, &typed)

	case *BinaryResult:
		this.writeBinaryResult(response, typed)
	case BinaryResult:
		this.writeBinaryResult(response, &typed)

	case *StreamResult:
		this.writeStreamResult(response, typed)
	case StreamResult:
		this.writeStreamResult(response, &typed)

	case *SerializeResult:
		this.writeSerializeResult(response, request, typed)
	case SerializeResult:
		this.writeSerializeResult(response, request, &typed)

	case *FixedContentResult:
		this.write(response, request, typed.ContentResult)

	case string:
		writeStringResult(response, typed)
	case []byte:
		writeByteResult(response, typed)
	case bool:
		writeBoolResult(response, typed)

	default:
		this.serializeBuffer.Content = result
		this.writeSerializeResult(response, request, this.serializeBuffer)
	}
}

func (this *defaultWriter) writeTextResult(response http.ResponseWriter, typed *TextResult) {
	hasContent := len(typed.Content) > 0
	this.writeHeader(response, typed.StatusCode, typed.ContentType, hasContent)
	if hasContent {
		_, _ = io.WriteString(response, typed.Content)
	}
}
func (this *defaultWriter) writeBinaryResult(response http.ResponseWriter, typed *BinaryResult) {
	hasContent := len(typed.Content) > 0
	this.writeHeader(response, typed.StatusCode, typed.ContentType, hasContent)
	if hasContent {
		_, _ = response.Write(typed.Content)
	}
}
func (this *defaultWriter) writeStreamResult(response http.ResponseWriter, typed *StreamResult) {
	hasContent := typed.Content != nil
	this.writeHeader(response, typed.StatusCode, typed.ContentType, hasContent)
	if hasContent {
		_, _ = io.CopyBuffer(response, typed.Content, this.bodyBuffer)
	}
}
func (this *defaultWriter) writeSerializeResult(response http.ResponseWriter, request *http.Request, typed *SerializeResult) {
	hasContent := typed.Content != nil

	serializer := this.loadSerializer(request.Header[headerAccept])
	contentType := typed.ContentType
	if len(contentType) == 0 {
		contentType = serializer.ContentType()
	}

	this.writeHeader(response, typed.StatusCode, contentType, hasContent)
	if hasContent {
		_ = this.defaultSerializer.Serialize(response, typed.Content)
	}
}
func (this *defaultWriter) loadSerializer(acceptTypes []string) Serializer {
	for _, acceptType := range acceptTypes {
		if serializer, contains := this.serializers[normalizeMediaType(acceptType)]; contains {
			return serializer
		}
	}

	return this.defaultSerializer
}

func writeStringResult(response http.ResponseWriter, typed string) {
	if len(typed) > 0 {
		_, _ = io.WriteString(response, typed)
	}
}
func writeByteResult(response http.ResponseWriter, typed []byte) {
	if len(typed) > 0 {
		_, _ = response.Write(typed)
	}
}
func writeBoolResult(response http.ResponseWriter, typed bool) {
	if typed {
		_, _ = io.WriteString(response, "true")
	} else {
		_, _ = io.WriteString(response, "false")
	}
}

func (this *defaultWriter) writeHeader(response http.ResponseWriter, statusCode int, contentType string, hasContent bool) {
	if hasContent && len(contentType) > 0 {
		this.headerBuffer[0] = contentType
		response.Header()[headerContentType] = this.headerBuffer
	}

	if statusCode > 0 {
		response.WriteHeader(statusCode)
	}
}
