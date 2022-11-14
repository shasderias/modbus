package modbus

import (
	"fmt"
	"sync"
)

type Client struct {
	t ClientTransport

	closed    bool
	closedMut sync.Mutex

	slaveAddress byte
}

type ClientTransport interface {
	WriteRequest(slaveAddress byte, r PDU) (PDU, error)
	Close() error
}

func NewClient(slaveAddress int, t ClientTransport) (*Client, error) {
	if slaveAddress < 0 || slaveAddress > 247 {
		return nil, fmt.Errorf("slave address must in the range [0:247]")
	}

	return &Client{
		t: t,

		slaveAddress: byte(slaveAddress),
	}, nil
}

func (c *Client) WriteBit(funcCode byte, startAddress int, value bool) (*WriteSingleBitResponse, error) {
	if c.isClosed() {
		return nil, fmt.Errorf("client: closed")
	}

	req, err := NewWriteSingleBitRequest(funcCode, startAddress, value)
	if err != nil {
		return nil, err
	}

	rawResp, err := c.t.WriteRequest(c.slaveAddress, req)
	if err != nil {
		return nil, fmt.Errorf("client: error writing request: %w", err)
	}

	if c.slaveAddress == 0 {
		return nil, nil
	}

	switch rawResp.FunctionCode() {
	case funcCode:
		var resp WriteSingleBitResponse
		if err := UnmarshalAs(rawResp, &resp); err != nil {
			return nil, err
		}

		if int(resp.startAddress) != startAddress {
			return nil, fmt.Errorf("client: unexpected address, sent: 0x%x, recv: 0x%x",
				startAddress, resp.startAddress)
		}
		return &resp, nil
	case funcCode + 0x80:
		var resp ExceptionResponse
		if err := UnmarshalAs(rawResp, &resp); err != nil {
			return nil, err
		}
		return nil, &resp
	default:
		return nil, fmt.Errorf("client: unexpected function code, sent: 0x%x, recv: 0x%x",
			funcCode, rawResp.FunctionCode())
	}
}

func (c *Client) WriteSingleCoil(address int, value bool) (*WriteSingleBitResponse, error) {
	return c.WriteBit(FuncCodeWriteSingleCoil, address, value)
}

func (c *Client) WriteBits(funcCode byte, startAddress int, values []bool) (*WriteMultipleBitsResponse, error) {
	if c.isClosed() {
		return nil, fmt.Errorf("client: closed")
	}

	req, err := NewWriteMultipleBitsRequestFromBools(funcCode, startAddress, values)
	if err != nil {
		return nil, err
	}

	rawResp, err := c.t.WriteRequest(c.slaveAddress, req)
	if err != nil {
		return nil, fmt.Errorf("client: error writing request: %w", err)
	}

	if c.slaveAddress == 0 {
		return nil, nil
	}

	switch rawResp.FunctionCode() {
	case funcCode:
		var resp WriteMultipleBitsResponse
		if err := UnmarshalAs(rawResp, &resp); err != nil {
			return nil, err
		}

		if int(resp.startAddress) != startAddress {
			return nil, fmt.Errorf("client: response start address (%d) does not match request start address (%d)",
				resp.startAddress, startAddress)
		}
		if int(resp.count) != len(values) {
			return nil, fmt.Errorf("client: response register count (%d) does not match request register count (%d)",
				resp.count, len(values))
		}

		return &resp, nil
	case funcCode + 0x80:
		var resp ExceptionResponse
		if err := UnmarshalAs(rawResp, &resp); err != nil {
			return nil, err
		}
		return nil, &resp
	default:
		return nil, fmt.Errorf("client: unexpected function code: %d", rawResp.FunctionCode())
	}
}

func (c *Client) WriteMultipleCoils(startAddress int, values []bool) (*WriteMultipleBitsResponse, error) {
	return c.WriteBits(FuncCodeWriteMultipleCoils, startAddress, values)
}

func (c *Client) ReadBits(funcCode byte, startAddress, count int) (*ReadBitResponse, error) {
	if c.isClosed() {
		return nil, fmt.Errorf("client: closed")
	}

	req, err := NewReadBitRequest(funcCode, startAddress, count)
	if err != nil {
		return nil, err
	}

	rawResp, err := c.t.WriteRequest(c.slaveAddress, req)
	if err != nil {
		return nil, fmt.Errorf("client: error writing request: %w", err)
	}

	switch rawResp.FunctionCode() {
	case funcCode:
		var resp ReadBitResponse
		if err := UnmarshalAs(rawResp, &resp); err != nil {
			return nil, err
		}
		return &resp, nil
	case funcCode + 0x80:
		var resp ExceptionResponse
		if err := UnmarshalAs(rawResp, &resp); err != nil {
			return nil, err
		}
		return nil, &resp
	default:
		return nil, fmt.Errorf("client: unexpected function code: %d", rawResp.FunctionCode())
	}
}

func (c *Client) ReadCoils(startAddress, count int) (*ReadBitResponse, error) {
	return c.ReadBits(FuncCodeReadCoils, startAddress, count)
}

