package mbtest

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/shasderias/modbus"
	"github.com/shasderias/modbus/transport/rtu"
	"github.com/shasderias/modbus/transport/tcp"
)

func TestRTUClient(t *testing.T) {
	port := StartDiagSlaveRTU(t)

	client, err := modbus.NewClient(1, rtu.NewClient(port))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	testClient(t, client)
}

func TestTCPClient(t *testing.T) {
	conn := StartDiagSlaveTCP(t)

	transport, err := tcp.NewClient(conn)
	if err != nil {
		t.Fatal(err)
	}

	client, err := modbus.NewClient(1, transport)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	testClient(t, client)
}

func testClient(t *testing.T, client *modbus.Client) {
	{ // verify coil initial state
		coils, err := client.ReadCoils(0, 16)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(coils.BitValues(),
			[]bool{
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
			}); diff != "" {
			t.Fatal(diff)
		}
		if diff := cmp.Diff(coils.Values(), []byte{0x00, 0x00}); diff != "" {
			t.Fatal(diff)
		}
	}

	{ // verify holding registers initial state
		holdingRegisters, err := client.ReadHoldingRegisters(0, 16)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(holdingRegisters.Uint16(),
			[]uint16{
				0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0,
			}); diff != "" {
			t.Fatal(diff)
		}
		if diff := cmp.Diff(holdingRegisters.Values(), []byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		}); diff != "" {
			t.Fatal(diff)
		}
	}

	if _, err := client.WriteSingleCoil(0, true); err != nil {
		t.Fatal(err)
	}
	if _, err := client.WriteMultipleCoils(7, []bool{true, false, true}); err != nil {
		t.Fatal(err)
	}

	{ // verify that coils were correctly written to
		coils, err := client.ReadCoils(0, 16)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(coils.BitValues(),
			[]bool{
				true, false, false, false, false, false, false, true,
				false, true, false, false, false, false, false, false,
			}); diff != "" {
			t.Fatal(diff)
		}
		if diff := cmp.Diff(coils.Values(), []byte{0b10000001, 0b00000010}); diff != "" {
			t.Fatal(diff)
		}
	}

	if _, err := client.WriteSingleRegister(15, 65535); err != nil {
		t.Fatal(err)
	}
	if _, err := client.WriteRegisters(7, []uint16{4, 2333, 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.WriteSingleRegister(0, 65535); err != nil {
		t.Fatal(err)
	}

	{ // verify that holding registers were correctly written to
		holdingRegisters, err := client.ReadHoldingRegisters(0, 16)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(holdingRegisters.Uint16(),
			[]uint16{
				65535, 0, 0, 0, 0, 0, 0, 4,
				2333, 1, 0, 0, 0, 0, 0, 65535,
			}); diff != "" {
			t.Fatal(diff)
		}
		if diff := cmp.Diff(holdingRegisters.Values(), []byte{
			0xff, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04,
			0x09, 0x1d, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff,
		}); diff != "" {
			t.Fatal(diff)
		}
	}
}
