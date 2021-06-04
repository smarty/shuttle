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

type deserializeReader struct {
	available                  map[string]Deserializer
	unsupportedMediaTypeResult interface{}
	resultFactory              func(error) interface{}
}

func newDeserializeReader(deserializerFactories map[string]func() Deserializer, unsupportedMediaTypeResult interface{}, resultFactory func(error) interface{}) Reader {
	available := make(map[string]Deserializer, len(deserializerFactories))
	for contentType, factory := range deserializerFactories {
		available[contentType] = factory()
	}

	return &deserializeReader{
		available:                  available,
		unsupportedMediaTypeResult: unsupportedMediaTypeResult,
		resultFactory:              resultFactory,
	}
}

func (this *deserializeReader) Read(input InputModel, request *http.Request) interface{} {
	if deserializer := this.loadDeserializer(request.Header[headerContentType]); deserializer == nil {
		return this.unsupportedMediaTypeResult
	} else if err := deserializer.Deserialize(input, request.Body); err != nil {
		return this.resultFactory(err)
	}

	return nil
}
func (this *deserializeReader) loadDeserializer(contentTypes []string) Deserializer {
	for _, contentType := range contentTypes {
		if deserializer, contains := this.available[normalizeMediaType(contentType)]; contains {
			return deserializer
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type bindReader struct {
	resultFactory func(error) interface{}
}

func newBindReader(resultFactory func(error) interface{}) Reader {
	return &bindReader{resultFactory: resultFactory}
}

func (this *bindReader) Read(target InputModel, request *http.Request) interface{} {
	if err := target.Bind(request); err != nil {
		return this.resultFactory(err)
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type validateReader struct {
	resultFactory func([]error) interface{}
	buffer        []error
}

func newValidateReader(resultFactory func([]error) interface{}, bufferSize int) Reader {
	return &validateReader{resultFactory: resultFactory, buffer: make([]error, bufferSize)}
}

func (this *validateReader) Read(target InputModel, _ *http.Request) interface{} {
	if count := target.Validate(this.buffer); count > 0 {
		return this.resultFactory(this.buffer[0:count])
	}

	return nil
}