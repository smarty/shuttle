package shuttle

import (
	"net/http"
	"strings"
)

type acceptReader struct {
	acceptable                 map[string][]string
	result                     interface{}
	useDefaultIfAcceptNotFound bool
	maxAcceptTypes             int
	monitor                    Monitor
}

func newAcceptReader(serializerFactories map[string]func() Serializer, result *TextResult, useDefaultIfAcceptNotFound bool, maxAcceptTypes int, monitor Monitor) Reader {
	acceptable := make(map[string][]string)
	for acceptType := range serializerFactories {
		acceptable[acceptType] = []string{acceptType}
	}

	return &acceptReader{
		acceptable:                 acceptable,
		result:                     result,
		useDefaultIfAcceptNotFound: useDefaultIfAcceptNotFound,
		maxAcceptTypes:             maxAcceptTypes,
		monitor:                    monitor,
	}
}

func (this *acceptReader) Read(_ InputModel, request *http.Request) interface{} {
	if normalized, found := this.findAcceptType(request.Header[headerAccept]); !found {
		this.monitor.NotAcceptable()
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
		acceptTypesIndex := 0
		for {
			index := strings.Index(value, ",")
			if index >= 0 {
				item = value[0:index]
				value = value[index+1:]
			} else {
				item = value
			}

			item = normalizeMediaType(item)
			if item == headerAcceptAnyValue {
				return nil, true // default
			} else if types, contains := this.acceptable[item]; contains {
				return types, true
			} else if this.maxAcceptTypes > -1 && this.maxAcceptTypes <= acceptTypesIndex+1 {
				return nil, true
			} else if index == -1 {
				break
			}
			acceptTypesIndex++
		}
	}

	if this.useDefaultIfAcceptNotFound {
		return nil, true
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
	result                     ResultContainer
	monitor                    Monitor
}

func newDeserializeReader(deserializerFactories map[string]func() Deserializer, unsupportedMediaTypeResult interface{}, result ResultContainer, monitor Monitor) Reader {
	available := make(map[string]Deserializer, len(deserializerFactories))
	for contentType, factory := range deserializerFactories {
		available[contentType] = factory()
	}

	return &deserializeReader{
		available:                  available,
		unsupportedMediaTypeResult: unsupportedMediaTypeResult,
		result:                     result,
		monitor:                    monitor,
	}
}

func (this *deserializeReader) Read(input InputModel, request *http.Request) interface{} {
	this.monitor.Deserialize()

	var target interface{} = input
	if value, ok := input.(DeserializeBody); ok {
		target = value.Body()
	}

	if deserializer := this.loadDeserializer(request.Header[headerContentType]); deserializer == nil {
		this.monitor.UnsupportedMediaType()
		return this.unsupportedMediaTypeResult
	} else if err := deserializer.Deserialize(target, request.Body); err != nil {
		this.monitor.DeserializeFailed()
		this.result.SetContent(err) // implementations of this may override and no-op SetContent
		return this.result.Result()
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

type parseFormReader struct {
	result  interface{}
	monitor Monitor
}

func newParseFormReader(result interface{}, monitor Monitor) Reader {
	return &parseFormReader{
		result:  result,
		monitor: monitor,
	}
}

func (this *parseFormReader) Read(_ InputModel, request *http.Request) interface{} {
	this.monitor.ParseForm()
	if err := request.ParseForm(); err != nil {
		this.monitor.ParseFormFailed(err)
		return this.result
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type bindReader struct {
	result  ResultContainer
	monitor Monitor
}

func newBindReader(result ResultContainer, monitor Monitor) Reader {
	return &bindReader{result: result, monitor: monitor}
}

func (this *bindReader) Read(target InputModel, request *http.Request) interface{} {
	this.monitor.Bind()
	if err := target.Bind(request); err != nil {
		this.monitor.BindFailed(err)
		this.result.SetContent(err)
		return this.result.Result()
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type validateReader struct {
	result  ResultContainer
	buffer  []error
	monitor Monitor
}

func newValidateReader(result ResultContainer, bufferSize int, monitor Monitor) Reader {
	return &validateReader{result: result, buffer: make([]error, bufferSize), monitor: monitor}
}

func (this *validateReader) Read(target InputModel, _ *http.Request) interface{} {
	this.monitor.Validate()
	if count := target.Validate(this.buffer); count > 0 {
		errs := this.buffer[0:count]
		this.monitor.ValidateFailed(errs)
		this.result.SetContent(errs)
		return this.result.Result()
	}

	return nil
}
