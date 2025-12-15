package main

import (
	"fmt"

	"github.com/tarm/serial"
)

type SerialMatrix struct {
	port   *serial.Port
	width  int
	height int
}

func NewSerialMatrix(portName string, width, height int) *SerialMatrix {
	c, err := serial.OpenPort(&serial.Config{Name: portName, Baud: 115200})
	if err != nil {
		panic(fmt.Sprintf("Nie można otworzyć portu %s: %v", portName, err))
	}
	return &SerialMatrix{
		port:   c,
		width:  width,
		height: height,
	}
}

func (s *SerialMatrix) Close() {
	if s.port != nil {
		s.port.Close()
	}
}

// SendFrame - konwersja 8x8 (lub NxM) komórek na bajty
func (s *SerialMatrix) SendFrame(cells [][]int) {
	frame := make([]byte, s.height)
	for y := 0; y < s.height; y++ {
		var b byte
		for x := 0; x < s.width; x++ {
			if cells[y][x] != 0 {
				b |= 1 << (7 - x)
			}
		}
		frame[y] = b
	}
	if s.port != nil {
		_, err := s.port.Write(frame)
		if err != nil {
			fmt.Println("Błąd wysyłki:", err)
		}
	}
}
