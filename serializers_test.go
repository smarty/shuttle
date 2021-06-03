package shuttle

import (
	"bytes"
	"testing"
)

func TestJSONDeserializer(t *testing.T) {
	var value string
	deserializer := newJSONDeserializer()
	err := deserializer.Deserialize(&value, bytes.NewBufferString(`"hello"`))

	Assert(t).That(err).IsNil()
	Assert(t).That(value).Equals("hello")
}
func TestJSONDeserializer_ReturnError(t *testing.T) {
	var value string
	deserializer := newJSONDeserializer()
	err := deserializer.Deserialize(&value, bytes.NewBufferString(`{`))

	Assert(t).That(err).Equals(ErrDeserializationFailure)
	Assert(t).That(value).Equals("")
}
