package resp

import (
	"avacado/internal/protocol"
	"io"
)

type Protocol struct {
	serializer *Serializer
}

func (r *Protocol) Serialize(value *protocol.Response) ([]byte, error) {
	return r.serializer.Serialize(value)
}

func (r *Protocol) SerializeError(e error) []byte {
	return r.serializer.SerializeError(e)
}

func (r *Protocol) Parse(reader io.Reader) (*protocol.Message, error) {
	return NewCommandParser().Parse(reader)
}

func NewRespProtocol() protocol.Protocol {
	return &Protocol{serializer: NewRESPSerializer()}
}
