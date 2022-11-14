package mbtest

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/shasderias/serial"

	"github.com/shasderias/modbus/internal/projectroot"
)

func StartDiagSlaveTCP(t *testing.T) net.Conn {
	exePath := path.Join(projectroot.Get(), "reference", "diagslave-3.4", "x86_64-linux-gnu", "diagslave")

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

	return conn
}

func StartDiagSlaveRTU(t *testing.T) serial.Port {
	port1, port2 := setupLoopbackPorts(t)

	exePath := path.Join(projectroot.Get(), "reference", "diagslave-3.4", "x86_64-linux-gnu", "diagslave")

	cmd := exec.Command(exePath, "-m", "rtu", port1)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		cmd.Process.Kill()
	})

	time.Sleep(1 * time.Second)

	port, err := serial.Open(port2, &serial.Config{
		BaudRate: 19200,
		DataBits: 8,
		StopBits: serial.StopBits1,
		Parity:   serial.ParityEven,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		port.Close()
	})

	return port
}

func startSocat(t *testing.T, args ...string) {
	_, err := exec.LookPath("socat")
	if err != nil {
		t.Skip("socat not found in path")
		return
	}

	cmd := exec.Command("socat", append([]string{"-D"}, args...)...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			t.Logf("error killing socat: %v", err)
		}
		switch err := cmd.Wait().(type) {
		case *exec.ExitError:
			if err.ExitCode() != 130 {
				t.Log(err)
			}
		default:
			t.Log(err)
		}
	})

	// when socat writes to stderr (because of the -D flag), it is ready
	buf := make([]byte, 1024)
	if _, err := stderr.Read(buf); err != nil {
		t.Fatal(err)
	}
}

func setupLoopbackPorts(t *testing.T) (string, string) {
	var (
		tempDir = t.TempDir()

		path1 = path.Join(tempDir, "port1")
		path2 = path.Join(tempDir, "port2")

		port1Def = fmt.Sprintf("pty,raw,echo=0,link=%s", path1)
		port2Def = fmt.Sprintf("pty,raw,echo=0,link=%s", path2)
	)

	startSocat(t, port1Def, port2Def)

	return path1, path2
}
