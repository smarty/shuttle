package shuttle

import (
	"io"
	"net/http"
)

type defaultWriter struct {
}

func newWriter(serializers map[string]func() Serializer) Writer {
	return &defaultWriter{
		// TODO
	}
}

func (this *defaultWriter) Write(response http.ResponseWriter, request *http.Request, result interface{}) {
	if result == nil {
		return
	}

	switch typed := result.(type) {
	case string:
		_, _ = io.WriteString(response, typed)
	case []byte:
		_, _ = response.Write(typed)
	case bool:
		if typed {
			_, _ = io.WriteString(response, "true")
		} else {
			_, _ = io.WriteString(response, "false")
		}
	}
}
