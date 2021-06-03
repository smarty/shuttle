package shuttle

import (
	"encoding/json"
	"io"
)

type jsonDeserializer struct {
	decoder *json.Decoder
	source  struct{ io.Reader }
}

func newJSONDeserializer() Deserializer {
	this := &jsonDeserializer{}
	this.decoder = json.NewDecoder(&this.source)
	return this
}

func (this *jsonDeserializer) Deserialize(target interface{}, source io.Reader) error {
	this.source.Reader = source

	if err := this.decoder.Decode(target); err != nil {
		return ErrDeserializationFailure
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type jsonSerializer struct {
	encoder *json.Encoder
	target  struct{ io.Writer }
}

func newJSONSerializer() Serializer {
	this := &jsonSerializer{}
	this.encoder = json.NewEncoder(&this.target)
	return this
}

func (this *jsonSerializer) Serialize(target io.Writer, source interface{}) error {
	this.target.Writer = target

	if err := this.encoder.Encode(source); err != nil {
		return ErrSerializationFailure
	}

	return nil
}
func (this *jsonSerializer) ContentType() string { return mimeTypeApplicationJSONUTF8 }
