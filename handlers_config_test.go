package shuttle

import (
	"bytes"
	"errors"
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
		// Options.DefaultSerializer(nil),
		// Options.DefaultSerializer(func() Serializer { return newJSONSerializer() }),
	)

	handler.ServeHTTP(response, request)

	Assert(t).That(response.Code).Equals(400)
}

func TestShuttleBindFailure(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", bytes.NewBufferString(`{"name":"value"}`))
	request.Header["Content-Type"] = []string{"application/json"}
	handler := NewHandler(
		Options.InputModel(func() InputModel { return &FakeDeserializeInputModel{Name: "garbage", bindFailure: errors.New("")} }),
		Options.DeserializeJSON(true),
	)

	handler.ServeHTTP(response, request)

	Assert(t).That(response.Code).Equals(400)
}

func TestShuttleValidationFailure(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", bytes.NewBufferString(`{"name":"value"}`))
	request.Header["Content-Type"] = []string{"application/json"}
	handler := NewHandler(
		Options.InputModel(func() InputModel { return &FakeDeserializeInputModel{Name: "garbage", validationFailure: 1} }),
		Options.DeserializeJSON(true),
	)

	handler.ServeHTTP(response, request)

	Assert(t).That(response.Code).Equals(422)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type FakeDeserializeInputModel struct {
	Name              string `json:"name"`
	bindFailure       error
	validationFailure int
}

func (this *FakeDeserializeInputModel) Reset()                   { this.Name = "" }
func (this *FakeDeserializeInputModel) Bind(*http.Request) error { return this.bindFailure }
func (this *FakeDeserializeInputModel) Validate([]error) int     { return this.validationFailure }
