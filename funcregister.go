package modbus

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/shasderias/modbus/internal/databuilder"
)

type ReadRegisterRequest struct {
	functionCode byte
	startAddress uint16
	count        uint16
}

func NewReadRegisterRequest(
	functionCode, startAddress, count int) (
	*ReadRegisterRequest, error) {

	if functionCode < 1 || functionCode >= 0x80 {
		return nil, fmt.Errorf("function code out of range [1, 0x80): %v", functionCode)
	}
	if startAddress < 0 || startAddress > 0xff {
		return nil, fmt.Errorf("start address out of range [0, 0xff]: %v", startAddress)
	}
	if startAddress+count > 0xff {
		return nil, fmt.Errorf("requested addresses out of range: start address: %v, register count: %v", startAddress, count)
	}

	// function code, byte count
	if 1+1+2*count > MaximumPDUSize {
		return nil, errors.New("response would exceed maximum PDU size")
	}

	return &ReadRegisterRequest{
		functionCode: byte(functionCode),
		startAddress: uint16(startAddress),
		count:        uint16(count),
	}, nil
}

func (r *ReadRegisterRequest) FunctionCode() byte { return r.functionCode }

func (r *ReadRegisterRequest) StartAddress() uint16  { return r.startAddress }
func (r *ReadRegisterRequest) RegisterCount() uint16 { return r.count }

func (r *ReadRegisterRequest) MarshalBinary() ([]byte, error) {
	buf := databuilder.New(5)
	buf.WriteBytes(r.functionCode)
	buf.WriteUint16(r.startAddress, r.count)
	return buf.Bytes(), nil
}
func (r *ReadRegisterRequest) UnmarshalBinary(data []byte) error {
	if len(data) != 5 {
		return fmt.Errorf("too few bytes to unmarshal as ReadRegisterRequest: %v", data)
	}

	functionCode := data[0]
	startAddress := binary.BigEndian.Uint16(data[1:3])
	count := binary.BigEndian.Uint16(data[3:5])

	req, err := NewReadRegisterRequest(int(functionCode), int(startAddress), int(count))
	if err != nil {
		return err
	}

	*r = *req

	return nil
}

type ReadRegisterResponse struct {
	functionCode byte
	values       []byte
}

func NewReadRegisterResponse(functionCode int, data []byte) (*ReadRegisterResponse, error) {
	if functionCode < 1 || functionCode >= 0x80 {
		return nil, fmt.Errorf("function code out of range [1, 0x80): %v", functionCode)
	}
	if 1+1+len(data) > MaximumPDUSize {
		return nil, fmt.Errorf("response would exceed maximum PDU size: %v", data)
	}
	if len(data)%2 != 0 {
		return nil, fmt.Errorf("register status should have an even number of bytes: %v", data)
	}

	return &ReadRegisterResponse{
		byte(functionCode), data,
	}, nil
}

func NewReadRegisterResponseFromUint16s(functionCode int, data []uint16) (*ReadRegisterResponse, error) {
	buf := databuilder.New(2 * len(data))
	for _, d := range data {
		buf.WriteUint16(d)
	}
	return NewReadRegisterResponse(functionCode, buf.Bytes())
}

func (r *ReadRegisterResponse) FunctionCode() byte { return r.functionCode }

func (r *ReadRegisterResponse) MarshalBinary() ([]byte, error) {
	buf := databuilder.New(2 + len(r.values))
	buf.WriteBytes(r.functionCode, byte(len(r.values)))
	buf.WriteBytes(r.values...)
	return buf.Bytes(), nil
}

func (r *ReadRegisterResponse) UnmarshalBinary(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("too few bytes to unmarshal as ReadRegisterResponse: %v", data)
	}

	functionCode := data[0]

	byteCount := data[1]
	if byteCount%2 != 0 {
		return fmt.Errorf("byte count for read register response should be even: %v", data)
	}
	if len(data)-2 != int(byteCount) {
		return fmt.Errorf("actual byte count does not match PDU byte count: %d != %d", len(data)-2, byteCount)
	}
	registerData := data[2:]

	resp, err := NewReadRegisterResponse(int(functionCode), registerData)
	if err != nil {
		return err
	}

	*r = *resp

	return nil
}