func (c *Client) ReadDiscreteInputs(startAddress, count int) (*ReadBitResponse, error) {
	return c.ReadBits(FuncCodeReadDiscreteInputs, startAddress, count)
}

func (c *Client) WriteRegister(funcCode byte, address int, value uint16) (*WriteSingleRegisterResponse, error) {
	if c.isClosed() {
		return nil, fmt.Errorf("client: closed")
	}

	req, err := NewWriteSingleRegisterRequestFromUint16(FuncCodeWriteSingleRegister, address, value)
	if err != nil {
		return nil, err
	}

	rawResp, err := c.t.WriteRequest(c.slaveAddress, req)
	if err != nil {
		return nil, fmt.Errorf("client: error writing request: %w", err)
	}

	if c.slaveAddress == 0 {
		return nil, nil
	}

	switch rawResp.FunctionCode() {
	case funcCode:
		var resp WriteSingleRegisterResponse
		if err := UnmarshalAs(rawResp, &resp); err != nil {
			return nil, err
		}

		if int(resp.address) != address {
			return nil, fmt.Errorf("client: response start address (%d) does not match request start address (%d)",
				resp.address, address)
		}

		return &resp, nil
	case funcCode + 0x80:
		var resp ExceptionResponse
		if err := UnmarshalAs(rawResp, &resp); err != nil {
			return nil, err
		}
		return nil, &resp
	default:
		return nil, fmt.Errorf("client: unexpected function code: %d", rawResp.FunctionCode())
	}
}

func (c *Client) WriteSingleRegister(address int, value uint16) (*WriteSingleRegisterResponse, error) {
	return c.WriteRegister(FuncCodeWriteSingleRegister, address, value)
}

func (c *Client) WriteRegisters(startAddress int, values []uint16) (*WriteMultipleRegistersResponse, error) {
	if c.isClosed() {
		return nil, fmt.Errorf("client: closed")
	}

	req, err := NewWriteMultipleRegistersRequestFromUint16s(FuncCodeWriteMultipleRegisters, startAddress, values)
	if err != nil {
		return nil, err
	}

	rawResp, err := c.t.WriteRequest(c.slaveAddress, req)
	if err != nil {
		return nil, fmt.Errorf("client: error writing request: %w", err)
	}

	if c.slaveAddress == 0 {
		return nil, nil
	}

	switch rawResp.FunctionCode() {
	case FuncCodeWriteMultipleRegisters:
		var resp WriteMultipleRegistersResponse
		if err := UnmarshalAs(rawResp, &resp); err != nil {
			return nil, err
		}

		if int(resp.address) != startAddress {
			return nil, fmt.Errorf("client: response start address (%d) does not match request start address (%d)",
				resp.address, startAddress)
		}
		if int(resp.count) != len(values) {
			return nil, fmt.Errorf("client: response register count (%d) does not match request register count (%d)",
				resp.count, len(values))
		}

		return &resp, nil
	case FuncCodeWriteMultipleRegisters + 0x80:
		var resp ExceptionResponse
		if err := UnmarshalAs(rawResp, &resp); err != nil {
			return nil, err
		}
		return nil, &resp
	default:
		return nil, fmt.Errorf("client: unexpected function code: %d", rawResp.FunctionCode())
	}
}

func (c *Client) ReadRegisters(funcCode byte, startAddress, count int) (*ReadRegisterResponse, error) {
	if c.isClosed() {
		return nil, fmt.Errorf("client: closed")
	}

	req, err := NewReadRegisterRequest(int(funcCode), startAddress, count)
	if err != nil {
		return nil, err
	}

	rawResp, err := c.t.WriteRequest(c.slaveAddress, req)
	if err != nil {
		return nil, fmt.Errorf("client: error writing request: %w", err)
	}

	switch rawResp.FunctionCode() {
	case funcCode:
		var resp ReadRegisterResponse
		if err := UnmarshalAs(rawResp, &resp); err != nil {
			return nil, err
		}
		return &resp, nil
	case funcCode + 0x80:
		var resp ExceptionResponse
		if err := UnmarshalAs(rawResp, &resp); err != nil {
			return nil, err
		}
		return nil, &resp
	default:
		return nil, fmt.Errorf("client: unexpected function code: %d", rawResp.FunctionCode())
	}
}

func (c *Client) ReadInputRegisters(startAddress, count int) (*ReadRegisterResponse, error) {
	return c.ReadRegisters(FuncCodeReadInputRegisters, startAddress, count)
}

func (c *Client) ReadHoldingRegisters(startAddress, count int) (*ReadRegisterResponse, error) {
	return c.ReadRegisters(FuncCodeReadHoldingRegisters, startAddress, count)
}

func (c *Client) Close() error {
	c.closedMut.Lock()
	defer c.closedMut.Unlock()
	c.closed = true
	return c.t.Close()
}

func (c *Client) isClosed() bool {
	c.closedMut.Lock()
	defer c.closedMut.Unlock()
	return c.closed
}
