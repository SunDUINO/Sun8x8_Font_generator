package main

import (
	"strconv"
	"time"

	"github.com/tarm/serial"
)

func detectSerialPort() string {

	for i := 3; i <= 40; i++ {
		port := "COM" + strconv.Itoa(i)

		c := &serial.Config{
			Name:        port,
			Baud:        115200,
			ReadTimeout: time.Millisecond * 200,
		}

		p, err := serial.OpenPort(c)
		if err != nil {
			continue
		}

		// handshake
		_, err = p.Write([]byte{0xAA})
		if err != nil {
			continue // błąd zapisu, pomijamy port
		}

		buf := make([]byte, 1)
		n, _ := p.Read(buf)

		err = p.Close()
		if err != nil {
			continue
		}

		if n == 1 && buf[0] == 0x55 {
			return port // TO JEST PICO
		}
	}

	return ""
}
