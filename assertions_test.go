package shuttle

import "reflect"

func Assert(t testingT) *That {
	return &That{t: t}
}

// That is an intermediate type, not to be instantiated directly
type That struct{ t testingT }

// That is an intermediate method call, as in: assert.With(t).That(actual).Equals(expected)
func (this *That) That(actual any) *Assertion {
	return &Assertion{
		testingT: this.t,
		actual:   actual,
	}
}

type testingT interface {
	Helper()
	Errorf(format string, args ...any)
}

// Assertion is an intermediate type, not to be instantiated directly.
type Assertion struct {
	testingT
	actual any
}

// IsNil asserts that the value provided to That is nil.
func (this *Assertion) IsNil() {
	this.Helper()
	if this.actual != nil && !reflect.ValueOf(this.actual).IsNil() {
		this.Equals(nil)
	}
}

// IsTrue asserts that the value provided to That is true.
func (this *Assertion) IsTrue() {
	this.Helper()
	this.Equals(true)
}

// IsFalse asserts that the value provided to That is false.
func (this *Assertion) IsFalse() {
	this.Helper()
	this.Equals(false)
}

// Equals asserts that the value provided is equal to the expected value.
func (this *Assertion) Equals(expected any) {
	this.Helper()

	if !reflect.DeepEqual(this.actual, expected) {
		this.Errorf("\n"+
			"Expected: %#v\n"+
			"Actual:   %#v",
			expected,
			this.actual,
		)
	}
}