func (r *ReadRegisterResponse) Values() []byte {
	return r.values
}

func (r *ReadRegisterResponse) Uint16() []uint16 {
	vals := make([]uint16, len(r.values)/2)
	for i := 0; i < len(vals); i++ {
		vals[i] = binary.BigEndian.Uint16(r.values[i*2 : i*2+2])
	}
	return vals
}

func (r *ReadRegisterResponse) Float32() ([]float32, error) {
	if len(r.values)%4 != 0 {
		return nil, fmt.Errorf("modbus: need a multiple of 4 registers to interpret as float32s: %v", r.values)
	}

	vals := make([]float32, len(r.values)/4)
	for i := 0; i < len(vals); i++ {
		// Open Modbus TCPIP.doc 3/29/99
		// Intel single precision real ... First register contains bits 15-0 ... Second register contains bits 31-16 ...
		lo := uint32(binary.BigEndian.Uint16(r.values[i*4+0 : i*4+2]))
		hi := uint32(binary.BigEndian.Uint16(r.values[i*4+2 : i*4+4]))
		vals[i] = math.Float32frombits(hi<<16 | lo)
	}
	return vals, nil
}

type WriteSingleRegisterRequest struct {
	functionCode byte
	address      uint16
	value        []byte
}

func NewWriteSingleRegisterRequest(functionCode, address int, value []byte) (
	*WriteSingleRegisterRequest, error) {

	if functionCode < 1 || functionCode >= 0x80 {
		return nil, fmt.Errorf("function code out of range [1, 0x80): %v", functionCode)
	}
	if address < 0 || address > 0xffff {
		return nil, fmt.Errorf("address out of range [0, 0xffff]: %v", address)
	}
	if len(value) != 2 {
		return nil, fmt.Errorf("register value must be 2 bytes long: %v", value)
	}

	return &WriteSingleRegisterRequest{
		functionCode: byte(functionCode),
		address:      uint16(address),
		value:        value,
	}, nil
}

func NewWriteSingleRegisterRequestFromUint16(functionCode, address int, value uint16) (
	*WriteSingleRegisterRequest, error) {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, value)
	return NewWriteSingleRegisterRequest(functionCode, address, buf)
}

func (w *WriteSingleRegisterRequest) FunctionCode() byte { return w.functionCode }

func (w *WriteSingleRegisterRequest) Address() uint16 { return w.address }
func (w *WriteSingleRegisterRequest) Value() []byte   { return w.value }

func (w *WriteSingleRegisterRequest) MarshalBinary() ([]byte, error) {
	buf := databuilder.New(5)
	buf.WriteBytes(w.functionCode)
	buf.WriteUint16(w.address)
	buf.WriteBytes(w.value...)
	return buf.Bytes(), nil
}
func (w *WriteSingleRegisterRequest) UnmarshalBinary(data []byte) error {
	if len(data) != 5 {
		return fmt.Errorf("too few bytes to unmarshal as WriteSingleRegisterRequest: %v", data)
	}

	functionCode := data[0]
	address := binary.BigEndian.Uint16(data[1:3])
	value := data[3:5]

	req, err := NewWriteSingleRegisterRequest(int(functionCode), int(address), value)
	if err != nil {
		return err
	}

	*w = *req

	return nil
}

type WriteSingleRegisterResponse struct {
	WriteSingleRegisterRequest
}

func NewWriteSingleRegisterResponse(functionCode, address int, value []byte) (
	*WriteSingleRegisterResponse, error) {

	req, err := NewWriteSingleRegisterRequest(functionCode, address, value)
	if err != nil {
		return nil, err
	}

	return &WriteSingleRegisterResponse{*req}, nil
}
func NewWriteSingleRegisterResponseFromUint16(functionCode, address int, value uint16) (
	*WriteSingleRegisterResponse, error) {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, value)
	return NewWriteSingleRegisterResponse(functionCode, address, buf)
}

type WriteMultipleRegistersRequest struct {
	functionCode byte
	address      uint16
	count        uint16
	values       []byte
}

