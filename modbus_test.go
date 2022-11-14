package modbus_test

import (
	"bytes"
	"encoding"
	"reflect"
	"testing"

	"github.com/shasderias/modbus"

	"github.com/google/go-cmp/cmp"
)

func TestNewReadRegisterRequest(t *testing.T) {
	testCases := []struct {
		name string

		functionCode  int
		startAddress  int
		registerCount int

		valid        bool
		expectedData []byte
	}{
		{"Sanity",
			modbus.FuncCodeReadHoldingRegisters,
			0, 1,
			true, []byte{0x03, 0x00, 0x00, 0x00, 0x01}},
		{"Reference",
			modbus.FuncCodeReadHoldingRegisters,
			107, 3,
			true, []byte{0x03, 0x00, 0x6b, 0x00, 0x03}},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			req, err := modbus.NewReadRegisterRequest(tt.functionCode, tt.startAddress, tt.registerCount)

			if tt.valid && err != nil {
				t.Fatalf("got err; want valid request: %v", err)
			}
			if !tt.valid && err == nil {
				t.Fatalf("did not get err; want invalid request")
			}

			pduBytes, err := req.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}

			if tt.valid && !bytes.Equal(pduBytes, tt.expectedData) {
				t.Fatalf("got %v; want %v", pduBytes, tt.expectedData)
			}
		})
	}
}

func TestNewReadRegisterResponse(t *testing.T) {
	testCases := []struct {
		name string

		fn func() (*modbus.ReadRegisterResponse, error)

		valid                bool
		expectedFunctionCode byte
		expectedData         []byte
	}{
		{"Sanity",
			func() (*modbus.ReadRegisterResponse, error) {
				return modbus.NewReadRegisterResponse(modbus.FuncCodeReadHoldingRegisters,
					[]byte{0x00, 0x01, 0x00, 0x02})
			},
			true,
			3, []byte{0x03, 0x04, 0x00, 0x01, 0x00, 0x02},
		},
		{
			"Reference",
			func() (*modbus.ReadRegisterResponse, error) {
				return modbus.NewReadRegisterResponseFromUint16s(modbus.FuncCodeReadHoldingRegisters,
					[]uint16{555, 0, 100})

			},
			true,
			3, []byte{0x03, 0x06, 0x02, 0x2b, 0x00, 0x00, 0x00, 0x64},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.fn()

			if tt.valid && err != nil {
				t.Fatalf("got err; want valid response: %v", err)
			}
			if !tt.valid && err == nil {
				t.Fatalf("did not get err; want invalid response")
			}

			pduBytes, err := resp.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}

			if tt.valid && !bytes.Equal(pduBytes, tt.expectedData) {
				t.Fatalf("got %v; want %v", pduBytes, tt.expectedData)
			}
			if tt.valid && resp.FunctionCode() != tt.expectedFunctionCode {
				t.Fatalf("got %v; want %v", resp.FunctionCode(), tt.expectedFunctionCode)
			}
		})
	}
}

