package modbus

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/shasderias/modbus/internal/databuilder"
)

type ReadBitRequest struct {
	functionCode byte
	startAddress uint16
	count        uint16
}

func NewReadBitRequest(functionCode byte, startAddress, count int) (*ReadBitRequest, error) {
	if functionCode < 1 || functionCode >= 0x80 {
		return nil, fmt.Errorf("modbus: function code out of range [1, 0x80): %x", functionCode)
	}
	if startAddress < 0 || startAddress > 0xffff {
		return nil, fmt.Errorf("modbus: start address out of range[0, 0xffff]: %x", startAddress)
	}
	if count < 1 || count > MaximumReadBitRequestCount {
		return nil, fmt.Errorf("modbus: count out of range [1, 0x7d0]: %x", count)
	}
	if startAddress+count > 0xffff {
		return nil, fmt.Errorf("modbus: requested addresses out of range: start address: %d, count: %d", startAddress, count)
	}

	return &ReadBitRequest{functionCode, uint16(startAddress), uint16(count)}, nil
}

func (r *ReadBitRequest) FunctionCode() byte   { return r.functionCode }
func (r *ReadBitRequest) StartAddress() uint16 { return r.startAddress }
func (r *ReadBitRequest) BitCount() uint16     { return r.count }

func (r *ReadBitRequest) MarshalBinary() ([]byte, error) {
	buf := databuilder.New(5)
	buf.WriteBytes(r.functionCode)
	buf.WriteUint16(r.startAddress, r.count)
	return buf.Bytes(), nil
}
func (r *ReadBitRequest) UnmarshalBinary(data []byte) error {
	if len(data) != 5 {
		return fmt.Errorf("modbus: too few bytes to unmarshal as ReadBitRequest: %v", data)
	}

	req, err := NewReadBitRequest(
		data[0],
		int(binary.BigEndian.Uint16(data[1:3])),
		int(binary.BigEndian.Uint16(data[3:5])))
	if err != nil {
		return err
	}

	*r = *req

	return nil
}

type ReadBitResponse struct {
	functionCode byte
	bitValues    []byte
}

func NewReadBitResponse(functionCode byte, bitStatus []byte) (*ReadBitResponse, error) {
	if functionCode < 1 || functionCode >= 0x80 {
		return nil, fmt.Errorf("modbus: function code out of range [1, 0x80): %v", functionCode)
	}

	// function code, byte count
	if 1+1+len(bitStatus) > MaximumPDUSize {
		return nil, fmt.Errorf("modbus: PDU exceeds size limit")
	}

	return &ReadBitResponse{
		functionCode, bitStatus,
	}, nil
}

func NewReadBitResponseFromBool(functionCode byte, bitStatus []bool) (*ReadBitResponse, error) {
	count := len(bitStatus)
	if count < 1 || count > MaximumReadBitRequestCount {
		return nil, fmt.Errorf("modbus: bit status count out of range [1, 2000]: %d", count)
	}

	bitStatusBytes := make([]byte, (count-1)/8+1)

	for i, b := range bitStatus {
		if b {
			bitStatusBytes[i/8] |= 1 << uint(i%8)
		}
	}

	return NewReadBitResponse(functionCode, bitStatusBytes)
}

func (r *ReadBitResponse) FunctionCode() byte { return r.functionCode }
func (r *ReadBitResponse) Values() []byte     { return r.bitValues }
func (r *ReadBitResponse) BitValues() []bool {
	return bytesToBools(r.bitValues)
}

func (r *ReadBitResponse) MarshalBinary() ([]byte, error) {
	buf := databuilder.New(2 + len(r.bitValues))
	buf.WriteBytes(r.functionCode, byte(len(r.bitValues)))
	buf.WriteBytes(r.bitValues...)
	return buf.Bytes(), nil
}
func (r *ReadBitResponse) UnmarshalBinary(data []byte) error {
	if len(data) < 3 {
		return fmt.Errorf("modbus: at least 3 bytes required to unmarshal as ReadBitResponse: %v", data)
	}

	byteCount := data[1]
	if len(data)-2 != int(byteCount) {
		return fmt.Errorf("modbus: expected frame length based on PDU byte count field: %d, actual frame length: %d", 2+byteCount, len(data))
	}

	values := make([]byte, byteCount)
	copy(values, data[2:])

	resp, err := NewReadBitResponse(data[0], values)
	if err != nil {
		return err
	}

	*r = *resp

	return nil
}

