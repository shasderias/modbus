package rtu

import (
	"fmt"
	"io"
	"time"

	"github.com/shasderias/modbus"
)

type Client struct {
	conf *ClientConfig
	port Port
}

type ClientConfig struct {
	RequestTimeout time.Duration

	ResponseHandlers map[byte]responseHandler
}

type responseHandler func(request modbus.PDU, buf []byte, port io.Reader) ([]byte, error)

type Port interface {
	io.ReadWriteCloser
	SetReadDeadline(time.Time) error
}

func NewClient(port Port, cFns ...func(config *ClientConfig)) *Client {
	conf := &ClientConfig{
		RequestTimeout: 300 * time.Millisecond,

		ResponseHandlers: map[byte]responseHandler{
			modbus.FuncCodeReadHoldingRegisters: readResponseHandler,
			modbus.FuncCodeReadInputRegisters:   readResponseHandler,
			modbus.FuncCodeReadCoils:            readResponseHandler,
			modbus.FuncCodeReadDiscreteInputs:   readResponseHandler,

			modbus.FuncCodeWriteSingleRegister:    writeResponseHandler,
			modbus.FuncCodeWriteMultipleRegisters: writeResponseHandler,
			modbus.FuncCodeWriteSingleCoil:        writeResponseHandler,
			modbus.FuncCodeWriteMultipleCoils:     writeResponseHandler,

			0x80 | modbus.FuncCodeReadHoldingRegisters: exceptionResponseHandler,
			0x80 | modbus.FuncCodeReadInputRegisters:   exceptionResponseHandler,
			0x80 | modbus.FuncCodeReadCoils:            exceptionResponseHandler,
			0x80 | modbus.FuncCodeReadDiscreteInputs:   exceptionResponseHandler,

			0x80 | modbus.FuncCodeWriteSingleRegister:    exceptionResponseHandler,
			0x80 | modbus.FuncCodeWriteMultipleRegisters: exceptionResponseHandler,
			0x80 | modbus.FuncCodeWriteSingleCoil:        exceptionResponseHandler,
			0x80 | modbus.FuncCodeWriteMultipleCoils:     exceptionResponseHandler,
		},
	}
	for _, cFn := range cFns {
		cFn(conf)
	}
	return &Client{
		conf: conf,
		port: port,
	}
}

func (c *Client) WriteRequest(slaveAddress byte, r modbus.PDU) (modbus.PDU, error) {
	reqFrame := assembleFrame(slaveAddress, r)

	n, err := c.port.Write(assembleFrame(slaveAddress, r))
	if err != nil {
		return nil, fmt.Errorf("rtu/client: error writing request: %w", err)
	}
	if n != len(reqFrame) {
		return nil, fmt.Errorf("rtu/client: short write: %d/%d", n, len(reqFrame))
	}

	if slaveAddress == 0 {
		return nil, nil
	}

	respFrame := make([]byte, maxFrameLength)

	if err := c.port.SetReadDeadline(time.Now().Add(c.conf.RequestTimeout)); err != nil {
		return nil, err
	}

	// read slave address and function code
	if _, err = io.ReadFull(c.port, respFrame[0:2]); err != nil {
		return nil, fmt.Errorf("rtu/client: error reading response [0:3]: %w", err)
	}

	respSlaveAddress, respFuncCode := respFrame[0], respFrame[1]

	if respSlaveAddress != slaveAddress {
		return nil, fmt.Errorf("rtu/client: unexpected slave address, sent: %d, recv: %d", slaveAddress, respSlaveAddress)
	}

	respHandler, ok := c.conf.ResponseHandlers[respFuncCode]
	if !ok {
		return nil, fmt.Errorf("rtu/client: unsupported response function code: %d", respFuncCode)
	}

	frame, err := respHandler(r, respFrame, c.port)
	if err != nil {
		return nil, err
	}

	return decodeFrame(frame)
}

func (c *Client) Close() error {
	return c.port.Close()
}

func writeResponseHandler(request modbus.PDU, buf []byte, port io.Reader) ([]byte, error) {
	n, err := io.ReadFull(port, buf[2:8])
	if err != nil {
		return nil, fmt.Errorf("rtu/client: error reading response[2:8]: %w", err)
	}
	if n < 6 {
		return nil, fmt.Errorf("rtu/client: short read: %d/6", n)
	}

	return buf[:8], nil
}

func readResponseHandler(request modbus.PDU, buf []byte, port io.Reader) ([]byte, error) {
	n, err := io.ReadFull(port, buf[2:3])
	if err != nil {
		return nil, fmt.Errorf("rtu/client: error reading response[2:3]: %w", err)
	}
	if n < 1 {
		return nil, fmt.Errorf("rtu/client: short read: %d/1", n)
	}

	remainderLength := int(buf[2])
	if 3+remainderLength+2 > maxFrameLength {
		return nil, fmt.Errorf("rtu/client: response length exceeds maximum RTU frame length")
	}

	n, err = io.ReadFull(port, buf[3:3+remainderLength+2])
	if err != nil {
		return nil, fmt.Errorf("rtu/client: error reading response[3:%d]: %w", 3+remainderLength+2, err)
	}
	if n < remainderLength+2 {
		return nil, fmt.Errorf("rtu/client: short read[3:%d]: %d/%d", 3+remainderLength+2, n, remainderLength+2)
	}

	return buf[:3+remainderLength+2], nil
}

func exceptionResponseHandler(request modbus.PDU, buf []byte, port io.Reader) ([]byte, error) {
	n, err := io.ReadFull(port, buf[2:5])
	if err != nil {
		return nil, fmt.Errorf("rtu/client: error reading response[2:3]: %w", err)
	}
	if n < 1 {
		return nil, fmt.Errorf("rtu/client: short read: %d/1", n)
	}
	return buf[:5], nil
}
