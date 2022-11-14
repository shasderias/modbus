package tcp

import (
	"io"
	"time"

	"github.com/shasderias/modbus"
	"github.com/shasderias/modbus/internal/databuilder"
)

const (
	maxFrameSize = 256 + 7 // MODBUS Messaging on TCP/IP Implementation Guide V1.0b, 4.3.2
)

type Conn interface {
	io.ReadWriteCloser
	SetReadDeadline(time.Time) error
}

func assembleFrame(txID uint16, unitID byte, r modbus.PDU) []byte {
	pduBytes, err := r.MarshalBinary()
	if err != nil {
		panic(err)
	}

	buf := databuilder.New(7 + len(pduBytes))

	buf.WriteUint16(txID)
	buf.WriteUint16(0, uint16(1+len(pduBytes)))
	buf.WriteBytes(unitID)
	buf.WriteBytes(pduBytes...)

	return buf.Bytes()
}
