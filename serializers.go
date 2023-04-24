package shuttle

import (
	"encoding/json"
	"encoding/xml"
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

	if this.decoder.Decode(target) == nil {
		return nil
	}

	this.decoder = json.NewDecoder(&this.source)
	return ErrDeserializationFailure
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

	if this.encoder.Encode(source) == nil {
		return nil
	}

	this.encoder = json.NewEncoder(&this.target)
	return ErrSerializationFailure
}
func (this *jsonSerializer) ContentType() string { return mimeTypeApplicationJSONUTF8 }

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type xmlDeserializer struct {
	decoder *xml.Decoder
	source  struct{ io.Reader }
}

func newXMLDeserializer() Deserializer {
	this := &xmlDeserializer{}
	this.decoder = xml.NewDecoder(&this.source)
	return this
}

func (this *xmlDeserializer) Deserialize(target interface{}, source io.Reader) error {
	this.source.Reader = source

	if this.decoder.Decode(target) == nil {
		return nil
	}

	this.decoder = xml.NewDecoder(&this.source)
	return ErrDeserializationFailure
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type xmlSerializer struct {
	encoder *xml.Encoder
	target  struct{ io.Writer }
}

func newXMLSerializer() Serializer {
	this := &xmlSerializer{}
	this.encoder = xml.NewEncoder(&this.target)
	return this
}

func (this *xmlSerializer) Serialize(target io.Writer, source interface{}) error {
	this.target.Writer = target

	if this.encoder.Encode(source) == nil {
		return nil
	}

	this.encoder = xml.NewEncoder(&this.target)
	return ErrSerializationFailure
}
func (this *xmlSerializer) ContentType() string { return mimeTypeApplicationXMLUTF8 }