type WriteSingleBitRequest struct {
	functionCode byte
	startAddress uint16
	value        bool
}

func NewWriteSingleBitRequest(functionCode byte, startAddress int, value bool) (*WriteSingleBitRequest, error) {
	if functionCode < 1 || functionCode >= 0x80 {
		return nil, fmt.Errorf("modbus: function code out of range [1, 0x80): %v", functionCode)
	}
	if startAddress < 0 || startAddress > 0xffff {
		return nil, fmt.Errorf("modbus: start address out of range[0, 0xffff]: %x", startAddress)
	}

	return &WriteSingleBitRequest{functionCode, uint16(startAddress), value}, nil
}

func (r *WriteSingleBitRequest) FunctionCode() byte   { return r.functionCode }
func (r *WriteSingleBitRequest) StartAddress() uint16 { return r.startAddress }
func (r *WriteSingleBitRequest) BitValue() bool       { return r.value }

func (r *WriteSingleBitRequest) MarshalBinary() ([]byte, error) {
	buf := databuilder.New(5)
	buf.WriteBytes(r.functionCode)
	buf.WriteUint16(r.startAddress)
	if r.value {
		buf.WriteUint16(0xff00)
	} else {
		buf.WriteUint16(0x0000)
	}
	return buf.Bytes(), nil
}
func (r *WriteSingleBitRequest) UnmarshalBinary(data []byte) error {
	if len(data) != 5 {
		return fmt.Errorf("modbus: exactly 5 bytes required to unmarshal as WriteSingleBitRequest: %v", data)
	}
	if bytes.Compare(data[3:5], []byte{0x00, 0x00}) != 0 && bytes.Compare(data[3:5], []byte{0xff, 0x00}) != 0 {
		return fmt.Errorf("modbus: invalid value field: %v", data[3:5])
	}

	req, err := NewWriteSingleBitRequest(data[0], int(binary.BigEndian.Uint16(data[1:3])), data[3] != 0x00)
	if err != nil {
		return err
	}

	*r = *req

	return nil
}

type WriteSingleBitResponse struct {
	WriteSingleBitRequest
}

func NewWriteSingleBitResponse(functionCode byte, startAddress int, value bool) (*WriteSingleBitResponse, error) {
	req, err := NewWriteSingleBitRequest(functionCode, startAddress, value)
	if err != nil {
		return nil, err
	}

	return &WriteSingleBitResponse{*req}, nil
}

type WriteMultipleBitsRequest struct {
	functionCode byte
	startAddress uint16
	count        uint16
	values       []byte
}

func NewWriteMultipleBitsRequest(functionCode byte, startAddress, count int, values []byte) (
	*WriteMultipleBitsRequest, error) {

	if functionCode < 1 || functionCode >= 0x80 {
		return nil, fmt.Errorf("modbus: function code out of range [1, 0x80): %v", functionCode)
	}
	if startAddress < 0 || startAddress > 0xffff {
		return nil, fmt.Errorf("modbus: address out of range [0, 0xffff]: %v", startAddress)
	}
	if count < 1 || count > 0x07b0 {
		return nil, fmt.Errorf("modbus: count out of range [1, 0x07b0]: %v", count)
	}
	if startAddress+count > 0xffff {
		return nil, fmt.Errorf("modbus: address + count out of range [1, 0xffff]: %v", startAddress+count)
	}

	expectedLen := (count-1)/8 + 1
	if len(values) != expectedLen {
		return nil, fmt.Errorf("modbus: expected %d bytes of values for %d bits, got %d", expectedLen, count, len(values))
	}

	return &WriteMultipleBitsRequest{
		functionCode: functionCode,
		startAddress: uint16(startAddress),
		count:        uint16(count),
		values:       values,
	}, nil
}
func NewWriteMultipleBitsRequestFromBools(functionCode byte, address int, values []bool) (
	*WriteMultipleBitsRequest, error) {

	count := len(values)
	if count < 1 || count > 0x07b0 {
		return nil, fmt.Errorf("modbus: number of values %d out of range [1, 0x07b0]", count)
	}

	buf := make([]byte, (count-1)/8+1)

	for i, v := range values {
		if v {
			buf[i/8] |= 1 << uint(i%8)
		}
	}

	return NewWriteMultipleBitsRequest(functionCode, address, count, buf)
}

