package shuttle

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWrite(t *testing.T) {
	assertions := []struct {
		Input interface{}
		HTTPResponse
	}{
		{Input: nil,
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: nil, Body: ""}},
		{Input: "body",
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: []string{"text/plain; charset=utf-8"}, Body: "body"}},
		{Input: []byte("body"),
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: []string{"text/plain; charset=utf-8"}, Body: "body"}},

		{Input: true,
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: []string{"text/plain; charset=utf-8"}, Body: "true"}},
		{Input: false,
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: []string{"text/plain; charset=utf-8"}, Body: "false"}},

		{Input: &TextResult{StatusCode: 0, ContentType: "", Content: "body"},
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: []string{"text/plain; charset=utf-8"}, Body: "body"}},
		{Input: &TextResult{StatusCode: 201, ContentType: "application/custom", Content: "body"},
			HTTPResponse: HTTPResponse{StatusCode: 201, ContentType: []string{"application/custom"}, Body: "body"}},
		{Input: &TextResult{StatusCode: 201, ContentType: "no-expected-content-type-for-empty-body", Content: ""},
			HTTPResponse: HTTPResponse{StatusCode: 201, ContentType: nil, Body: ""}},
		{Input: TextResult{StatusCode: 0, ContentType: "application/custom", Content: "body"},
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: []string{"application/custom"}, Body: "body"}},
		{Input: TextResult{StatusCode: 202, ContentType: "", Content: "body"},
			HTTPResponse: HTTPResponse{StatusCode: 202, ContentType: nil, Body: "body"}},
		{Input: TextResult{StatusCode: 201, ContentType: "no-expected-content-type-for-empty-body", Content: ""},
			HTTPResponse: HTTPResponse{StatusCode: 201, ContentType: nil, Body: ""}},

		{Input: &BinaryResult{StatusCode: 404, ContentType: "", Content: []byte("body")},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: nil, Body: "body"}},
		{Input: BinaryResult{StatusCode: 404, ContentType: "", Content: []byte("body")},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: nil, Body: "body"}},
		{Input: &BinaryResult{StatusCode: 404, ContentType: "no-expected-content-type-for-empty-body", Content: nil},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: nil, Body: ""}},
		{Input: BinaryResult{StatusCode: 404, ContentType: "no-expected-content-type-for-empty-body", Content: nil},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: nil, Body: ""}},

		{Input: &StreamResult{StatusCode: 404, ContentType: "", Content: bytes.NewBufferString("body")},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: nil, Body: "body"}},
		{Input: StreamResult{StatusCode: 422, ContentType: "application/custom", Content: bytes.NewBufferString("body")},
			HTTPResponse: HTTPResponse{StatusCode: 422, ContentType: []string{"application/custom"}, Body: "body"}},
		{Input: &StreamResult{StatusCode: 404, ContentType: "no-expected-content-type-for-empty-body", Content: nil},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: nil, Body: ""}},
		{Input: StreamResult{StatusCode: 404, ContentType: "no-expected-content-type-for-empty-body", Content: nil},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: nil, Body: ""}},
	}

	for _, assertion := range assertions {
		response := recordResponse(assertion.Input)
		assertResponse(t, response, assertion.HTTPResponse)
	}
}
func recordResponse(result interface{}) *httptest.ResponseRecorder {
	writer := newTestWriter(nil) // TODO: serializers
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil) // TODO: headers

	writer.Write(response, request, result)

	return response
}
func newTestWriter(serializers map[string]func() Serializer) Writer {
	return newWriter(serializers)
}
func assertResponse(t *testing.T, response *httptest.ResponseRecorder, expected HTTPResponse) {
	Assert(t).That(response.Code).Equals(expected.StatusCode)
	Assert(t).That(response.Header()["Content-Type"]).Equals(expected.ContentType)
	Assert(t).That(response.Body.String()).Equals(expected.Body)
}

func TestWriteHTTPHandler(t *testing.T) {
	handler := &TestHTTPHandlerResult{}
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)

	newTestWriter(nil).Write(response, request, handler)

	Assert(t).That(handler.response == response).IsTrue()
	Assert(t).That(handler.request == request).IsTrue()
}

type HTTPResponse struct {
	StatusCode  int
	ContentType []string
	Body        string
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type TestHTTPHandlerResult struct {
	response http.ResponseWriter
	request  *http.Request
}

func (this *TestHTTPHandlerResult) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	this.response = response
	this.request = request
}
