/*
Autor: SunRiver / Lothar TeaM
  WWW: https://forum.lothar-team.pl/
  Git: https://github.com/SunDUINO/Sun8x8_Font_generator.git
 Plik: glyph.go

*/

package main

import "fmt"

/*
func (g *Game) addGlyph() {
	glyph := g.currentGlyph1Bit()
	g.glyphs = append(g.glyphs, glyph)
	g.glyphIndex = len(g.glyphs)
	g.clear()
	g.updateFontPreview()
}
*/

// addGlyph dodaje aktualny znak do listy glyphów, przesuwa wyświetlane znaki
// i wyczyści edytor
func (g *Game) addGlyph() {
	glyph := g.currentGlyph1Bit()

	// dodaj znak do pełnej listy
	g.glyphs = append(g.glyphs, glyph)

	// ustaw aktywny znak (ostatnio zapisany)
	g.activeGlyph = len(g.glyphs) - 1

	// przesuwamy okno podglądu (max 8 znaków)
	if g.activeGlyph >= g.glyphViewOfs+8 {
		g.glyphViewOfs = g.activeGlyph - 7
	}

	// aktualizujemy wyświetlane znaki (dla matryc M1–M3)
	g.displayGlyphs = [][]byte{}
	for i := 0; i < 3; i++ {
		idx := g.activeGlyph + i - 2 // M1 = poprzedni, M2 = kolejny...
		if idx >= 0 && idx < len(g.glyphs) && idx != g.activeGlyph {
			g.displayGlyphs = append(g.displayGlyphs, g.glyphs[idx])
		}
	}

	g.glyphIndex = g.activeGlyph

	// wyczyść edytor
	g.clear()

	// odśwież preview fontu
	g.updateFontPreview()
}

func (g *Game) currentGlyph1Bit() []byte {
	out := make([]byte, 8)
	for y := 0; y < GridH; y++ {
		var row byte
		for x := 0; x < GridW; x++ {
			if g.cells[y][x] != 0 {
				row |= 1 << (7 - x)
			}
		}
		out[y] = row
	}
	return out
}

func (g *Game) updateFontPreview() {
	s := fmt.Sprintf("char font8x8[%d][8] = {\n", len(g.glyphs))
	for i, glyph := range g.glyphs {
		s += "  { "
		for j, b := range glyph {
			s += fmt.Sprintf("0x%02X", b)
			if j < 7 {
				s += ", "
			}
		}
		s += fmt.Sprintf(" }, // znak %d\n", i)
	}
	s += "};\n"
	g.previewText = s
}
