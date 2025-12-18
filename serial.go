/*
Autor: SunRiver / Lothar TeaM
  WWW: https://forum.lothar-team.pl/
  Git: https://github.com/SunDUINO/Sun8x8_Font_generator.git
 Plik: serial.go

*/

package main

import (
	"fmt"

	"github.com/tarm/serial"
)

type SerialMatrix struct {
	port     *serial.Port
	portName string
	width    int
	height   int
}

func NewSerialMatrix(portName string, width, height int) *SerialMatrix {
	if portName == "" {
		fmt.Println("Serial: brak portu")
		return nil
	}

	c, err := serial.OpenPort(&serial.Config{
		Name: portName,
		Baud: 115200,
	})
	if err != nil {
		fmt.Println("Serial: nie można otworzyć", portName)
		return nil
	}

	fmt.Println("Serial: podłączono", portName)

	return &SerialMatrix{
		port:   c,
		width:  width,
		height: height,
	}
}

func (s *SerialMatrix) Close() {
	if s.port != nil {
		if err := s.port.Close(); err != nil {
			fmt.Println("Warning: nie udało się zamknąć portu:", err)
		}
		s.port = nil
	}
}

// SendFrame - konwersja 8x8 (lub NxM) komórek na bajty
func (s *SerialMatrix) SendFrame(cells [][]int) error {
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
			fmt.Println("Błąd wysyłki do matrycy:", err)
			return err
		}
	}
	return nil
}
