package shuttle

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWrite(t *testing.T) {
	assertions := []struct {
		Input  interface{}
		Accept string
		HTTPResponse
	}{
		{Input: nil,
			HTTPResponse: HTTPResponse{StatusCode: 204, ContentType: nil, Body: ""}},
		{Input: "body",
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: []string{"text/plain; charset=utf-8"}, Body: "body"}},
		{Input: []byte("body"),
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: []string{"text/plain; charset=utf-8"}, Body: "body"}},

		{Input: true,
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: []string{"text/plain; charset=utf-8"}, Body: "true"}},
		{Input: false,
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: []string{"text/plain; charset=utf-8"}, Body: "false"}},

		{Input: &TextResult{StatusCode: 201, ContentType: "no-expected-content-type-for-empty-body", Content: ""},
			HTTPResponse: HTTPResponse{StatusCode: 201, ContentType: nil, Body: ""}},
		{Input: TextResult{StatusCode: 201, ContentType: "no-expected-content-type-for-empty-body", Content: ""},
			HTTPResponse: HTTPResponse{StatusCode: 201, ContentType: nil, Body: ""}},
		{Input: &TextResult{StatusCode: 0, ContentType: "", Content: "body"},
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: []string{"text/plain; charset=utf-8"}, Body: "body"}},
		{Input: &TextResult{StatusCode: 201, ContentType: "application/custom", Content: "body"},
			HTTPResponse: HTTPResponse{StatusCode: 201, ContentType: []string{"application/custom"}, Body: "body"}},
		{Input: TextResult{StatusCode: 0, ContentType: "application/custom", Content: "body"},
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: []string{"application/custom"}, Body: "body"}},
		{Input: TextResult{StatusCode: 202, ContentType: "", Content: "body"},
			HTTPResponse: HTTPResponse{StatusCode: 202, ContentType: nil, Body: "body"}},

		{Input: &BinaryResult{StatusCode: 404, ContentType: "no-expected-content-type-for-empty-body", Content: nil},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: nil, Body: ""}},
		{Input: BinaryResult{StatusCode: 404, ContentType: "no-expected-content-type-for-empty-body", ContentDisposition: "no-expected-content-disposition-for-empty-body", Content: nil},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: nil, Body: ""}},
		{Input: &BinaryResult{StatusCode: 404, ContentType: "custom-type", Content: []byte("body")},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: []string{"custom-type"}, Body: "body"}},
		{Input: BinaryResult{StatusCode: 404, ContentType: "", ContentDisposition: "custom-disposition", Content: []byte("body")},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: nil, ContentDisposition: []string{"custom-disposition"}, Body: "body"}},
		{Input: BinaryResult{StatusCode: 404, ContentType: "", Content: []byte("body")},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: nil, Body: "body"}},

		{Input: &StreamResult{StatusCode: 404, ContentType: "no-expected-content-type-for-empty-body", Content: nil},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: nil, Body: ""}},
		{Input: StreamResult{StatusCode: 404, ContentType: "no-expected-content-type-for-empty-body", Content: nil},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: nil, Body: ""}},
		{Input: StreamResult{StatusCode: 404, ContentType: "no-expected-content-type-for-empty-body", ContentDisposition: "no-expected-content-disposition-for-empty-body", Content: nil},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: nil, ContentDisposition: nil, Body: ""}},
		{Input: &StreamResult{StatusCode: 404, ContentType: "", Content: bytes.NewBufferString("body")},
			HTTPResponse: HTTPResponse{StatusCode: 404, ContentType: nil, Body: "body"}},
		{Input: StreamResult{StatusCode: 422, ContentType: "application/custom", Content: bytes.NewBufferString("body")},
			HTTPResponse: HTTPResponse{StatusCode: 422, ContentType: []string{"application/custom"}, Body: "body"}},
		{Input: StreamResult{StatusCode: 422, ContentType: "application/custom", ContentDisposition: "custom-disposition", Content: bytes.NewBufferString("body")},
			HTTPResponse: HTTPResponse{StatusCode: 422, ContentType: []string{"application/custom"}, ContentDisposition: []string{"custom-disposition"}, Body: "body"}},

		{Input: &SerializeResult{StatusCode: 401, ContentType: "no-expected-content-type-for-empty-body", Content: nil},
			HTTPResponse: HTTPResponse{StatusCode: 401, ContentType: nil, Body: ""}},
		{Input: SerializeResult{StatusCode: 401, ContentType: "no-expected-content-type-for-empty-body", Content: nil},
			HTTPResponse: HTTPResponse{StatusCode: 401, ContentType: nil, Body: ""}},

		{Input: &SerializeResult{StatusCode: 422, ContentType: "", Content: "body"},
			Accept:       "", // default serializer
			HTTPResponse: HTTPResponse{StatusCode: 422, ContentType: []string{"application/json; charset=utf-8"}, Body: "{body}"}},
		{Input: &SerializeResult{StatusCode: 422, ContentType: "application/override-default", Content: "body"},
			Accept:       "", // default serializer
			HTTPResponse: HTTPResponse{StatusCode: 422, ContentType: []string{"application/override-default"}, Body: "{body}"}},
		{Input: &SerializeResult{StatusCode: 422, ContentType: "", Content: "body"},
			Accept:       "application/xml", // serializer matching this Accept value
			HTTPResponse: HTTPResponse{StatusCode: 422, ContentType: []string{"application/xml; charset=utf-8"}, Body: string(xmlPrefix) + "{body}"}},
		{Input: &SerializeResult{StatusCode: 422, ContentType: "application/override-default", Content: "body"},
			Accept:       "application/xml", // serializer matching this Accept value
			HTTPResponse: HTTPResponse{StatusCode: 422, ContentType: []string{"application/override-default"}, Body: string(xmlPrefix) + "{body}"}},
		{Input: &SerializeResult{StatusCode: 422, ContentType: "", Content: "body"},
			Accept:       "application/not-acceptable", // use default serializer
			HTTPResponse: HTTPResponse{StatusCode: 422, ContentType: []string{"application/json; charset=utf-8"}, Body: "{body}"}},

		{Input: &SerializeResult{StatusCode: 200, ContentType: "", Content: "body"},
			Accept:       "application/xml;q=0.8", // simplify and use correct serializer
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: []string{"application/xml; charset=utf-8"}, Body: string(xmlPrefix) + "{body}"}},

		{Input: &JSONPResult{StatusCode: 201, ContentType: "", Content: "body"},
			HTTPResponse: HTTPResponse{StatusCode: 201, ContentType: []string{"application/javascript; charset=utf-8"}, Body: "callback({body})"}},
		{Input: JSONPResult{StatusCode: 201, ContentType: "", Content: "body"},
			HTTPResponse: HTTPResponse{StatusCode: 201, ContentType: []string{"application/javascript; charset=utf-8"}, Body: "callback({body})"}},

		{Input: 42, // use serializer for unknown type
			HTTPResponse: HTTPResponse{StatusCode: 200, ContentType: []string{"application/json; charset=utf-8"}, Body: "{42}"}},
	}

	for _, assertion := range assertions {
		response := recordResponse(assertion.Input, assertion.Accept)
		assertResponse(t, response, assertion.HTTPResponse)
	}
}
func recordResponse(result interface{}, acceptHeader string) *httptest.ResponseRecorder {
	writer := newTestWriter()
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)
	if len(acceptHeader) > 0 {
		request.Header["Accept"] = []string{acceptHeader}
	}

	writer.Write(response, request, result)

	return response
}
func newTestWriter() Writer {
	return newWriter(map[string]func() Serializer{
		emptyContentType:  func() Serializer { return newFakeWriteSerializer("application/json; charset=utf-8") },
		"application/xml": func() Serializer { return newFakeWriteSerializer("application/xml; charset=utf-8") },
	}, &nopMonitor{})
}
func assertResponse(t *testing.T, response *httptest.ResponseRecorder, expected HTTPResponse) {
	Assert(t).That(response.Code).Equals(expected.StatusCode)
	Assert(t).That(response.Header()["Content-Type"]).Equals(expected.ContentType)
	Assert(t).That(response.Header()["Content-Disposition"]).Equals(expected.ContentDisposition)
	Assert(t).That(response.Body.String()).Equals(expected.Body)
}

