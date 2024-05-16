package inputs

import (
	"net/http"
	"strconv"

	"github.com/smarty/shuttle"
)

// Addition represents the data from the client's request.
type Addition struct {
	A int
	B int
}

// NewAddition is the constructor and, by convention, sets the data to garbage values.
func NewAddition() *Addition {
	return &Addition{
		A: -1,
		B: -1,
	}
}

// Reset is called by shuttle to prepare the instance for use with the current request.
// This instance will be re-used over the lifetime of the application.
func (this *Addition) Reset() {
	this.A = 0
	this.B = 0
}

// Bind is your only opportunity to get data from the request.
// Returning an error indicates that the request data is completely inscrutable.
// In such a case, processing will be short-circuited resulting in HTTP 400 Bad Request.
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

// Validate is an opportunity to ensure that the values gathered during Bind are usable.
// The errors slice provided is pre-initialized and can be directly assigned to, beginning at index 0.
// The presence of any errors short-circuits processing and results in HTTP 422 Unprocessable Entity.
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
