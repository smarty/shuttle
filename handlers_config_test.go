package shuttle

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestShuttleNopConfig(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)
	handler := NewHandler()

	handler.ServeHTTP(response, request)

	Assert(t).That(response.Code).Equals(200)
}
func TestShuttleDeserialize(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", bytes.NewBufferString(`{"name":"name-here"}`))
	request.Header["Content-Type"] = []string{"application/json"}
	handler := NewHandler(
		Options.InputModel(func() InputModel { return &FakeDeserializeInputModel{Name: "garbage"} }),
		Options.DeserializeJSON(true),
	)

	handler.ServeHTTP(response, request)

	Assert(t).That(response.Code).Equals(200)
}

func TestShuttleSkipSerialization(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", bytes.NewBufferString(`{"name":"name-here"}`))
	handler := NewHandler(
		Options.InputModel(func() InputModel { return &FakeDeserializeInputModel{Name: "garbage"} }),
		Options.DeserializeJSON(false),
		Options.SerializeJSON(false),
	)

	handler.ServeHTTP(response, request)

	Assert(t).That(response.Code).Equals(200)
}
func TestShuttleDeserializationFailure(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", bytes.NewBufferString(`{`))
	request.Header["Content-Type"] = []string{"application/json"}
	handler := NewHandler(
		Options.InputModel(func() InputModel { return &FakeDeserializeInputModel{Name: "garbage"} }),
		Options.DeserializeJSON(true),
	)

	handler.ServeHTTP(response, request)

	Assert(t).That(response.Code).Equals(400)
}

func TestShuttleBindFailure(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", bytes.NewBufferString(`{"name":"value"}`))
	request.Header["Content-Type"] = []string{"application/json"}
	bindError := InputError{Message: "bind-failure"}
	handler := NewHandler(
		Options.InputModel(func() InputModel { return &FakeDeserializeInputModel{Name: "garbage", bindFailure: bindError} }),
		Options.DeserializeJSON(true),
	)

	handler.ServeHTTP(response, request)

	Assert(t).That(response.Code).Equals(400)
	Assert(t).That(response.Body.String()).Equals(`{"errors":[{"message":"bind-failure"}]}` + "\n")
}

func TestShuttleValidationFailure(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)
	handler := NewHandler(
		Options.InputModel(func() InputModel { return &FakeDeserializeInputModel{Name: "garbage", validationFailure: 2} }),
	)

	handler.ServeHTTP(response, request)

	Assert(t).That(response.Code).Equals(422)
	Assert(t).That(response.Body.String()).Equals(`{"errors":[{"message":"validation-error"},{"message":"validation-error"}]}` + "\n")

}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TestInputError_Error(t *testing.T) {
	input := &InputError{Message: "hello"}
	Assert(t).That(input.Error()).Equals(input.Message)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type FakeDeserializeInputModel struct {
	Name              string `json:"name"`
	bindFailure       error
	validationFailure int
}

func (this *FakeDeserializeInputModel) Reset()                   { this.Name = "" }
func (this *FakeDeserializeInputModel) Bind(*http.Request) error { return this.bindFailure }
func (this *FakeDeserializeInputModel) Validate(target []error) int {
	for i := 0; i < this.validationFailure; i++ {
		target[i] = InputError{Message: "validation-error"}
	}

	return this.validationFailure
}
