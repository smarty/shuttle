package shuttle

import (
	"net/http"
	"strings"
)

type acceptReader struct {
	acceptable map[string][]string
	result     interface{}
}

func newAcceptReader(serializerFactories map[string]func() Serializer, result *TextResult) Reader {
	acceptable := make(map[string][]string)
	for acceptType, _ := range serializerFactories {
		acceptable[acceptType] = []string{acceptType}
	}

	return &acceptReader{acceptable: acceptable, result: result}
}

func (this *acceptReader) Read(_ InputModel, request *http.Request) interface{} {
	if normalized, found := this.findAcceptType(request.Header[headerAccept]); !found {
		return this.result
	} else {
		request.Header[headerAccept] = normalized
	}

	return nil
}
func (this *acceptReader) findAcceptType(acceptTypes []string) ([]string, bool) {
	if len(acceptTypes) == 0 {
		return nil, true
	}

	for _, value := range acceptTypes {
		var item string
		for {
			index := strings.Index(value, ",")
			if index >= 0 {
				item = value[0:index]
				value = value[index+1:]
			} else {
				item = value
			}

			if item = strings.TrimSpace(item); item == headerAcceptAnyValue {
				return nil, true // default
			} else if types, contains := this.acceptable[normalizeMediaType(item)]; contains {
				return types, true
			} else if index == -1 {
				break
			}
		}
	}

	return nil, false
}
func normalizeMediaType(value string) string {
	if index := strings.Index(value, ";"); index >= 0 {
		value = value[0:index]
	}

	return strings.TrimSpace(value)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// deserializeReader

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type bindReader struct {
	factory func(error) interface{}
}

func newBindReader(factory func(error) interface{}) Reader {
	return &bindReader{factory: factory}
}

func (this *bindReader) Read(target InputModel, request *http.Request) interface{} {
	if err := target.Bind(request); err != nil {
		return this.factory(err)
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type validationReader struct {
	buffer []error
}

func newValidationReader(bufferSize int) Reader {
	return &validationReader{buffer: make([]error, bufferSize)}
}

func (this *validationReader) Read(target InputModel, _ *http.Request) interface{} {
	if count := target.Validate(this.buffer); count > 0 {
		return this.buffer[0:count]
	}

	return nil
}