func TestSpecification(t *testing.T) {
	testCases := []struct {
		name     string
		fn       func(t *testing.T) (encoding.BinaryMarshaler, error)
		expected []byte
	}{
		{
			"ReadCoilsRequest",
			func(t *testing.T) (encoding.BinaryMarshaler, error) {
				return modbus.NewReadBitRequest(modbus.FuncCodeReadCoils, 19, 19)
			},
			[]byte{0x01, 0x00, 0x13, 0x00, 0x13},
		},
		{
			"ReadCoilResponse",
			func(t *testing.T) (encoding.BinaryMarshaler, error) {
				return modbus.NewReadBitResponseFromBool(modbus.FuncCodeReadCoils,
					[]bool{
						true, false, true, true, false, false, true, true,
						true, true, false, true, false, true, true, false,
						true, false, true,
					})
			},
			[]byte{0x01, 0x03, 0xcd, 0x6b, 0x05},
		},
		{
			"ReadCoilException",
			func(t *testing.T) (encoding.BinaryMarshaler, error) {
				return modbus.NewExceptionResponse(modbus.FuncCodeReadCoils|0x80, 0x01)
			},
			[]byte{0x81, 0x01},
		},
		{
			"ReadDiscreteInputsRequest",
			func(t *testing.T) (encoding.BinaryMarshaler, error) {
				return modbus.NewReadBitRequest(modbus.FuncCodeReadDiscreteInputs, 196, 22)
			},
			[]byte{0x02, 0x00, 0xc4, 0x00, 0x16},
		},
		{
			"ReadDiscreteInputsResponse",
			func(t *testing.T) (encoding.BinaryMarshaler, error) {
				return modbus.NewReadBitResponse(modbus.FuncCodeReadDiscreteInputs,
					[]byte{0b10101100, 0xdb, 0b00110101})
			},
			[]byte{0x02, 0x03, 0xac, 0xdb, 0x35},
		},
		{
			"ReadHoldingRegistersRequest",
			func(t *testing.T) (encoding.BinaryMarshaler, error) {
				return modbus.NewReadRegisterRequest(modbus.FuncCodeReadHoldingRegisters, 107, 3)

			},
			[]byte{0x03, 0x00, 0x6b, 0x00, 0x03},
		},
		{
			"ReadHoldingRegistersResponse",
			func(t *testing.T) (encoding.BinaryMarshaler, error) {
				return modbus.NewReadRegisterResponseFromUint16s(modbus.FuncCodeReadHoldingRegisters,
					[]uint16{555, 0, 100})
			},
			[]byte{0x03, 0x06, 0x02, 0x2b, 0x00, 0x00, 0x00, 0x64},
		},
		{
			"ReadInputRegistersRequest",
			func(t *testing.T) (encoding.BinaryMarshaler, error) {
				return modbus.NewReadRegisterRequest(modbus.FuncCodeReadInputRegisters, 8, 1)
			},
			[]byte{0x04, 0x00, 0x08, 0x00, 0x01},
		},
		{
			"ReadInputRegistersResponse",
			func(t *testing.T) (encoding.BinaryMarshaler, error) {
				return modbus.NewReadRegisterResponseFromUint16s(modbus.FuncCodeReadInputRegisters,
					[]uint16{10})
			},
			[]byte{0x04, 0x02, 0x00, 0x0a},
		},
		{
			"WriteSingleCoilRequest",
			func(t *testing.T) (encoding.BinaryMarshaler, error) {
				return modbus.NewWriteSingleBitRequest(modbus.FuncCodeWriteSingleCoil, 172, true)
			},
			[]byte{0x05, 0x00, 0xac, 0xff, 0x00},
		},
		{
			"WriteSingleCoilResponse",
			func(t *testing.T) (encoding.BinaryMarshaler, error) {
				return modbus.NewWriteSingleBitResponse(modbus.FuncCodeWriteSingleCoil, 172, true)
			},
			[]byte{0x05, 0x00, 0xac, 0xff, 0x00},
		},
		{
			"WriteSingleRegisterRequest",
			func(t *testing.T) (encoding.BinaryMarshaler, error) {
				return modbus.NewWriteSingleRegisterRequestFromUint16(modbus.FuncCodeWriteSingleRegister, 1, 0x03)
			},
			[]byte{0x06, 0x00, 0x01, 0x00, 0x03},
		},
		{
			"WriteSingleRegisterResponse",
			func(t *testing.T) (encoding.BinaryMarshaler, error) {
				return modbus.NewWriteSingleRegisterResponseFromUint16(modbus.FuncCodeWriteSingleRegister, 1, 0x03)
			},
			[]byte{0x06, 0x00, 0x01, 0x00, 0x03},
		},
		{
			"WriteMultipleCoilsRequest",
			func(t *testing.T) (encoding.BinaryMarshaler, error) {
				return modbus.NewWriteMultipleBitsRequestFromBools(modbus.FuncCodeWriteMultipleCoils, 19,
					[]bool{true, false, true, true, false, false, true, true, true, false})
			},
			[]byte{0x0f, 0x00, 0x13, 0x00, 0x0a, 0x02, 0xcd, 0x01},
		},
		{
			"WriteMultipleCoilsResponse",
			func(t *testing.T) (encoding.BinaryMarshaler, error) {
				return modbus.NewWriteMultipleBitsResponse(modbus.FuncCodeWriteMultipleCoils, 19, 10)
			},
			[]byte{0x0f, 0x00, 0x13, 0x00, 0x0a},
		},
		{
			"WriteMultipleRegistersRequest",
			func(t *testing.T) (encoding.BinaryMarshaler, error) {
				return modbus.NewWriteMultipleRegistersRequestFromUint16s(modbus.FuncCodeWriteMultipleRegisters, 1,
					[]uint16{0x000a, 0x0102})
			},
			[]byte{0x10, 0x00, 0x01, 0x00, 0x02, 0x04, 0x00, 0x0a, 0x01, 0x02},
		},
		{
			"WriteMultipleRegistersResponse",
			func(t *testing.T) (encoding.BinaryMarshaler, error) {
				return modbus.NewWriteMultipleRegistersResponse(modbus.FuncCodeWriteMultipleRegisters, 1, 2)
			},
			[]byte{0x10, 0x00, 0x01, 0x00, 0x02},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			pdu, err := tt.fn(t)
			if err != nil {
				t.Fatal(err)
			}
			data, err := pdu.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(data, tt.expected); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestMarshalUnmarshalBinaryRoundTrip(t *testing.T) {
	testCases := []struct {
		name string
		fn   func(t *testing.T) (modbus.PDU, error)
	}{
		{
			"BasePDU",
			func(t *testing.T) (modbus.PDU, error) {
				return modbus.NewRawPDU([]byte{0x03, 0x00, 0x6b, 0x00, 0x03})
			},
		},
		{
			"ExceptionResponse",
			func(t *testing.T) (modbus.PDU, error) {
				return modbus.NewExceptionResponse(modbus.FuncCodeReadHoldingRegisters+0x80, modbus.ExceptionCodeIllegalFunction)
			},
		},
		{
			"ReadBitRequest",
			func(t *testing.T) (modbus.PDU, error) {
				return modbus.NewReadBitRequest(modbus.FuncCodeReadCoils, 0, 1)
			},
		},
		{
			"ReadBitResponse",
			func(t *testing.T) (modbus.PDU, error) {
				return modbus.NewReadBitResponse(modbus.FuncCodeReadCoils, []byte{0b00010001})
			},
		},
		{
			"ReadBitResponseFromBool",
			func(t *testing.T) (modbus.PDU, error) {
				return modbus.NewReadBitResponseFromBool(modbus.FuncCodeReadCoils, []bool{false})
			},
		},
		{
			"WriteSingleBitRequest",
			func(t *testing.T) (modbus.PDU, error) {
				return modbus.NewWriteSingleBitRequest(modbus.FuncCodeWriteSingleCoil, 3, true)
			},
		},
		{
			"WriteSingleBitResponse",
			func(t *testing.T) (modbus.PDU, error) {
				return modbus.NewWriteSingleBitResponse(modbus.FuncCodeWriteSingleCoil, 3, true)
			},
		},
		{
			"ReadRegisterRequest",
			func(t *testing.T) (modbus.PDU, error) {
				return modbus.NewReadRegisterRequest(modbus.FuncCodeReadHoldingRegisters, 0, 1)
			},
		},
		{
			"ReadRegisterResponse",
			func(t *testing.T) (modbus.PDU, error) {
				return modbus.NewReadRegisterResponse(modbus.FuncCodeReadHoldingRegisters, []byte{0x00, 0x01, 0x00, 0x02})
			},
		},
		{
			"ReadRegisterResponseFromUint16",
			func(t *testing.T) (modbus.PDU, error) {
				return modbus.NewReadRegisterResponseFromUint16s(modbus.FuncCodeReadHoldingRegisters, []uint16{0x0001})
			},
		},
		{
			"WriteSingleRegisterRequest",
			func(t *testing.T) (modbus.PDU, error) {
				return modbus.NewWriteSingleRegisterRequest(modbus.FuncCodeWriteSingleRegister, 3, []byte{0x00, 0x01})
			},
		},
		{
			"WriteSingleRegisterResponse",
			func(t *testing.T) (modbus.PDU, error) {
				return modbus.NewWriteSingleRegisterResponse(modbus.FuncCodeWriteSingleRegister, 3, []byte{0x00, 0x01})
			},
		},
		{
			"WriteMultipleBitsRequest",
			func(t *testing.T) (modbus.PDU, error) {
				return modbus.NewWriteMultipleBitsRequest(modbus.FuncCodeWriteMultipleCoils, 3, 5, []byte{0b00010001})
			},
		},
		{
			"WriteMultipleBitsResponse",
			func(t *testing.T) (modbus.PDU, error) {
				return modbus.NewWriteMultipleBitsResponse(modbus.FuncCodeWriteMultipleCoils, 3, 5)
			},
		},
		{
			"WriteMultipleRegistersRequest",
			func(t *testing.T) (modbus.PDU, error) {
				return modbus.NewWriteMultipleRegistersRequestFromUint16s(modbus.FuncCodeWriteMultipleRegisters, 4,
					[]uint16{0x00, 0x01, 0x00, 0x02})
			},
		},
		{
			"WriteMultipleRegistersResponse",
			func(t *testing.T) (modbus.PDU, error) {
				return modbus.NewWriteMultipleRegistersResponse(modbus.FuncCodeWriteMultipleRegisters, 4, 2)
			},
		},
	}

	optCompareUnexported := cmp.Exporter(func(r reflect.Type) bool {
		return true
	})

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			pdu1, err := tt.fn(t)
			if err != nil {
				t.Fatal(err)
			}

			pduBytes, err := pdu1.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}

			pdu2 := reflect.New(reflect.ValueOf(pdu1).Elem().Type()).Interface().(modbus.PDU)

			if err := pdu2.UnmarshalBinary(pduBytes); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(pdu1, pdu2, optCompareUnexported); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
