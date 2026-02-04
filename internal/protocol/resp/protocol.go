package resp

import (
	"avacado/internal/protocol"
	"io"
)

type Protocol struct {
	serializer *Serializer
}

func (r *Protocol) CreateParser(reader io.Reader) protocol.Parser {
	return NewCommandParser(reader)
}

func (r *Protocol) Serialize(value *protocol.Response) ([]byte, error) {
	return r.serializer.Serialize(value)
}

func (r *Protocol) SerializeError(e error) []byte {
	return r.serializer.SerializeError(e)
}

func NewRespProtocol() protocol.Protocol {
	return &Protocol{serializer: NewRESPSerializer()}
}
