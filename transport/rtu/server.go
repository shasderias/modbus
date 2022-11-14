package rtu

//func (t *Client) WriteResponse(slaveAddress byte, r modbus.PDU) error {
//	respFrame := assembleFrame(slaveAddress, r)
//	n, err := t.port.Write(respFrame)
//	if err != nil {
//		return fmt.Errorf("error writing response: %w", err)
//	}
//	if n != len(respFrame) {
//		return fmt.Errorf("short write: %d/%d", n, len(respFrame))
//	}
//	return nil
//}
//
//func (t *Client) ReadRequest() (byte, modbus.PDU, error) {
//	buf := make([]byte, maxFrameLength)
//
//	if _, err := io.ReadFull(t.port, buf[0:2]); err != nil {
//		return 0, nil, fmt.Errorf("error reading request: %w", err)
//	}
//
//	switch buf[1] {
//	case
//		modbus.FuncCodeReadCoils,
//		modbus.FuncCodeReadDiscreteInputs,
//		modbus.FuncCodeReadHoldingRegisters,
//		modbus.FuncCodeReadInputRegisters,
//
//		modbus.FuncCodeWriteSingleCoil,
//		modbus.FuncCodeWriteSingleRegister:
//		// need to read another 2 (startAddress) + 2 (quantity) + 2 (crc) bytes
//		if _, err := io.ReadFull(t.port, buf[2:8]); err != nil {
//			return 0, nil, fmt.Errorf("error reading request: %w", err)
//		}
//
//		buf = buf[:8]
//	case
//		modbus.FuncCodeWriteMultipleCoils,
//		modbus.FuncCodeWriteMultipleRegisters:
//		// need to read another 2 (startAddress) + 2 (quantity) + 1 (byteCount) + byteCount + 2 (crc) bytes
//		// read 5 bytes first to determine how many more bytes to read
//		if _, err := io.ReadFull(t.port, buf[2:7]); err != nil {
//			return 0, nil, fmt.Errorf("error reading request: %w", err)
//		}
//
//		// 1 (slaveAddress) + 1 (functionCode) + 2 (startAddress) + 2 (quantity) + 1 (byteCount) + byteCount (data) + 2 (crc)
//		frameLen := 1 + 1 + 2 + 2 + 1 + int(buf[6]) + 2
//
//		if frameLen > maxFrameLength {
//			return 0, nil, fmt.Errorf("expected frame %d exceeds maximum RTU frame length", frameLen)
//		}
//
//		// read remaining byteCount + 2 bytes
//		if _, err := io.ReadFull(t.port, buf[7:frameLen]); err != nil {
//			return 0, nil, fmt.Errorf("error reading request: %w", err)
//		}
//
//		buf = buf[:frameLen]
//	default:
//		return 0, nil, fmt.Errorf("unsupported function code: %v", buf[0])
//	}
//
//	pdu, err := decodeFrame(buf)
//	if err != nil {
//		return 0, nil, err
//	}
//
//	return buf[0], pdu, nil
//}
