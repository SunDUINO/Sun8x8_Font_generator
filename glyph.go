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

func (g *Game) addGlyph() {
	glyph := g.currentGlyph1Bit()

	// dodaj znak do listy zapisanych
	g.glyphs = append(g.glyphs, glyph)

	// ustaw aktywny znak (ostatnio zapisany)
	g.activeGlyph = len(g.glyphs) - 1

	// przesuwaj okno podglądu (max 8 znaków)
	if g.activeGlyph >= g.glyphViewOfs+8 {
		g.glyphViewOfs = g.activeGlyph - 7
	}

	// wyczyść edytor (M0)
	g.clear()

	// zaktualizuj displayGlyphs – tylko poprzednie 3 znaki na M1–M3
	g.displayGlyphs = nil
	start := len(g.glyphs) - 4
	if start < 0 {
		start = 0
	}
	if len(g.glyphs) > 1 {
		g.displayGlyphs = g.glyphs[start : len(g.glyphs)-1]
	}

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
