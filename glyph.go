package main

import "fmt"

func (g *Game) addGlyph() {
	glyph := g.currentGlyph1Bit()
	g.glyphs = append(g.glyphs, glyph)
	g.glyphIndex = len(g.glyphs)
	g.clear()
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
