package crc_test

import (
	"fmt"
	"testing"

	"github.com/shasderias/modbus/internal/crc"
)

var (
	testCases = []struct {
		data []byte
		want uint16
	}{
		{
			[]byte{},
			0xffff,
		},
		{
			[]byte{0x02, 0x07},
			0x1241,
		},
		{

			[]byte{0x01, 0x04, 0x02, 0xff, 0xff},
			0x80b8,
		},
		{
			[]byte("123456789"),
			0x4b37,
		},
		{
			[]byte("the quick brown fox jumps over the lazy dog"),
			0x10c2,
		},
	}
)

func TestCRC(t *testing.T) {
	for i, tt := range testCases {
		t.Run(fmt.Sprintf("%d", i+1), func(t *testing.T) {
			if c := crc.Checksum(tt.data); c != tt.want {
				t.Fatalf("got %v; want %v", c, tt.want)
			}
		})
	}
}

func TestIncremental(t *testing.T) {
	const want = 0x1241

	h := crc.New()
	h.Write([]byte{0x02})
	h.Write([]byte{0x07})

	if c := h.Checksum(); c != want {
		t.Fatalf("got %v; want %v", c, want)
	}
}
