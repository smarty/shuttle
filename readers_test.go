package shuttle

import (
	"errors"
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

func TestDeserializeReader(t *testing.T) {
	// TODO
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TestBindReader_NoErrors(t *testing.T) {
	input := &TestInputModel{}
	request := httptest.NewRequest("GET", "/", nil)

	result := newBindReader(nil).Read(input, request)

	Assert(t).That(result).IsNil()
	Assert(t).That(input.boundRequest == request).IsTrue()
}
func TestBindReader_Error(t *testing.T) {
	var bindError error
	input := &TestInputModel{bindError: errors.New("bind error")}
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
	input := &TestInputModel{}

	result := newValidationReader(4).Read(input, nil)

	Assert(t).That(result).IsNil()
}
func TestValidateReader_ErrorResult(t *testing.T) {
	input := &TestInputModel{
		validationErrors: []error{errors.New("1"), errors.New("2")},
	}

	result := newValidationReader(4).Read(input, nil)

	Assert(t).That(result).Equals(input.validationErrors)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type TestInputModel struct {
	boundRequest     *http.Request
	bindError        error
	validationErrors []error
}

func (this *TestInputModel) Reset() {}
func (this *TestInputModel) Bind(request *http.Request) error {
	this.boundRequest = request
	return this.bindError
}
func (this *TestInputModel) Validate(errs []error) int {
	for i := range this.validationErrors {
		errs[i] = this.validationErrors[i]
	}

	return len(this.validationErrors)
}
