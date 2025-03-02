package shuttle

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAcceptReader_AcceptTypeProvided_NotFound_ReturnFailure(t *testing.T) {
	assertAcceptReader(t, "fail", []string{"not-found"}, []string{"not-found"}, false, -1)
}
func TestAcceptReader_NoAcceptTypesProvided_NoError(t *testing.T) {
	assertAcceptReader(t, "", nil, nil, false, -1)
}
func TestAcceptReader_MultipleAcceptTypesProvided_Found_OverwriteAccept(t *testing.T) {
	assertAcceptReader(t, "", []string{"not-found-1, found"}, []string{"found"}, false, -1)
}
func TestAcceptReader_MultipleComplexAcceptTypesProvided_Found_OverwriteAccept(t *testing.T) {
	assertAcceptReader(t, "", []string{"not-found-1, found;q=0, not-found-2"}, []string{"found"}, false, -1)
}
func TestAcceptReader_WildcardAcceptTypeProvided_Found_OverwriteAccept(t *testing.T) {
	assertAcceptReader(t, "", []string{"*/*"}, nil, false, -1)
}
func TestAcceptReader_RealWorldExampleWithWildcard_Found(t *testing.T) {
	assertAcceptReader(t, "", []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"}, nil, false, -1)
}
func TestAcceptReader_UseDefaultIfNotFound_True(t *testing.T) {
	assertAcceptReader(t, "", []string{"text/html,application/xhtml+xml,application/xml;q=0.9,"}, nil, true, -1)
}
func TestAcceptReader_UseDefaultIfNotFound_False(t *testing.T) {
	assertAcceptReader(t, "fail", []string{"text/html,application/xhtml+xml,application/xml;q=0.9,"}, nil, false, -1)
}
func TestAcceptReader_UseDefaultIfNotFound_MaxAcceptTypes(t *testing.T) {
	assertAcceptReader(t, "", []string{"text/html,application/xhtml+xml,application/xml;q=0.9,"}, nil, true, 1)
}
func TestAcceptReader_MaxAcceptTypes_Single(t *testing.T) {
	assertAcceptReader(t, "", []string{"text/html,application/xhtml+xml,application/xml;q=0.9,"}, nil, false, 1)
}
func TestAcceptReader_MaxAcceptTypes_Limit_NotFound(t *testing.T) {
	assertAcceptReader(t, "", []string{"text/html,application/xhtml+xml,found,application/xml;q=0.9,"}, nil, false, 2)
}
func TestAcceptReader_MaxAcceptTypes_Limit_Found(t *testing.T) {
	assertAcceptReader(t, "", []string{"text/html,application/xhtml+xml,found,application/xml;q=0.9,"}, []string{"found"}, false, 3)
}
func TestAcceptReader_MaxAcceptTypes_All(t *testing.T) {
	assertAcceptReader(t, "fail", []string{"text/html,application/xhtml+xml,application/xml;q=0.9,"}, nil, false, -1)
}
func TestAcceptReader_MaxAcceptTypes_All_Found(t *testing.T) {
	assertAcceptReader(t, "", []string{"text/html,application/xhtml+xml,application/xml;q=0.9,found"}, []string{"found"}, false, -1)
}
func assertAcceptReader(t *testing.T, expectedResult string, acceptTypes, acceptTypesWhenSuccessful []string, useDefaultIfNotFound bool, maxTypes int) {
	request := httptest.NewRequest("GET", "/", nil)
	request.Header["Accept"] = acceptTypes
	notAcceptable := &TextResult{Content: expectedResult}
	serializers := map[string]func() Serializer{
		"found": func() Serializer { return nil },
	}

	result := newAcceptReader(serializers, notAcceptable, useDefaultIfNotFound, maxTypes, &nopMonitor{}).Read(nil, request)

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
func TestDeserializeReader_XMLDeserializationFailure_ReturnFailure(t *testing.T) {
	assertDeserializeReader(t, "unsupported-media-type", []string{"application/xml"}, errors.New("fail"))
}
func TestDeserializeReader_KnownContentType_Success(t *testing.T) {
	assertDeserializeReader(t, nil, []string{"application/json"}, nil)
}
func TestDeserializeReader_XMLKnownContentType_Success(t *testing.T) {
	assertDeserializeReader(t, nil, []string{"application/xml"}, nil)
}
func TestDeserializeReader_KnownAdvancedContentType_Success(t *testing.T) {
	assertDeserializeReader(t, nil, []string{"application/json; charset=utf-8"}, nil)
}
func TestDeserializeReader_XMLKnownAdvancedContentType_Success(t *testing.T) {
	assertDeserializeReader(t, nil, []string{"application/xml; charset=utf-8"}, nil)
}
func assertDeserializeReader(t *testing.T, expectedResult any, contentTypes []string, deserializeError error) {
	input := &FakeInputModel{}
	request := httptest.NewRequest("GET", "/", nil)
	request.Header["Content-Type"] = contentTypes

	fakeResult := &FakeContentResult{}
	deserializer := &FakeDeserializer{err: deserializeError}
	factories := map[string]func() Deserializer{
		"application/json": func() Deserializer { return deserializer },
		"application/xml":  func() Deserializer { return deserializer },
	}

	reader := newDeserializeReader(factories, "unsupported-media-type", fakeResult, &nopMonitor{})
	result := reader.Read(input, request)

	if result != "unsupported-media-type" {
		Assert(t).That(request.Body).Equals(deserializer.source)
		Assert(t).That(input).Equals(deserializer.target)
	}

	// Assert(t).That(result).Equals(expectedResult) // TODO
	Assert(t).That(fakeResult.value).Equals(deserializeError)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TestBindReader_NoErrors(t *testing.T) {
	input := &FakeInputModel{}
	request := httptest.NewRequest("GET", "/", nil)

	result := newBindReader(nil, &nopMonitor{}).Read(input, request)

	Assert(t).That(result).IsNil()
	Assert(t).That(input.boundRequest == request).IsTrue()
}
func TestBindReader_Error(t *testing.T) {
	input := &FakeInputModel{bindError: errors.New("bind error")}
	request := httptest.NewRequest("GET", "/", nil)
	fakeBindErrorResult := &FakeContentResult{}

	result := newBindReader(fakeBindErrorResult, &nopMonitor{}).Read(input, request)

	Assert(t).That(result).Equals(fakeBindErrorResult)
	Assert(t).That(fakeBindErrorResult.value).Equals(input.bindError)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TestValidateReader_NoErrors(t *testing.T) {
	input := &FakeInputModel{}

	result := newValidateReader(nil, 4, &nopMonitor{}).Read(input, nil)

	Assert(t).That(result).IsNil()
}
func TestValidateReader_ErrorResult(t *testing.T) {
	input := &FakeInputModel{
		validationErrors: []error{errors.New("1"), errors.New("2")},
	}
	fakeValidationErrorsResult := &FakeContentResult{}

	result := newValidateReader(fakeValidationErrorsResult, 4, &nopMonitor{}).Read(input, nil)

	Assert(t).That(result).Equals(fakeValidationErrorsResult)
	Assert(t).That(fakeValidationErrorsResult.value).Equals(input.validationErrors)
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
	copy(errs, this.validationErrors)
	return len(this.validationErrors)
}

func (this *FakeInputModel) Body() any { return this }

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type FakeDeserializer struct {
	source io.Reader
	target any
	err    error
}

func (this *FakeDeserializer) Deserialize(target any, source io.Reader) error {
	this.target = target
	this.source = source
	return this.err
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type FakeContentResult struct{ value any }

func (this *FakeContentResult) SetContent(value any) { this.value = value }
func (this *FakeContentResult) Result() any          { return this }
