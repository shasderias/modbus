package tcp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/shasderias/modbus"
)

type request struct {
	txID      uint16
	unitID    byte
	req, resp modbus.PDU
	err       error
	done      chan *request
}

type Client struct {
	c Conn

	closed    bool
	closedMut sync.Mutex

	requestTimeout time.Duration

	writeLoopDone       chan struct{}
	requestQueue        chan *request
	inflightRequestsMut sync.Mutex
	inflightRequests    map[uint16]*request

	txID      uint16
	txIDMutex sync.Mutex
}

type ClientConfig struct {
	RequestTimeout time.Duration
}

func NewClient(c Conn, fns ...func(c *ClientConfig)) (*Client, error) {
	config := ClientConfig{
		RequestTimeout: 500 * time.Millisecond,
	}

	for _, fn := range fns {
		fn(&config)
	}

	client := &Client{
		c: c,

		requestTimeout: config.RequestTimeout,

		writeLoopDone:    make(chan struct{}),
		requestQueue:     make(chan *request),
		inflightRequests: make(map[uint16]*request),
	}

	go client.readLoop()
	go client.writeLoop()

	return client, nil
}

func (c *Client) readLoop() {
	for {
		err := func() (err error) {
			var req *request
			defer func() {
				if err != nil && req != nil {
					req.err = err
					req.done <- req
				}
			}()
			buf := make([]byte, maxFrameSize)

			n, err := io.ReadFull(c.c, buf[:7])
			if err != nil {
				return fmt.Errorf("modbus/tcp: error reading [:7]: %w", err)
			}
			if n < 7 {
				return fmt.Errorf("modbus/tcp: short read [:7]: %d/7", n)
			}

			txID := binary.BigEndian.Uint16(buf[:2])

			c.inflightRequestsMut.Lock()
			req = c.inflightRequests[txID]
			c.inflightRequestsMut.Unlock()

			if req == nil {
				return fmt.Errorf("modbus/tcp: unexpected transaction ID: %d", txID)
			}
			if protocolID := binary.BigEndian.Uint16(buf[2:4]); protocolID != 0 {
				return fmt.Errorf("modbus/tcp: unexpected protocol ID: %d", protocolID)
			}
			if unitID := buf[6]; unitID != req.unitID {
				return fmt.Errorf("modbus/tcp: unexpected unit ID: %d", unitID)
			}

			remainingBytes := binary.BigEndian.Uint16(buf[4:6])

			if remainingBytes < 2 {
				return fmt.Errorf("modbus/tcp: short frame, expected frame to be at least 9 bytes long: %d", 6+remainingBytes)
			}
			if remainingBytes > maxFrameSize-7 {
				return fmt.Errorf("modbus/tcp: frame too long: %d", 6+remainingBytes)
			}

			n, err = io.ReadFull(c.c, buf[7:6+remainingBytes])
			if err != nil {
				return fmt.Errorf("modbus/tcp: error reading [7:%d]: %w", 6+remainingBytes, err)
			}
			if n < int(remainingBytes)-1 {
				return fmt.Errorf("modbus/tcp: short read [7:%d]: %d/%d", 6+remainingBytes, n, remainingBytes)
			}

			req.resp, err = modbus.NewRawPDU(buf[7 : 6+remainingBytes])
			if err != nil {
				return fmt.Errorf("modbus/tcp: error parsing PDU: %w", err)
			}

			req.done <- req

			return nil
		}()
		if err != nil {
			c.log(err.Error())
			c.Close()
			return
		}
	}
}

func (c *Client) log(msg string) {
	fmt.Println(msg)
}

func (c *Client) writeLoop() {
	for {
		select {
		case r := <-c.requestQueue:
			frame := assembleFrame(r.txID, r.unitID, r.req)

			n, err := c.c.Write(frame)
			if err != nil {
				r.err = fmt.Errorf("modbus/tcp: error writing request: %w", err)
				r.done <- r

				if errors.Is(err, net.ErrClosed) {
					return
				}
				continue
			}
			if n != len(frame) {
				r.err = fmt.Errorf("modbus/tcp: short write: %d/%d", n, len(frame))
				r.done <- r
				continue
			}
		case <-c.writeLoopDone:
			break
		}
	}
}

func (c *Client) queueRequest(unitID byte, requestPDU modbus.PDU) *request {
	txID := c.getTxID()

	r := request{
		txID:   txID,
		unitID: unitID,
		req:    requestPDU,
		done:   make(chan *request),
	}

	c.inflightRequestsMut.Lock()
	c.inflightRequests[txID] = &r
	c.inflightRequestsMut.Unlock()

	c.requestQueue <- &r

	return &r
}

func (c *Client) WriteRequest(unitID byte, r modbus.PDU) (modbus.PDU, error) {
	if c.isClosed() {
		return nil, fmt.Errorf("modbus/tcp: client closed")
	}
	result := c.queueRequest(unitID, r)

	select {
	case <-result.done:
		return result.resp, result.err
	case <-time.After(c.requestTimeout):
		c.inflightRequestsMut.Lock()
		delete(c.inflightRequests, result.txID)
		c.inflightRequestsMut.Unlock()
		return nil, fmt.Errorf("modbus/tcp: timeout waiting for response")
	}
}

func (c *Client) Close() error {
	c.closedMut.Lock()
	defer c.closedMut.Unlock()
	if !c.closed {
		close(c.writeLoopDone)
	}
	c.closed = true
	return c.c.Close()
}

func (c *Client) isClosed() bool {
	c.closedMut.Lock()
	defer c.closedMut.Unlock()
	return c.closed
}

func (c *Client) getTxID() uint16 {
	c.txIDMutex.Lock()
	defer c.txIDMutex.Unlock()

	c.txID++
	return c.txID
}
