package main

type WS2812Color struct {
	G uint8
	R uint8
	B uint8
}

// WS2812Index mapuje współrzędne (x,y) na liniowy indeks 0..63
// przy układzie zigzag (serpentine):
// rząd parzysty: 0→7, rząd nieparzysty: 15←8 dla y=1 itd.
func WS2812Index(x, y int) int {
	width := 8
	if y%2 == 0 {
		return y*width + x
	}
	return y*width + (width - 1 - x)
}
