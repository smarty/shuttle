package shuttle

import "testing"

func TestReadPathElement(t *testing.T) {
	assertPathElement(t, "/path/users/value", "users", "value")
	assertPathElement(t, "/path/users/value/", "users", "value")
	assertPathElement(t, "/path/users/value/other/stuff", "users", "value")
}
func assertPathElement(t *testing.T, raw, element, expected string) {
	Assert(t).That(ReadPathElement(raw, element)).Equals(expected)
}

func TestReadNumericPathElement(t *testing.T) {
	assertNumericPathElement(t, "/path/users/123", "users", 123)
	assertNumericPathElement(t, "/path/users/123/", "users", 123)
	assertNumericPathElement(t, "/path/users/123/other/stuff", "users", 123)
	assertNumericPathElement(t, "/path/users/-123/other/stuff", "users", 0)
	assertNumericPathElement(t, "/path/users/abc/other/stuff", "users", 0)
	assertNumericPathElement(t, "/path/users/abc123/other/stuff", "users", 0)
	assertNumericPathElement(t, "/path/users/_123/other/stuff", "users", 0)
}
func assertNumericPathElement(t *testing.T, raw, element string, expected uint64) {
	Assert(t).That(ReadNumericPathElement(raw, element)).Equals(expected)
}
