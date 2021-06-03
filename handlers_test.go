package shuttle

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_ReadFailure_RenderErrorToResponse(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)
	readers := []Reader{
		newTestReader(t, 0, nil, request),
		newTestReader(t, 1, nil, request),
		newTestReader(t, 2, "fail", request),
		newTestReader(t, 10, nil, request), // never called
	}
	writer := newTestCaptureWriter(t, response, request)
	input := newSequentialInputModel()
	handler := newHandler(input, readers, nil, writer)

	handler.ServeHTTP(response, request)

	Assert(t).That(writer.result).Equals("fail")
}
func TestHandler_RenderProcessorResultToResponse(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)
	readers := []Reader{newTestReader(t, 0, nil, request)}
	writer := newTestCaptureWriter(t, response, request)
	input := newSequentialInputModel()
	processor := newTestProcessor(t, request.Context(), input, "success")
	handler := newHandler(input, readers, processor, writer)

	handler.ServeHTTP(response, request)

	Assert(t).That(writer.result).Equals("success")
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type TestReader struct {
	t               *testing.T
	expectedRequest *http.Request
	callSequence    int
	result          interface{}
}

func newTestReader(t *testing.T, expectedSequence int, result interface{}, request *http.Request) Reader {
	return &TestReader{t: t, callSequence: expectedSequence, result: result, expectedRequest: request}
}

func (this *TestReader) Read(input InputModel, request *http.Request) interface{} {
	sequential := input.(*SequentialInputModel)
	Assert(this.t).That(this.expectedRequest).Equals(request)
	Assert(this.t).That(this.callSequence).Equals(sequential.ID)
	sequential.ID++
	return this.result
}

type SequentialInputModel struct{ ID int }

func newSequentialInputModel() *SequentialInputModel        { return &SequentialInputModel{ID: 42} } // garbage init
func (this *SequentialInputModel) Reset()                   { this.ID = 0 }
func (this *SequentialInputModel) Bind(*http.Request) error { return nil }
func (this *SequentialInputModel) Validate([]error) int     { return 0 }

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type TestProcessor struct {
	t             *testing.T
	expectedCtx   context.Context
	expectedInput interface{}
	result        interface{}
}

func newTestProcessor(t *testing.T, expectedCtx context.Context, expectedInput, result interface{}) *TestProcessor {
	return &TestProcessor{t: t, expectedCtx: expectedCtx, expectedInput: expectedInput, result: result}
}
func (this *TestProcessor) Process(ctx context.Context, input interface{}) interface{} {
	Assert(this.t).That(ctx).Equals(this.expectedCtx)
	Assert(this.t).That(input).Equals(this.expectedInput)
	return this.result
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type TestWriter struct {
	t        *testing.T
	response http.ResponseWriter
	request  *http.Request
	result   interface{}
}

func newTestCaptureWriter(t *testing.T, response http.ResponseWriter, request *http.Request) *TestWriter {
	return &TestWriter{t: t, response: response, request: request}
}

func (this *TestWriter) Write(response http.ResponseWriter, request *http.Request, result interface{}) {
	Assert(this.t).That(response).Equals(this.response)
	Assert(this.t).That(request).Equals(this.request)
	this.result = result
}
