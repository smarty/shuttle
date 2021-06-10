package shuttle

import (
	"encoding/json"
	"io"
)

type jsonDeserializer struct{}

func newJSONDeserializer() Deserializer {
	return &jsonDeserializer{}
}

func (this *jsonDeserializer) Deserialize(target interface{}, source io.Reader) error {
	if err := json.NewDecoder(source).Decode(target); err != nil {
		return ErrDeserializationFailure
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type jsonSerializer struct{}

func newJSONSerializer() Serializer {
	return &jsonSerializer{}
}

func (this *jsonSerializer) Serialize(target io.Writer, source interface{}) error {
	if err := json.NewEncoder(target).Encode(source); err != nil {
		return ErrSerializationFailure
	}

	return nil
}
func (this *jsonSerializer) ContentType() string { return mimeTypeApplicationJSONUTF8 }
