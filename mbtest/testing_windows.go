package mbtest

import (
	"net"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/shasderias/serial"

	"github.com/shasderias/modbus/internal/projectroot"
)

const (
	defaultLoopbackPort1 = "COM5"
	defaultLoopbackPort2 = "COM6"
)

func envString(name, value string) string {
	envValue, ok := os.LookupEnv(name)
	if !ok {
		return value
	}
	return envValue
}

func StartDiagSlaveTCP(t *testing.T) (net.Conn, func()) {
	exePath := path.Join(projectroot.Get(), "reference", "diagslave-3.4", "win", "diagslave.exe")

	cmd := exec.Command(exePath, "-p", "5502")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		cmd.Process.Kill()
	})

	time.Sleep(1 * time.Second)

	conn, err := net.DialTimeout("tcp", ":5502", 1*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		cmd.Process.Kill()
	}

	return conn, cleanup
}

func StartDiagSlaveRTU(t *testing.T) serial.Port {
	loopbackPort1 := envString("MBTEST_LOOPBACK_PORT", defaultLoopbackPort1)
	loopbackPort2 := envString("MBTEST_LOOPBACK_PORT", defaultLoopbackPort2)

	exePath := path.Join(projectroot.Get(), "reference", "diagslave-3.4", "win", "diagslave.exe")

	cmd := exec.Command(exePath, "-m", "rtu", loopbackPort1)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		cmd.Process.Kill()
	})

	time.Sleep(1 * time.Second)

	port, err := serial.Open(loopbackPort2, func(c *serial.Config) {
		c.BaudRate = 19200
		c.DataBits = 8
		c.StopBits = serial.StopBits1
		c.Parity = serial.ParityEven
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		port.Close()
	})

	return port
}