func NewWriteMultipleRegistersRequest(functionCode, address int, values []byte) (*WriteMultipleRegistersRequest, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("number of values bytes not even: %d", len(values))
	}

	count := len(values) / 2
	if count < 1 || count > 0x007b {
		return nil, fmt.Errorf("number of values %d out of range [1, 0x007b]", count)
	}

	return &WriteMultipleRegistersRequest{
		functionCode: byte(functionCode),
		address:      uint16(address),
		count:        uint16(count),
		values:       values,
	}, nil
}

func NewWriteMultipleRegistersRequestFromUint16s(functionCode, address int, values []uint16) (*WriteMultipleRegistersRequest, error) {
	buf := make([]byte, len(values)*2)
	for i, v := range values {
		binary.BigEndian.PutUint16(buf[i*2:], v)
	}

	return NewWriteMultipleRegistersRequest(functionCode, address, buf)
}

func (w *WriteMultipleRegistersRequest) FunctionCode() byte { return w.functionCode }

func (w *WriteMultipleRegistersRequest) MarshalBinary() ([]byte, error) {
	buf := databuilder.New(6 + len(w.values))
	buf.WriteBytes(w.functionCode)
	buf.WriteUint16(w.address)
	buf.WriteUint16(w.count)
	buf.WriteBytes(byte(len(w.values)))
	buf.WriteBytes(w.values...)
	return buf.Bytes(), nil
}

func (w *WriteMultipleRegistersRequest) UnmarshalBinary(data []byte) error {
	if len(data) < 7 {
		return fmt.Errorf("too few bytes to unmarshal as WriteMultipleRegistersRequest: %v", data)
	}

	functionCode := data[0]
	address := binary.BigEndian.Uint16(data[1:3])
	count := binary.BigEndian.Uint16(data[3:5])
	byteCount := data[5]

	if len(data) != 6+int(byteCount) {
		return fmt.Errorf("PDU length %d inconsistent with byte count value %d: %v", len(data), 6+byteCount, data)
	}
	if int(byteCount) != int(count*2) {
		return fmt.Errorf("data length %d inconsistent with register count %d, expected %d bytes of data: %v",
			byteCount, count, count*2, data)
	}
	values := data[6:]

	req, err := NewWriteMultipleRegistersRequest(int(functionCode), int(address), values)
	if err != nil {
		return err
	}

	*w = *req

	return nil
}

type WriteMultipleRegistersResponse struct {
	functionCode byte
	address      uint16
	count        uint16
}

func NewWriteMultipleRegistersResponse(functionCode, address, count int) (*WriteMultipleRegistersResponse, error) {
	if functionCode < 1 || functionCode >= 0x80 {
		return nil, fmt.Errorf("function code out of range [1, 0x80): %v", functionCode)
	}
	if address < 0 || address > 0xffff {
		return nil, fmt.Errorf("address out of range [0, 0xffff]: %v", address)
	}
	if count < 1 || count > 0x007b {
		return nil, fmt.Errorf("count out of range [1, 0x007b]: %v", count)
	}
	if address+count > 0xffff {
		return nil, fmt.Errorf("address + count out of range [1, 0xffff]: %v", address+count)
	}

	return &WriteMultipleRegistersResponse{
		functionCode: byte(functionCode),
		address:      uint16(address),
		count:        uint16(count),
	}, nil
}

func (w *WriteMultipleRegistersResponse) FunctionCode() byte { return w.functionCode }

func (w *WriteMultipleRegistersResponse) MarshalBinary() ([]byte, error) {
	buf := databuilder.New(5)
	buf.WriteBytes(w.functionCode)
	buf.WriteUint16(w.address, w.count)
	return buf.Bytes(), nil
}

func (w *WriteMultipleRegistersResponse) UnmarshalBinary(data []byte) error {
	if len(data) != 5 {
		return fmt.Errorf("too few bytes to unmarshal as WriteMultipleRegistersResponse: %v", data)
	}

	functionCode := data[0]
	address := binary.BigEndian.Uint16(data[1:3])
	count := binary.BigEndian.Uint16(data[3:5])

	req, err := NewWriteMultipleRegistersResponse(int(functionCode), int(address), int(count))
	if err != nil {
		return err
	}

	*w = *req

	return nil
}
