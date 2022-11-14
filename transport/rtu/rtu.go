package rtu

import (
	"encoding/binary"

	"github.com/shasderias/modbus"
	"github.com/shasderias/modbus/internal/crc"
	"github.com/shasderias/modbus/internal/databuilder"
)

const (
	maxFrameLength int = 256
)

func assembleFrame(slaveAddress byte, pdu modbus.PDU) []byte {
	pduBytes, err := pdu.MarshalBinary()
	if err != nil {
		panic(err)
	}

	buf := databuilder.New(1 + len(pduBytes) + 2)

	buf.WriteBytes(slaveAddress)
	buf.WriteBytes(pduBytes...)

	return buf.BytesWithCRC()
}

func decodeFrame(b []byte) (*modbus.RawPDU, error) {
	var (
		_        = b[0] // slave address
		pduBytes = b[1 : len(b)-2]
		pduCRC   = binary.LittleEndian.Uint16(b[len(b)-2:])
		wantCRC  = crc.Checksum(b[:len(b)-2])
	)

	if pduCRC != wantCRC {
		return nil, modbus.ErrBadCRC{
			Got: pduCRC, Want: wantCRC,
			Frame: b,
		}
	}

	return modbus.NewRawPDU(pduBytes)
}
