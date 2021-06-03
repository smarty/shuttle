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

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TestJSONSerializer(t *testing.T) {
	serializer := newJSONSerializer()
	buffer := bytes.NewBufferString("")

	err := serializer.Serialize(buffer, "hello")

	Assert(t).That(err).IsNil()
	Assert(t).That(buffer.String()).Equals(`"hello"` + "\n")
}
func TestJSONSerializer_Failure(t *testing.T) {
	serializer := newJSONSerializer()
	buffer := bytes.NewBufferString("")

	err := serializer.Serialize(buffer, make(chan string))

	Assert(t).That(err).Equals(ErrSerializationFailure)
	Assert(t).That(buffer.Len()).Equals(0)
}
