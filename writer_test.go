package shuttle

import (
	"net/http/httptest"
	"testing"
)

func TestWrite(t *testing.T) {
	assertions := []struct {
		Input interface{}
		HTTPResponse
	}{
		{Input: nil, HTTPResponse: HTTPResponse{
			StatusCode: 200, ContentType: nil, Body: ""}},
		{Input: "body", HTTPResponse: HTTPResponse{
			StatusCode: 200, ContentType: []string{"text/plain; charset=utf-8"}, Body: "body"}},
		{Input: []byte("body"), HTTPResponse: HTTPResponse{
			StatusCode: 200, ContentType: []string{"text/plain; charset=utf-8"}, Body: "body"}},
		{Input: true, HTTPResponse: HTTPResponse{
			StatusCode: 200, ContentType: []string{"text/plain; charset=utf-8"}, Body: "true"}},
		{Input: false, HTTPResponse: HTTPResponse{
			StatusCode: 200, ContentType: []string{"text/plain; charset=utf-8"}, Body: "false"}},
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

type HTTPResponse struct {
	StatusCode  int
	ContentType []string
	Body        string
}
