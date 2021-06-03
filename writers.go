package shuttle

import (
	"io"
	"net/http"
)

type defaultWriter struct {
	bodyBuffer   []byte
	headerBuffer []string
}

func newWriter(serializers map[string]func() Serializer) Writer {
	return &defaultWriter{
		bodyBuffer:   make([]byte, 1024*4),
		headerBuffer: make([]string, 1),
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

	case string:
		writeStringResult(response, typed)
	case []byte:
		writeByteResult(response, typed)
	case bool:
		writeBoolResult(response, typed)
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
		response.Header()["Content-Type"] = this.headerBuffer
	}

	if statusCode > 0 {
		response.WriteHeader(statusCode)
	}
}
