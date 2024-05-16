package inputs

import (
	"net/http"
	"strconv"

	"github.com/smarty/shuttle"
)

type Addition struct {
	A int
	B int
}

func NewAddition() *Addition {
	return &Addition{
		A: -1,
		B: -1,
	}
}

func (this *Addition) Reset() {
	this.A = 0
	this.B = 0
}

func (this *Addition) Bind(request *http.Request) error {
	rawA := request.URL.Query().Get("a")
	a, err := strconv.Atoi(rawA)
	if err != nil {
		return shuttle.InputError{
			Fields:  []string{"query:a"},
			Name:    "bind:integer",
			Message: "failed to convert parameter to integer",
		}
	}
	this.A = a
	rawB := request.URL.Query().Get("b")
	b, err := strconv.Atoi(rawB)
	if err != nil {
		return shuttle.InputError{
			Fields:  []string{"query:b"},
			Name:    "bind:integer",
			Message: "failed to convert parameter to integer",
		}
	}
	this.B = b
	return nil
}
func (this *Addition) Validate(errors []error) (count int) {
	if this.A <= 0 {
		errors[count] = shuttle.InputError{
			Fields:  []string{"query:a"},
			Name:    "validate:positive",
			Message: "parameter must be a positive integer",
		}
		count++
	}
	if this.B <= 0 {
		errors[count] = shuttle.InputError{
			Fields:  []string{"query:b"},
			Name:    "validate:positive",
			Message: "parameter must be a positive integer",
		}
		count++
	}
	return count
}
