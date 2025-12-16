// ----------------------------------------------------------------------------------
//
// Generator czcionek 8x8 dla matryc LED (Ebiten)
// Autor:  SunRiver / Lothar TeaM
// Strona: https://forum.lothar-team.pl/
//
// ----------------------------------------------------------------------------------

package main

import (
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var version = "0.0.11"
var fontFace font.Face

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
	ebiten.SetWindowTitle("Sun8x8 - Generator czcionek 8x8 wer: " + version)

	g.updatePreviewText()

	//if err := ebiten.RunGame(g); err != nil {
	//	log.Fatal(err)
	//}

	// zamykamy port na koniec programu
	// zabezpieczenie przy zamykaniu serial
	if matrixSerial != nil {
		defer matrixSerial.Close()
	}

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
