package modbus

import (
	"fmt"
)

const (
	FuncCodeReadCoils                  = 0x01
	FuncCodeReadDiscreteInputs         = 0x02
	FuncCodeReadHoldingRegisters       = 0x03
	FuncCodeReadInputRegisters         = 0x04
	FuncCodeWriteSingleCoil            = 0x05
	FuncCodeWriteSingleRegister        = 0x06
	FuncCodeReadExceptionStatus        = 0x07
	FuncCodeDiagnostic                 = 0x08
	FuncCodeGetCommEventCounter        = 0x0b
	FuncCodeGetCommEventLog            = 0x0c
	FuncCodeReportServerID             = 0x11
	FuncCodeReadFileRecord             = 0x14
	FuncCodeWriteFileRecord            = 0x15
	FuncCodeWriteMultipleCoils         = 0x0f
	FuncCodeWriteMultipleRegisters     = 0x10
	FuncCodeMaskWriteRegister          = 0x16
	FuncCodeReadWriteMultipleRegisters = 0x17
	FuncCodeReadFIFOQueue              = 0x18
)

const (
	MaximumPDUSize             = 253
	MaximumReadBitRequestCount = 0x7d0 // 2000
)

type ErrBadCRC struct {
	Got, Want uint16
	Frame     []byte
}

func (e ErrBadCRC) Error() string {
	return fmt.Sprintf("bad CRC; got: %x, want: %x, frame: %x", e.Got, e.Want, e.Frame)
}

type Transport interface {
	WriteRequest(slaveAddress byte, r PDU) (PDU, error)
	WriteResponse(slaveAddress byte, r PDU) error
	ReadRequest() (slaveAddress byte, r PDU, err error)
}

type PDU interface {
	FunctionCode() byte

	MarshalBinary() ([]byte, error)
	UnmarshalBinary(data []byte) error
}

func UnmarshalAs(p PDU, target PDU) error {
	pduBytes, err := p.MarshalBinary()
	if err != nil {
		return err
	}

	return target.UnmarshalBinary(pduBytes)
}

type RawPDU struct {
	b []byte
}

func NewRawPDU(b []byte) (*RawPDU, error) {
	if len(b) < 2 {
		return nil, fmt.Errorf("insufficient bytes for PDU: %v", b)
	}
	return &RawPDU{b}, nil
}

func (p *RawPDU) FunctionCode() byte { return p.b[0] }

func (p *RawPDU) MarshalBinary() ([]byte, error) {
	return p.b, nil
}
func (p *RawPDU) UnmarshalBinary(data []byte) error {
	d := make([]byte, len(data))
	copy(d, data)

	pdu, err := NewRawPDU(d)
	if err != nil {
		return err
	}

	*p = *pdu

	return nil
}
