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
		newFakeReader(t, 0, nil, request),
		newFakeReader(t, 1, nil, request),
		newFakeReader(t, 2, "fail", request),
		newFakeReader(t, 10, nil, request), // should never be called
	}
	writer := newFakeCaptureWriter(t, response, request)
	input := newFakeSequentialInputModel()
	handler := newTransientHandler(input, readers, nil, writer, &nopMonitor{})

	handler.ServeHTTP(response, request)

	Assert(t).That(writer.result).Equals("fail")
}
func TestHandler_RenderProcessorResultToResponse(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)
	readers := []Reader{newFakeReader(t, 0, nil, request)}
	writer := newFakeCaptureWriter(t, response, request)
	input := newFakeSequentialInputModel()
	processor := newFakeProcessor(t, request.Context(), input, "success")
	handler := newTransientHandler(input, readers, processor, writer, &nopMonitor{})

	handler.ServeHTTP(response, request)

	Assert(t).That(writer.result).Equals("success")
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type FakeReader struct {
	t               *testing.T
	expectedRequest *http.Request
	callSequence    int
	result          any
}

func newFakeReader(t *testing.T, expectedSequence int, result any, request *http.Request) Reader {
	return &FakeReader{t: t, callSequence: expectedSequence, result: result, expectedRequest: request}
}

func (this *FakeReader) Read(input InputModel, request *http.Request) any {
	sequential := input.(*FakeSequentialInputModel)
	Assert(this.t).That(this.expectedRequest).Equals(request)
	Assert(this.t).That(this.callSequence).Equals(sequential.ID)
	sequential.ID++
	return this.result
}

type FakeSequentialInputModel struct{ ID int }

func newFakeSequentialInputModel() *FakeSequentialInputModel {
	return &FakeSequentialInputModel{ID: 42}
}                                                               // garbage init
func (this *FakeSequentialInputModel) Reset()                   { this.ID = 0 }
func (this *FakeSequentialInputModel) Bind(*http.Request) error { return nil }
func (this *FakeSequentialInputModel) Validate([]error) int     { return 0 }

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type FakeProcessor struct {
	t             *testing.T
	expectedCtx   context.Context
	expectedInput any
	result        any
}

func newFakeProcessor(t *testing.T, expectedCtx context.Context, expectedInput, result any) *FakeProcessor {
	return &FakeProcessor{t: t, expectedCtx: expectedCtx, expectedInput: expectedInput, result: result}
}
func (this *FakeProcessor) Process(ctx context.Context, input any) any {
	Assert(this.t).That(ctx).Equals(this.expectedCtx)
	Assert(this.t).That(input).Equals(this.expectedInput)
	return this.result
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type FakeCaptureWriter struct {
	t        *testing.T
	response http.ResponseWriter
	request  *http.Request
	result   any
}

func newFakeCaptureWriter(t *testing.T, response http.ResponseWriter, request *http.Request) *FakeCaptureWriter {
	return &FakeCaptureWriter{t: t, response: response, request: request}
}

func (this *FakeCaptureWriter) Write(response http.ResponseWriter, request *http.Request, result any) {
	Assert(this.t).That(response).Equals(this.response)
	Assert(this.t).That(request).Equals(this.request)
	this.result = result
}
