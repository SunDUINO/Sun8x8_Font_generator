/*
       Generator czcionek 8x8 dla matryc LED (Ebiten)

Autor: SunRiver / Lothar TeaM
  WWW: https://forum.lothar-team.pl/
  Git: https://github.com/SunDUINO/Sun8x8_Font_generator.git
 Plik: main.go

*/

package main

import (
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var version = "0.0.16"
var fontFace font.Face
var appName = "Sun8x8 - Generator czcionek 8x8 wer: "
var matrixSerial *SerialMatrix

func main() {
	// Wczytujemy czcionkę obsługującą Unicode
	ttfBytes, err := os.ReadFile("resource/Itim-Regular.ttf")
	if err != nil {
		log.Fatal(err)
	}

	ttf, err := opentype.Parse(ttfBytes)
	if err != nil {
		log.Fatal(err)
	}

	fontFace, err = opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    16,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	g := NewGame()
	ebiten.SetWindowSize(CanvasW, CanvasH)
	ebiten.SetWindowTitle(appName + version)

	g.updatePreviewText()

	//if err := ebiten.RunGame(g); err != nil {
	//	log.Fatal(err)
	//}

	// zamykamy port na koniec programu
	// dodatkowe zabezpieczenie przy zamykaniu portu serial
	if matrixSerial != nil {
		defer matrixSerial.Close()
	}

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
