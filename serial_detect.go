/*
Autor: SunRiver / Lothar TeaM
  WWW: https://forum.lothar-team.pl/
  Git: https://github.com/SunDUINO/Sun8x8_Font_generator.git
 Plik: serial_detect.go

Moduł do wykrywania mikrokontrolera Raspberry Pi Pico (lub kompatybilnego urządzenia)
podłączonego przez port szeregowy (COM).

Działanie:
- Funkcja `detectSerialPort` skanuje porty COM od 3 do 40 i sprawdza, czy jest podłączone Pico.
- Funkcja `isPicoOnPort` wykonuje handshake wysyłając bajt 0xAA i oczekując odpowiedzi 0x55.
- Po wykryciu urządzenia zwracany jest numer portu COM, na którym jest Pico.

*/

package main

import (
	"log"
	"strconv"
	"time"

	"github.com/tarm/serial"
)

// Wykrywanie pico na porcie --  (0xAA) odpowiedź (0x55)
func isPicoOnPort(port string) bool {

	c := &serial.Config{
		Name:        port,
		Baud:        115200,
		ReadTimeout: time.Millisecond * 200,
	}

	p, err := serial.OpenPort(c)
	if err != nil {
		return false
	}
	defer func() {
		if err := p.Close(); err != nil {
			log.Printf("Warning: nie można otwożyć portu:%s: %v", port, err)
		}
	}()

	// handshake
	if _, err := p.Write([]byte{0xAA}); err != nil {
		return false
	}

	buf := make([]byte, 1)
	n, err := p.Read(buf)
	if err != nil {
		return false
	}

	return n == 1 && buf[0] == 0x55
}

func detectSerialPort() string {

	for i := 3; i <= 40; i++ {
		port := "COM" + strconv.Itoa(i)

		if isPicoOnPort(port) {
			return port // to jest port piko
		}
	}

	return ""
}
