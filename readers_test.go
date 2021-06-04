package shuttle

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAcceptReader_AcceptTypeProvided_NotFound_ReturnFailure(t *testing.T) {
	assertAcceptReader(t, "fail", []string{"not-found"}, []string{"not-found"})
}
func TestAcceptReader_NoAcceptTypesProvided_NoError(t *testing.T) {
	assertAcceptReader(t, "", nil, nil)
}
func TestAcceptReader_MultipleAcceptTypesProvided_Found_OverwriteAccept(t *testing.T) {
	assertAcceptReader(t, "", []string{"not-found-1, found"}, []string{"found"})
}
func TestAcceptReader_MultipleComplexAcceptTypesProvided_Found_OverwriteAccept(t *testing.T) {
	assertAcceptReader(t, "", []string{"not-found-1, found;q=0, not-found-2"}, []string{"found"})
}
func TestAcceptReader_WildcardAcceptTypeProvided_Found_OverwriteAccept(t *testing.T) {
	assertAcceptReader(t, "", []string{"*/*"}, nil)
}
func assertAcceptReader(t *testing.T, expectedResult string, acceptTypes, acceptTypesWhenSuccessful []string) {
	request := httptest.NewRequest("GET", "/", nil)
	request.Header["Accept"] = acceptTypes
	notAcceptable := &TextResult{Content: expectedResult}
	serializers := map[string]func() Serializer{
		"found": func() Serializer { return nil },
	}

	result := newAcceptReader(serializers, notAcceptable).Read(nil, request)

	if len(expectedResult) == 0 {
		Assert(t).That(result).IsNil()
	} else {
		Assert(t).That(result).Equals(notAcceptable)
	}

	if result == nil {
		Assert(t).That(request.Header["Accept"]).Equals(acceptTypesWhenSuccessful)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TestDeserializeReader_NoContentType_ReturnFailure(t *testing.T) {
	assertDeserializeReader(t, "unsupported-media-type", nil, nil)
}
func TestDeserializeReader_UnknownContentType_ReturnFailure(t *testing.T) {
	assertDeserializeReader(t, "unsupported-media-type", []string{"application/garbage"}, nil)
}
func TestDeserializeReader_DeserializationFailure_ReturnFailure(t *testing.T) {
	assertDeserializeReader(t, "unsupported-media-type", []string{"application/json"}, errors.New("fail"))
}
func TestDeserializeReader_KnownContentType_Success(t *testing.T) {
	assertDeserializeReader(t, nil, []string{"application/json"}, nil)
}
func TestDeserializeReader_KnownAdvancedContentType_Success(t *testing.T) {
	assertDeserializeReader(t, nil, []string{"application/json; charset=utf-8"}, nil)
}
func assertDeserializeReader(t *testing.T, expectedResult interface{}, contentTypes []string, deserializeError error) {
	input := &FakeInputModel{}
	request := httptest.NewRequest("GET", "/", nil)
	request.Header["Content-Type"] = contentTypes

	var errToCallback error
	callback := func(err error) interface{} {
		errToCallback = err
		return expectedResult
	}
	deserializer := &FakeDeserializer{err: deserializeError}
	factories := map[string]func() Deserializer{
		"application/json": func() Deserializer { return deserializer },
	}

	reader := newDeserializeReader(factories, "unsupported-media-type", callback)
	result := reader.Read(input, request)

	if result != "unsupported-media-type" {
		Assert(t).That(request.Body).Equals(deserializer.source)
		Assert(t).That(input).Equals(deserializer.target)
	}

	Assert(t).That(result).Equals(expectedResult)
	Assert(t).That(errToCallback).Equals(deserializeError)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TestBindReader_NoErrors(t *testing.T) {
	input := &FakeInputModel{}
	request := httptest.NewRequest("GET", "/", nil)

	result := newBindReader(nil).Read(input, request)

	Assert(t).That(result).IsNil()
	Assert(t).That(input.boundRequest == request).IsTrue()
}
func TestBindReader_Error(t *testing.T) {
	var bindError error
	input := &FakeInputModel{bindError: errors.New("bind error")}
	request := httptest.NewRequest("GET", "/", nil)

	result := newBindReader(func(err error) interface{} {
		bindError = err
		return "result"
	}).Read(input, request)

	Assert(t).That(result).Equals("result")
	Assert(t).That(bindError).Equals(input.bindError)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TestValidateReader_NoErrors(t *testing.T) {
	input := &FakeInputModel{}

	result := newValidateReader(nil, 4).Read(input, nil)

	Assert(t).That(result).IsNil()
}
func TestValidateReader_ErrorResult(t *testing.T) {
	input := &FakeInputModel{
		validationErrors: []error{errors.New("1"), errors.New("2")},
	}
	var errorsProvidedToFactor []error
	resultFactory := func(errs []error) interface{} { errorsProvidedToFactor = errs; return "fail" }

	result := newValidateReader(resultFactory, 4).Read(input, nil)

	Assert(t).That(result).Equals("fail")
	Assert(t).That(errorsProvidedToFactor).Equals(input.validationErrors)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type FakeInputModel struct {
	boundRequest     *http.Request
	bindError        error
	validationErrors []error
}

func (this *FakeInputModel) Reset() {}
func (this *FakeInputModel) Bind(request *http.Request) error {
	this.boundRequest = request
	return this.bindError
}
func (this *FakeInputModel) Validate(errs []error) int {
	for i := range this.validationErrors {
		errs[i] = this.validationErrors[i]
	}

	return len(this.validationErrors)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type FakeDeserializer struct {
	source io.Reader
	target interface{}
	err    error
}

func (this *FakeDeserializer) Deserialize(target interface{}, source io.Reader) error {
	this.target = target
	this.source = source
	return this.err
}
