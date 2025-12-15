// ----------------------------------------------------------------------------------
//
// Generator czcionek 8x8 dla matryc LED (Ebiten)
// Autor:  SunRiver / Lothar TeaM
// Strona: https://forum.lothar-team.pl/
// Program dla picopi
//
// ----------------------------------------------------------------------------------

#include <string.h>
#include "pico/stdlib.h"
#include "hardware/spi.h"

#define SPI_PORT spi0
#define PIN_CS   17
#define PIN_SCK  18
#define PIN_MOSI 19

#define NUM_MATRICES 4
#define FRAME_BYTES (NUM_MATRICES * 8)

// -------- CS helper --------
static inline void cs_low()  { gpio_put(PIN_CS, 0); }
static inline void cs_high() { gpio_put(PIN_CS, 1); }

// -------- MAX7219 --------
void max7219_send_all(uint8_t reg, uint8_t data)
{
	uint8_t buf[NUM_MATRICES * 2];
	for (int i = 0; i < NUM_MATRICES; i++) {
		buf[i * 2]     = reg;
		buf[i * 2 + 1] = data;
	}
	cs_low();
	spi_write_blocking(SPI_PORT, buf, sizeof(buf));
	cs_high();
}

void max7219_init()
{
	max7219_send_all(0x0F, 0x00); // test off
	max7219_send_all(0x0C, 0x01); // normal mode
	max7219_send_all(0x0B, 0x07); // scan limit
	max7219_send_all(0x09, 0x00); // no decode
	max7219_send_all(0x0A, 0x08); // brightness
}

void max7219_draw(uint8_t *frame)
{
	for (int row = 0; row < 8; row++) {
		cs_low();
		for (int m = NUM_MATRICES - 1; m >= 0; m--) {
			uint8_t data[2] = { row + 1, frame[m * 8 + row] };
			spi_write_blocking(SPI_PORT, data, 2);
		}
		cs_high();
	}
}

// -------- main --------
int main()
{
	stdio_init_all(); // USB CDC
	sleep_ms(500); // poczekaj na inicjalizację USB

	spi_init(SPI_PORT, 10 * 1000 * 1000);
	gpio_set_function(PIN_SCK, GPIO_FUNC_SPI);
	gpio_set_function(PIN_MOSI, GPIO_FUNC_SPI);

	gpio_init(PIN_CS);
	gpio_set_dir(PIN_CS, GPIO_OUT);
	gpio_put(PIN_CS, 1);

	max7219_init();

	uint8_t frame[FRAME_BYTES];
	memset(frame, 0, sizeof(frame));

	while (true) {
		int received = 0;
		while (received < FRAME_BYTES) {
			int c = getchar_timeout_us(1000); // 1ms timeout
			if (c >= 0) {
				frame[received++] = (uint8_t)c;
			}
		}
		max7219_draw(frame);
	}
}