func (r *WriteMultipleBitsRequest) FunctionCode() byte   { return r.functionCode }
func (r *WriteMultipleBitsRequest) StartAddress() uint16 { return r.startAddress }
func (r *WriteMultipleBitsRequest) Values() []byte       { return r.values }
func (r *WriteMultipleBitsRequest) BitValues() []bool {
	return bytesToBools(r.values)
}

func (r *WriteMultipleBitsRequest) MarshalBinary() ([]byte, error) {
	buf := databuilder.New(6 + len(r.values))
	buf.WriteBytes(r.functionCode)
	buf.WriteUint16(r.startAddress)
	buf.WriteUint16(r.count)
	buf.WriteBytes(byte(len(r.values)))
	buf.WriteBytes(r.values...)
	return buf.Bytes(), nil
}

func (r *WriteMultipleBitsRequest) UnmarshalBinary(data []byte) error {
	if len(data) < 7 {
		return fmt.Errorf("modbus: at least 7 bytes required to unmarshal as WriteMultipleBitsRequest: %v", data)
	}

	count := binary.BigEndian.Uint16(data[3:5])
	byteCount := data[5]

	if len(data) != 6+int(byteCount) {
		return fmt.Errorf("modbus: PDU length %d inconsistent with byte count value %d: %v", len(data), 6+byteCount, data)
	}
	if int(byteCount) != int((count-1)/8+1) {
		return fmt.Errorf("modbus: data length %d inconsistent with bit count %d, expected %d bytes of data: %v",
			byteCount, count, (count-1)/8+1, data)
	}

	values := make([]byte, byteCount)
	copy(values, data[6:])

	req, err := NewWriteMultipleBitsRequest(data[0], int(binary.BigEndian.Uint16(data[1:3])), int(binary.BigEndian.Uint16(data[3:5])), values)
	if err != nil {
		return err
	}

	*r = *req

	return nil
}

type WriteMultipleBitsResponse struct {
	functionCode byte
	startAddress uint16
	count        uint16
}

func NewWriteMultipleBitsResponse(functionCode byte, address, count int) (*WriteMultipleBitsResponse, error) {
	if functionCode < 1 || functionCode >= 0x80 {
		return nil, fmt.Errorf("modbus: function code out of range [1, 0x80): %v", functionCode)
	}
	if address < 0 || address > 0xffff {
		return nil, fmt.Errorf("modbus: address out of range [0, 0xffff]: %v", address)
	}
	if count < 1 || count > 0x07b0 {
		return nil, fmt.Errorf("modbus: count out of range [1, 0x07b0]: %v", count)
	}
	if address+count > 0xffff {
		return nil, fmt.Errorf("modbus: address + count out of range [1, 0xffff]: %v", address+count)
	}

	return &WriteMultipleBitsResponse{
		functionCode: functionCode,
		startAddress: uint16(address),
		count:        uint16(count),
	}, nil
}

func (r *WriteMultipleBitsResponse) FunctionCode() byte   { return r.functionCode }
func (r *WriteMultipleBitsResponse) StartAddress() uint16 { return r.startAddress }
func (r *WriteMultipleBitsResponse) Count() uint16        { return r.count }
func (r *WriteMultipleBitsResponse) MarshalBinary() ([]byte, error) {
	buf := databuilder.New(5)
	buf.WriteBytes(r.functionCode)
	buf.WriteUint16(r.startAddress)
	buf.WriteUint16(r.count)
	return buf.Bytes(), nil
}
func (r *WriteMultipleBitsResponse) UnmarshalBinary(data []byte) error {
	if len(data) != 5 {
		return fmt.Errorf("modbus: too few bytes to unmarshal as WriteMultipleBitsResponse: %v", data)
	}

	req, err := NewWriteMultipleBitsResponse(data[0], int(binary.BigEndian.Uint16(data[1:3])), int(binary.BigEndian.Uint16(data[3:5])))
	if err != nil {
		return err
	}

	*r = *req

	return nil
}

func bytesToBools(slice []byte) []bool {
	values := make([]bool, len(slice)*8)
	for i := 0; i < len(slice); i++ {
		for j := 0; j < 8; j++ {
			values[i*8+j] = slice[i]&(1<<uint(j)) != 0
		}
	}
	return values
}
