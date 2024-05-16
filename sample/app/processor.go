package app

import (
	"context"
	"net/http"

	"github.com/smarty/shuttle"
	"github.com/smarty/shuttle/sample/inputs"
	"github.com/smarty/shuttle/sample/outputs"
)

type Processor struct{}

func NewProcessor() *Processor {
	return &Processor{}
}

func (this *Processor) Process(_ context.Context, v any) any {
	switch input := v.(type) {
	case *inputs.Addition:
		return outputs.Addition{
			A: input.A,
			B: input.B,
			C: input.A + input.B,
		}
	case *inputs.Subtraction:
		return outputs.Subtraction{
			A: input.A,
			B: input.B,
			C: input.A - input.B,
		}
	default:
		return shuttle.SerializeResult{
			StatusCode:  http.StatusInternalServerError,
			ContentType: "text/plain; charset=utf-8",
			Content:     http.StatusText(http.StatusInternalServerError),
		}
	}
}
