package rtu_test

import (
	"fmt"
	"testing"

	"github.com/shasderias/modbus"
	"github.com/shasderias/modbus/mbtest"
	"github.com/shasderias/modbus/transport/rtu"
)

func logfAsErr(t *testing.T, format string, args ...any) error {
	t.Helper()
	t.Logf(format, args...)
	return fmt.Errorf(format, args...)
}

//	func TestReadWriteSanity(t *testing.T) {
//		const (
//			startAddress  = 3
//			registerCount = 2
//			register3     = 0x34
//			register4     = 0x12
//		)
//		var (
//			port1, port2 = net.Pipe()
//
//			master = rtu.NewClient(port1)
//			slave  = rtu.NewClient(port2)
//
//			wg   sync.WaitGroup
//			done = make(chan struct{})
//		)
//
//		wg.Add(2)
//
//		go func() {
//			err := func() error {
//				req, err := modbus.NewReadRegisterRequest(modbus.FuncCodeReadHoldingRegisters, startAddress, registerCount)
//				if err != nil {
//					return logfAsErr(t, "error creating request: %v", err)
//				}
//
//				pdu, err := master.WriteRequest(1, req)
//				if err != nil {
//					return logfAsErr(t, "error writing request: %v", err)
//				}
//
//				if pdu.FunctionCode() != modbus.FuncCodeReadHoldingRegisters {
//					return logfAsErr(t, "got %d, want %d", pdu.FunctionCode(), modbus.FuncCodeReadHoldingRegisters)
//				}
//
//				var resp modbus.ReadRegisterResponse
//				if err := modbus.UnmarshalAs(pdu, &resp); err != nil {
//					return logfAsErr(t, "error decoding response: %v", err)
//				}
//
//				if resp.Uint16()[0] != register3 {
//					return logfAsErr(t, "got %d; want %d; pdu: %v; resp: %v", resp.Uint16()[0], register3, pdu, resp)
//				}
//				if resp.Uint16()[1] != register4 {
//					return logfAsErr(t, "got %d; want %d; pdu: %v; resp: %v", resp.Uint16()[1], register4, pdu, resp)
//				}
//
//				return nil
//			}()
//			if err != nil {
//				t.Fail()
//			}
//			wg.Done()
//		}()
//
//		go func() {
//			err := func() error {
//				slaveID, pdu, err := slave.ReadRequest()
//				if err != nil {
//					t.Log(err)
//				}
//
//				if slaveID != 1 {
//					return logfAsErr(t, "got %d; want %d; pdu: %v", slaveID, 1, pdu)
//				}
//				if pdu.FunctionCode() != modbus.FuncCodeReadHoldingRegisters {
//					return logfAsErr(t, "got %d; want %d; pdu: %v", pdu.FunctionCode(), modbus.FuncCodeReadHoldingRegisters, pdu)
//				}
//
//				var req modbus.ReadRegisterRequest
//				if err := modbus.UnmarshalAs(pdu, &req); err != nil {
//					return logfAsErr(t, "error decoding request: %v", err)
//				}
//
//				if req.StartAddress() != startAddress {
//					return logfAsErr(t, "got %d; want %d", req.StartAddress(), startAddress)
//				}
//				if req.RegisterCount() != registerCount {
//					return logfAsErr(t, "got %d; want %d", req.RegisterCount(), registerCount)
//				}
//
//				resp, err := modbus.NewReadRegisterResponseFromUint16s(modbus.FuncCodeReadHoldingRegisters, []uint16{register3, register4})
//				if err != nil {
//					return logfAsErr(t, "error creating response: %v", err)
//				}
//
//				if err := slave.WriteResponse(slaveID, resp); err != nil {
//					return logfAsErr(t, "error writing response: %v", err)
//				}
//
//				return nil
//			}()
//			if err != nil {
//				t.Fail()
//			}
//
//			wg.Done()
//		}()
//
//		go func() {
//			wg.Wait()
//			done <- struct{}{}
//		}()
//
//		select {
//		case <-time.NewTimer(3 * time.Second).C:
//			t.Fatal("timeout")
//		case <-done:
//		}
//	}
//
//	func TestReadWrite(t *testing.T) {
//		testCases := []struct {
//			name          string
//			reqFn, respFn func() (modbus.PDU, error)
//		}{
//			{
//				"ReadBit",
//				func() (modbus.PDU, error) {
//					return modbus.NewReadBitRequest(modbus.FuncCodeReadCoils, 3, 2)
//				},
//				func() (modbus.PDU, error) {
//					return modbus.NewReadBitResponse(modbus.FuncCodeReadCoils, []byte{0b00000011})
//				},
//			},
//			{
//				"ReadBitFromBool",
//				func() (modbus.PDU, error) {
//					return modbus.NewReadBitRequest(modbus.FuncCodeReadCoils, 3, 2)
//				},
//				func() (modbus.PDU, error) {
//					return modbus.NewReadBitResponseFromBool(modbus.FuncCodeReadCoils, []bool{true, false})
//				},
//			},
//			{
//				"ReadBitException",
//				func() (modbus.PDU, error) {
//					return modbus.NewReadBitRequest(modbus.FuncCodeReadCoils, 3, 2)
//				},
//				func() (modbus.PDU, error) {
//					return modbus.NewExceptionResponse(modbus.FuncCodeReadCoils+0x80, modbus.ExceptionCodeIllegalFunction)
//				},
//			},
//			{
//				"WriteSingleBit",
//				func() (modbus.PDU, error) {
//					return modbus.NewWriteSingleBitRequest(modbus.FuncCodeWriteSingleCoil, 3, true)
//				},
//				func() (modbus.PDU, error) {
//					return modbus.NewWriteSingleBitResponse(modbus.FuncCodeWriteSingleCoil, 3, true)
//				},
//			},
//			{
//				"WriteSingleBitException",
//				func() (modbus.PDU, error) {
//					return modbus.NewWriteSingleBitRequest(modbus.FuncCodeWriteSingleCoil, 3, true)
//				},
//				func() (modbus.PDU, error) {
//					return modbus.NewExceptionResponse(modbus.FuncCodeWriteSingleCoil+0x80, modbus.ExceptionCodeIllegalFunction)
//				},
//			},
//			{
//				"ReadRegister",
//				func() (modbus.PDU, error) {
//					return modbus.NewReadRegisterRequest(modbus.FuncCodeReadHoldingRegisters, 3, 2)
//				},
//				func() (modbus.PDU, error) {
//					return modbus.NewReadRegisterResponse(modbus.FuncCodeReadHoldingRegisters, []byte{0x00, 0x03, 0x00, 0x04})
//				},
//			},
//			{
//				"ReadRegisterFromUint16",
//				func() (modbus.PDU, error) {
//					return modbus.NewReadRegisterRequest(modbus.FuncCodeReadHoldingRegisters, 3, 2)
//				},
//				func() (modbus.PDU, error) {
//					return modbus.NewReadRegisterResponseFromUint16s(modbus.FuncCodeReadHoldingRegisters, []uint16{3, 4})
//				},
//			},
//			{
//				"WriteSingleRegister",
//				func() (modbus.PDU, error) {
//					return modbus.NewWriteSingleRegisterRequest(modbus.FuncCodeWriteSingleRegister, 3, []byte{0x00, 0x04})
//				},
//				func() (modbus.PDU, error) {
//					return modbus.NewWriteSingleRegisterResponse(modbus.FuncCodeWriteSingleRegister, 3, []byte{0x00, 0x04})
//				},
//			},
//		}
//
//		optCompareUnexported := cmp.Exporter(func(r reflect.Type) bool {
//			return true
//		})
//
//		for _, tt := range testCases {
//
//			t.Run(tt.name, func(t *testing.T) {
//				var (
//					port1, port2 = net.Pipe()
//
//					master = rtu.NewClient(port1)
//					slave  = rtu.NewClient(port2)
//
//					wg   sync.WaitGroup
//					done = make(chan struct{})
//				)
//
//				wg.Add(2)
//
//				modelReq, err := tt.reqFn()
//				if err != nil {
//					t.Fatal(err)
//				}
//				modelResp, err := tt.respFn()
//				if err != nil {
//					t.Fatal(err)
//				}
//
//				go func() {
//					err := func() error {
//						pdu, err := master.WriteRequest(1, modelReq)
//						if err != nil {
//							return logfAsErr(t, "error writing request: %v", err)
//						}
//
//						resp := reflect.New(reflect.ValueOf(modelResp).Elem().Type()).Interface().(modbus.PDU)
//						if err := modbus.UnmarshalAs(pdu, resp); err != nil {
//							return logfAsErr(t, "error decoding response: %v", err)
//						}
//
//						if diff := cmp.Diff(modelResp, resp, optCompareUnexported); diff != "" {
//							return logfAsErr(t, "%s", diff)
//						}
//
//						return nil
//					}()
//					if err != nil {
//						t.Fail()
//					}
//					wg.Done()
//				}()
//
//				go func() {
//					err := func() error {
//						slaveID, pdu, err := slave.ReadRequest()
//						if err != nil {
//							t.Log(err)
//						}
//
//						req := reflect.New(reflect.ValueOf(modelReq).Elem().Type()).Interface().(modbus.PDU)
//
//						if err := modbus.UnmarshalAs(pdu, req); err != nil {
//							return logfAsErr(t, "error decoding request: %v", err)
//						}
//
//						if diff := cmp.Diff(modelReq, req, optCompareUnexported); diff != "" {
//							return logfAsErr(t, "%s", diff)
//						}
//
//						if err := slave.WriteResponse(slaveID, modelResp); err != nil {
//							return logfAsErr(t, "error writing response: %v", err)
//						}
//
//						return nil
//					}()
//					if err != nil {
//						t.Fail()
//					}
//
//					wg.Done()
//				}()
//
//				go func() {
//					wg.Wait()
//					done <- struct{}{}
//				}()
//
//				select {
//				case <-time.NewTimer(3 * time.Second).C:
//					t.Fatal("timeout")
//				case <-done:
//				}
//			})
//		}
//	}
func TestDiagSlave(t *testing.T) {
	port := mbtest.StartDiagSlaveRTU(t)
	defer port.Close()

	transport := rtu.NewClient(port)

	writeReq, err := modbus.NewWriteMultipleRegistersRequestFromUint16s(
		modbus.FuncCodeWriteMultipleRegisters, 1, []uint16{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	if err != nil {
		t.Fatal(err)
	}

	writeResp, err := transport.WriteRequest(1, writeReq)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(writeResp)

	req, err := modbus.NewReadRegisterRequest(modbus.FuncCodeReadHoldingRegisters, 0, 12)
	if err != nil {
		t.Fatal(err)
	}

	respPDU, err := transport.WriteRequest(1, req)
	if err != nil {
		t.Fatal(err)
	}

	switch respPDU.FunctionCode() {
	case modbus.FuncCodeReadHoldingRegisters:
		var resp modbus.ReadRegisterResponse
		if err := modbus.UnmarshalAs(respPDU, &resp); err != nil {
			t.Fatal(err)
		}
		t.Log(resp.Uint16())
	case modbus.FuncCodeReadHoldingRegisters + 0x80:
		t.Fatalf("unexpected exception response: %v", respPDU)
	default:
		t.Fatalf("unexpected function code: %v", respPDU.FunctionCode())
	}
}