func TestWriteHTTPHandler(t *testing.T) {
	handler := &FakeHTTPHandlerResult{}
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)

	newTestWriter().Write(response, request, handler)

	Assert(t).That(handler.response == response).IsTrue()
	Assert(t).That(handler.request == request).IsTrue()
}

func TestSerializeResultWriteHTTPHeaders(t *testing.T) {
	headers := map[string][]string{
		"Header-1": {"value1-a", "value1-b"},
		"Header-2": {"value2-a", "value2-b"},
	}

	assertHTTPHeaders(t, SerializeResult{Headers: headers})
	assertHTTPHeaders(t, TextResult{Headers: headers})
	assertHTTPHeaders(t, BinaryResult{Headers: headers})
	assertHTTPHeaders(t, StreamResult{Headers: headers})
	assertHTTPHeaders(t, JSONPResult{Headers: headers})
}
func assertHTTPHeaders(t *testing.T, result interface{}) {
	response := recordResponse(result, "application/json")
	headers := response.Header()
	Assert(t).That(headers["Header-1"]).Equals([]string{"value1-a", "value1-b"})
	Assert(t).That(headers["Header-2"]).Equals([]string{"value2-a", "value2-b"})
}

type HTTPResponse struct {
	StatusCode         int
	ContentType        []string
	ContentDisposition []string
	Body               string
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TestJSONPCallbackQueryStringParsing(t *testing.T) {
	assertJSONPQueryStringCallback(t, "callback=jsonCallback", "jsonCallback")                           // simple
	assertJSONPQueryStringCallback(t, "key=&callback=jsonCallback", "jsonCallback")                      // multiple keys
	assertJSONPQueryStringCallback(t, "key=value&callback=jsonCallback", "jsonCallback")                 // multiple keys and values
	assertJSONPQueryStringCallback(t, "key=&=value&callback=jsonCallback&other=stuff", "jsonCallback")   // blank keys and values
	assertJSONPQueryStringCallback(t, "callback=_json_Callback_0123456789", "_json_Callback_0123456789") // complex callback name
	assertJSONPQueryStringCallback(t, "key=&=value&other=stuff", "callback")                             // doesn't exist, use default
	assertJSONPQueryStringCallback(t, "callback=malicious!", "callback")                                 // malicious
	assertJSONPQueryStringCallback(t, "callback=<malicious>", "callback")                                // malicious
	assertJSONPQueryStringCallback(t, "callback=alert('malicious');", "callback")                        // malicious

}
func assertJSONPQueryStringCallback(t *testing.T, raw, callback string) {
	Assert(t).That(parseJSONPCallbackQueryStringParameter(raw)).Equals(callback)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type FakeHTTPHandlerResult struct {
	response http.ResponseWriter
	request  *http.Request
}

func (this *FakeHTTPHandlerResult) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	this.response = response
	this.request = request
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type FakeWriteSerializer string

func newFakeWriteSerializer(contentType string) Serializer {
	return FakeWriteSerializer(contentType)
}
func (this FakeWriteSerializer) ContentType() string { return string(this) }
func (this FakeWriteSerializer) Serialize(writer io.Writer, value interface{}) error {
	raw, _ := json.Marshal(value)
	_, _ = io.WriteString(writer, "{"+strings.ReplaceAll(string(raw), `"`, ``)+"}")
	return nil
}
