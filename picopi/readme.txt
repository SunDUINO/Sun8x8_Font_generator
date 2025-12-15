Program dla rasppberry pico pi

raspberry Pico pi SDK

-- matryca 32x8 z max7219  4* 8x8
-- pico pi

Podłączenie matrycy do picopi:

--> SPI_PORT spi0
-- PIN_CS   17
-- PIN_SCK  18
-- PIN_MOSI 19

Uwaga !!

Program ma tymczasowo na sztywno ustawiony COM13 dla Windows
można to zmienić w kodzie -- plik game.go
    // ------ tutaj inicjalizacja serial -------
	matrixSerial = NewSerialMatrix("COM13", GridW, GridH)

jeśli kompiliujecie samodzielnie można sobie dobrać ręcznie.
lub zmienić poret dla picoPi w komputerze.

--> później dodam możliwośc wyboru w programie i dorobię podgląd animacji.

