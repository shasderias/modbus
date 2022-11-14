package modbus

import "fmt"

const (
	ExceptionCodeIllegalFunction                    = 0x01
	ExceptionCodeIllegalDataAddress                 = 0x02
	ExceptionCodeIllegalDataValue                   = 0x03
	ExceptionCodeServerDeviceFailure                = 0x04
	ExceptionCodeAcknowledge                        = 0x05
	ExceptionCodeServerDeviceBusy                   = 0x06
	ExceptionCodeMemoryParityError                  = 0x08
	ExceptionCodeGatewayPathUnavailable             = 0x0a
	ExceptionCodeGatewayTargetDeviceFailedToRespond = 0x0b
)

type ExceptionResponse struct {
	errorCode     byte
	exceptionCode byte
}

func NewExceptionResponse(errorCode, exceptionCode byte) (*ExceptionResponse, error) {
	if errorCode <= 0x80 || errorCode > 0xff {
		return nil, fmt.Errorf("modbus: error code out of range (0x80, 0xff]: %v", errorCode)
	}
	if exceptionCode < 1 || exceptionCode > 0xff {
		return nil, fmt.Errorf("modbus: exception code out of range [1, 0xff]: %v", exceptionCode)
	}

	return &ExceptionResponse{errorCode, exceptionCode}, nil
}

func (r *ExceptionResponse) Error() string {
	return fmt.Sprintf("modbus: exception 0x%x:0x%x", r.errorCode, r.exceptionCode)
}

func (r *ExceptionResponse) FunctionCode() byte  { return r.errorCode }
func (r *ExceptionResponse) ExceptionCode() byte { return r.exceptionCode }

func (r *ExceptionResponse) MarshalBinary() ([]byte, error) {
	return []byte{r.errorCode, r.exceptionCode}, nil
}
func (r *ExceptionResponse) UnmarshalBinary(data []byte) error {
	if len(data) != 2 {
		return fmt.Errorf("modbus: exactly 2 bytes required to unmarshal as ExceptionResponse: %v", data)
	}

	resp, err := NewExceptionResponse(data[0], data[1])
	if err != nil {
		return err
	}

	*r = *resp

	return nil
}
