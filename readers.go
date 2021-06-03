package shuttle

import "net/http"

// acceptReader

// deserializeReader

// bindReader

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
