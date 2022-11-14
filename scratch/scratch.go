package main

import (
	"fmt"

	"github.com/shasderias/serial"

	"github.com/shasderias/modbus"
	"github.com/shasderias/modbus/transport/rtu"
)

func main() {
	port, err := serial.Open("/tmp/port2", &serial.Config{
		BaudRate: 19200,
		DataBits: 8,
		StopBits: serial.StopBits1,
		Parity:   serial.ParityEven,
	})
	if err != nil {
		panic(err)
	}

	client, err := modbus.NewClient(1, rtu.NewClient(port))
	if err != nil {
		panic(err)
	}

	resp, err := client.ReadInputRegisters(1, 6)
	if err != nil {
		panic(err)
	}

	fmt.Println(resp)

	//resp, err := client.WriteMultipleCoils(0, []bool{true})
	//if err != nil {
	//	panic(err)
	//}
	//
	//fmt.Println(resp)
	//if _, err := client.WriteBits(modbus.FuncCodeWriteMultipleCoils, 7, []bool{true, false, true}); err != nil {
	//	panic(err)
	//}
}
