package databuilder

import (
	"encoding/binary"

	"github.com/shasderias/modbus/internal/crc"
)

type DataBuilder struct {
	buf []byte
}

func New(l int) *DataBuilder {
	return &DataBuilder{
		buf: make([]byte, 0, l),
	}
}

func (b *DataBuilder) WriteCRC(checksum uint16) *DataBuilder {
	b.buf = binary.LittleEndian.AppendUint16(b.buf, checksum)
	return b
}

func (b *DataBuilder) WriteUint16(uint16s ...uint16) *DataBuilder {
	for _, v := range uint16s {
		b.buf = binary.BigEndian.AppendUint16(b.buf, v)
	}
	return b
}

func (b *DataBuilder) WriteBytes(bytes ...byte) *DataBuilder {
	b.buf = append(b.buf, bytes...)
	return b
}

func (b *DataBuilder) Bytes() []byte {
	return b.buf
}

func (b *DataBuilder) BytesWithCRC() []byte {
	return binary.LittleEndian.AppendUint16(b.buf, crc.Checksum(b.buf))
}
