package inputs

import "github.com/smarty/shuttle"

type Subtraction struct {
	shuttle.BaseInputModel
	A int `json:"a"`
	B int `json:"b"`
}

func NewSubtraction() *Subtraction {
	return &Subtraction{
		A: -1,
		B: -1,
	}
}

func (this *Subtraction) Reset() {
	this.A = 0
	this.B = 0
}

func (this *Subtraction) Validate(errors []error) (count int) {
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
	if this.B > this.A {
		errors[count] = shuttle.InputError{
			Fields:  []string{"query:a", "query:b"},
			Name:    "validate:a>b",
			Message: "a must be greater than or equal to b",
		}
		count++
	}
	return count
}
