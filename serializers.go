package shuttle

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"io"
)

type jsonDeserializer struct{}

func newJSONDeserializer() Deserializer { return &jsonDeserializer{} }

func (this *jsonDeserializer) Deserialize(target any, source io.Reader) error {
	if err := json.NewDecoder(source).Decode(target); err == nil {
		return nil
	} else if err == io.EOF {
		return nil
	} else {
		return ErrDeserializationFailure
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type jsonSerializer struct {
	encoder *json.Encoder
	target  struct{ io.Writer }
}

func NewJSONSerializer() Serializer {
	this := &jsonSerializer{}
	this.encoder = json.NewEncoder(&this.target)
	return this
}

func (this *jsonSerializer) Serialize(target io.Writer, source any) error {
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

func (this *xmlDeserializer) Deserialize(target any, source io.Reader) error {
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

func (this *xmlSerializer) Serialize(target io.Writer, source any) error {
	this.target.Writer = target

	if this.encoder.Encode(source) == nil {
		return nil
	}

	this.encoder = xml.NewEncoder(&this.target)
	return ErrSerializationFailure
}
func (this *xmlSerializer) ContentType() string { return mimeTypeApplicationXMLUTF8 }

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type csvSerializer struct {
	writer *csv.Writer
	target struct{ io.Writer }
}

func NewCSVSerializer() Serializer {
	this := &csvSerializer{}
	this.writer = csv.NewWriter(&this.target)
	return this
}

func (this *csvSerializer) Serialize(target io.Writer, source any) error {
	csvSource, ok := source.(CSV)
	if !ok {
		return ErrSerializationFailure
	}

	this.target.Writer = target

	if err := this.writer.Write(csvSource.Header()); err != nil {
		return ErrSerializationFailure
	}
	for row := range csvSource.CSV() {
		if err := this.writer.Write(row); err != nil {
			return ErrSerializationFailure
		}
	}
	this.writer.Flush()
	return this.writer.Error()
}

func (this *csvSerializer) ContentType() string { return mimeTypeTextCSV }
