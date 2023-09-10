package shuttle

import (
	"bytes"
	"errors"
	"testing"
)

func TestJSONDeserializer(t *testing.T) {
	var value string
	deserializer := newJSONDeserializer()
	err := deserializer.Deserialize(&value, bytes.NewBufferString(`"hello"`))

	Assert(t).That(err).IsNil()
	Assert(t).That(value).Equals("hello")
}
func TestXMLDeserializer(t *testing.T) {
	var value string
	deserializer := newXMLDeserializer()
	err := deserializer.Deserialize(&value, bytes.NewBufferString(`"<string>hello</string>"`))

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
func TestJSONDeserializer_CorruptedStream_NewDeserializer(t *testing.T) {
	var items []uint64
	deserializer := newJSONDeserializer()

	err := deserializer.Deserialize(&items, bytes.NewBufferString(`[1,2,3]I AM MALFORMED`))
	Assert(t).That(err).Equals(ErrDeserializationFailure)
	Assert(t).That(items).Equals([]uint64{1, 2, 3})

	err = deserializer.Deserialize(&items, bytes.NewBufferString(`[1,2,3]`))
	Assert(t).That(err).IsNil()
	Assert(t).That(items).Equals([]uint64{1, 2, 3})
}

func TestXMLDeserializer_ReturnError(t *testing.T) {
	var value string
	deserializer := newXMLDeserializer()
	err := deserializer.Deserialize(&value, bytes.NewBufferString(`{`))

	Assert(t).That(err).Equals(ErrDeserializationFailure)
	Assert(t).That(value).Equals("")
}
func TestJSONDeserializer_SuccessAfterFailure(t *testing.T) {
	var value1, value2 string
	deserializer := newJSONDeserializer()

	err1 := deserializer.Deserialize(&value1, &FakeFailingStream{})
	err2 := deserializer.Deserialize(&value2, bytes.NewBufferString(`"hello"`))

	Assert(t).That(err1).Equals(ErrDeserializationFailure)
	Assert(t).That(err2).IsNil()
	Assert(t).That(value2).Equals("hello")
}
func TestXMLDeserializer_SuccessAfterFailure(t *testing.T) {
	var value1, value2 string
	deserializer := newXMLDeserializer()

	err1 := deserializer.Deserialize(&value1, &FakeFailingStream{})
	err2 := deserializer.Deserialize(&value2, bytes.NewBufferString(`"<string>hello</string>"`))

	Assert(t).That(err1).Equals(ErrDeserializationFailure)
	Assert(t).That(err2).IsNil()
	Assert(t).That(value2).Equals("hello")
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TestJSONSerializer(t *testing.T) {
	serializer := newJSONSerializer()
	buffer := bytes.NewBufferString("")

	err := serializer.Serialize(buffer, "hello")

	Assert(t).That(err).IsNil()
	Assert(t).That(buffer.String()).Equals(`"hello"` + "\n")
	Assert(t).That(serializer.ContentType()).Equals("application/json; charset=utf-8")
}
func TestXMLSerializer(t *testing.T) {
	serializer := newXMLSerializer()
	buffer := bytes.NewBufferString("")

	err := serializer.Serialize(buffer, "hello")

	Assert(t).That(err).IsNil()
	Assert(t).That(buffer.String()).Equals("<string>hello</string>")
	Assert(t).That(serializer.ContentType()).Equals("application/xml; charset=utf-8")
}
func TestJSONSerializer_Failure(t *testing.T) {
	serializer := newJSONSerializer()
	buffer := bytes.NewBufferString("")

	err := serializer.Serialize(buffer, make(chan string))

	Assert(t).That(err).Equals(ErrSerializationFailure)
	Assert(t).That(buffer.Len()).Equals(0)
}
func TestXMLSerializer_Failure(t *testing.T) {
	serializer := newXMLSerializer()
	buffer := bytes.NewBufferString("")

	err := serializer.Serialize(buffer, make(chan string))

	Assert(t).That(err).Equals(ErrSerializationFailure)
	Assert(t).That(buffer.Len()).Equals(0)
}
func TestJSONSerializer_SuccessAfterFailure(t *testing.T) {
	serializer := newJSONSerializer()
	buffer := bytes.NewBufferString("")

	err1 := serializer.Serialize(FakeFailingStream{}, "hello")
	err2 := serializer.Serialize(buffer, "hello")

	Assert(t).That(err1).Equals(ErrSerializationFailure)
	Assert(t).That(err2).IsNil()
	Assert(t).That(buffer.String()).Equals(`"hello"` + "\n")
}
func TestXMLSerializer_SuccessAfterFailure(t *testing.T) {
	serializer := newXMLSerializer()
	buffer := bytes.NewBufferString("")

	err1 := serializer.Serialize(FakeFailingStream{}, "hello")
	err2 := serializer.Serialize(buffer, "hello")

	Assert(t).That(err1).Equals(ErrSerializationFailure)
	Assert(t).That(err2).IsNil()
	Assert(t).That(buffer.String()).Equals("<string>hello</string>")
}

type FakeFailingStream struct{}

func (this FakeFailingStream) Write([]byte) (int, error) { return 0, errors.New("write failure!") }
func (this FakeFailingStream) Read([]byte) (int, error)  { return 0, errors.New("read failure!") }
