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
