package resp

import (
	"avacado/internal/protocol"
	"bytes"
	"fmt"
	"strconv"
)

const newLineCarriageReturn = "\r\n"

type Serializer struct {
}

func NewRESPSerializer() *Serializer {
	return &Serializer{}
}

func (s *Serializer) Serialize(response *protocol.Response) ([]byte, error) {
	buf := &bytes.Buffer{}
	if err := s.writeValue(buf, response.Value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *Serializer) SerializeError(e error) []byte {
	buf := &bytes.Buffer{}
	buf.WriteByte(TypeError)
	buf.WriteString(e.Error())
	buf.WriteString(newLineCarriageReturn)
	return buf.Bytes()
}

func (s *Serializer) writeValue(buf *bytes.Buffer, value protocol.Value) error {
	switch value.Type {
	case protocol.TypeSimpleString:
		return s.writeSimpleString(buf, value.Str)
	case protocol.TypeBulkString:
		return s.writeBulkString(buf, value.Bytes)
	case protocol.TypeNumber:
		return s.writeNumber(buf, value.Number)
	case protocol.TypeArray:
		return s.writeArray(buf, value.Array)
	}
	return fmt.Errorf("no serializer found for value type: %s", string(value.Type))
}

func (s *Serializer) writeSimpleString(buf *bytes.Buffer, str string) error {
	buf.WriteByte(TypeSimpleString)
	buf.Write([]byte(str))
	buf.WriteString(newLineCarriageReturn)
	return nil
}

func (s *Serializer) writeBulkString(buf *bytes.Buffer, str []byte) error {
	buf.WriteByte(TypeBulkString)
	buf.WriteString(strconv.FormatInt(int64(len(str)), 10))
	buf.WriteString(newLineCarriageReturn)
	buf.Write(str)
	buf.WriteString(newLineCarriageReturn)
	return nil
}

func (s *Serializer) writeNumber(buf *bytes.Buffer, number int64) error {
	buf.WriteByte(TypeInteger)
	buf.Write([]byte(strconv.FormatInt(number, 10)))
	buf.WriteString(newLineCarriageReturn)
	return nil
}

func (s *Serializer) writeArray(buf *bytes.Buffer, array []protocol.Value) error {
	buf.WriteByte(TypeArray)
	buf.WriteString(strconv.FormatInt(int64(len(array)), 10))
	buf.WriteString(newLineCarriageReturn)
	for _, item := range array {
		if err := s.writeValue(buf, item); err != nil {
			return err
		}
	}
	return nil
}
