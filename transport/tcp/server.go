package tcp

//import (
//	"encoding/binary"
//	"errors"
//	"fmt"
//	"io"
//	"net"
//
//	"github.com/shasderias/modbus"
//)
//
//type Server struct {
//	address string
//
//	l net.Listener
//
//	h handlerFunc
//}
//
//type ServerConfig struct {
//}
//
//type handlerFunc func(request modbus.PDU) (modbus.PDU, error)
//
//func NewServer(address string) (*Server, error) {
//	server := &Server{address: address}
//
//	return server, nil
//}
//
//func (s *Server) Start() error {
//	if s.l != nil {
//		return fmt.Errorf("tcp/server: already started")
//	}
//
//	listener, err := net.Listen("tcp", s.address)
//	if err != nil {
//		return fmt.Errorf("tcp/server: error listening: %w", err)
//	}
//
//	s.l = listener
//
//	go s.acceptLoop()
//
//	return nil
//}
//
//func (s *Server) RegisterHandler(h handlerFunc) {
//	s.h = h
//}
//
//func (s *Server) Stop() error {
//	err := s.l.Close()
//
//	s.l = nil
//
//	return err
//}
//
//func (s *Server) acceptLoop() {
//	for {
//		conn, err := s.l.Accept()
//		if errors.Is(err, net.ErrClosed) {
//			return
//		} else if err != nil {
//			s.log(err)
//		}
//		go s.handleConn(conn)
//	}
//}
//
//func (s *Server) handleConn(conn net.Conn) {
//	defer conn.Close()
//
//	for {
//		err := func() (err error) {
//			buf := make([]byte, maxFrameSize)
//
//			n, err := io.ReadFull(conn, buf[:6])
//			if err != nil {
//				return fmt.Errorf("tcp/server: error reading [:6]: %w", err)
//			}
//			if n < 6 {
//				return fmt.Errorf("tcp/server: short read [:6]: %d/6", n)
//			}
//
//			protoID := binary.BigEndian.Uint16(buf[2:4])
//			if protoID != 0 {
//				return fmt.Errorf("tcp/server: unexpected protocol id: %d", protoID)
//			}
//
//			remainderLen := binary.BigEndian.Uint16(buf[4:6])
//
//			if remainderLen < 3 {
//				return fmt.Errorf("tcp/server: short frame, expected frame to be at least 9 bytes long: %d", 6+remainderLen)
//			}
//			if remainderLen > maxFrameSize-6 {
//				return fmt.Errorf("tcp/server: frame too long: %d", 6+remainderLen)
//			}
//
//			n, err = io.ReadFull(conn, buf[6:6+remainderLen])
//			if err != nil {
//				return fmt.Errorf("tcp/server: error reading [6:]: %w", err)
//			}
//			if n < int(remainderLen) {
//				return fmt.Errorf("tcp/server: short read [6:]: %d/%d", n, remainderLen)
//			}
//
//			txID := binary.BigEndian.Uint16(buf[0:2])
//			unitID := buf[6]
//
//			request, err := modbus.NewRawPDU(buf[7 : 6+remainderLen])
//			if err != nil {
//				return fmt.Errorf("tcp/server: error decoding request PDU: %w", err)
//			}
//
//			response, err := s.h(request)
//			if err != nil {
//				return fmt.Errorf("tcp/server: error handling request: %w", err)
//			}
//
//			_, err = conn.Write(assembleFrame(txID, unitID, response))
//			if err != nil {
//				return fmt.Errorf("tcp/server: error writing response: %w", err)
//			}
//
//			return nil
//		}()
//		if errors.Is(err, net.ErrClosed) {
//			return
//		} else if err != nil {
//			s.log(err)
//		}
//	}
//}
//
//func (*Server) log(a ...any) {
//	fmt.Println(a...)
//}
