package main

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

// Draw rysuje UI i podglądy
func (g *Game) Draw(screen *ebiten.Image) {
	// ----------------------
	// 1. Tło
	// ----------------------
	screen.Fill(color.RGBA{R: 0x10, G: 0x10, B: 0x12, A: 0xff})

	// ----------------------
	// Helper do rysowania prostokąta
	// ----------------------
	drawRect := func(img *ebiten.Image, x, y, w, h int, col color.Color) {
		rect := ebiten.NewImage(w, h)
		rect.Fill(col)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		img.DrawImage(rect, op)
	}

	// ----------------------
	// 2. Siatka edytora (cells) z przewijaniem
	// ----------------------
	_, editorH := GridW*CellSize, GridH*CellSize

	for y := 0; y < GridH; y++ {
		py := y*CellSize - g.cellScroll
		if py+CellSize < 0 || py > editorH {
			continue
		}
		for x := 0; x < GridW; x++ {
			px := x * CellSize
			drawRect(screen, px+1, py+1, CellSize-2, CellSize-2, color.RGBA{R: 0x22, G: 0x22, B: 0x26, A: 0xff})
			idx := g.cells[y][x]
			if idx != 0 {
				drawRect(screen, px+3, py+3, CellSize-6, CellSize-6, palette[idx])
			}
			// obramowanie
			drawRect(screen, px, py, CellSize, 1, color.RGBA{R: 0x44, G: 0x44, B: 0x50, A: 0xff})
			drawRect(screen, px, py+CellSize-1, CellSize, 1, color.RGBA{R: 0x44, G: 0x44, B: 0x50, A: 0xff})
			drawRect(screen, px, py, 1, CellSize, color.RGBA{R: 0x44, G: 0x44, B: 0x50, A: 0xff})
			drawRect(screen, px+CellSize-1, py, 1, CellSize, color.RGBA{R: 0x44, G: 0x44, B: 0x50, A: 0xff})
		}
	}

	// ----------------------
	// 2A. Przycisk: Dodaj znak
	// ----------------------
	btnX := 12
	btnY := GridH*CellSize + 12
	btnW := GridW*CellSize - 24
	btnH := 32
	drawRect(screen, btnX, btnY, btnW, btnH, color.RGBA{R: 0x2A, G: 0x80, B: 0xFF, A: 0xff})
	drawText(screen, fmt.Sprintf("Zapisz znak (%d)", g.glyphIndex), btnX+12, btnY+22)

	// ----------------------
	// 3. Suwak prędkości animacji
	// ----------------------
	speedPercent := int(g.sliderValue * 100)
	drawText(screen, fmt.Sprintf("Prędkość animacji: %d%%", speedPercent), g.sliderX, g.sliderY-20)
	drawRect(screen, g.sliderX, g.sliderY, g.sliderW, g.sliderH, color.RGBA{R: 0x30, G: 0x30, B: 0x36, A: 0xff})
	handleX := g.sliderX + int(g.sliderValue*float64(g.sliderW))
	drawRect(screen, handleX-4, g.sliderY-2, 8, g.sliderH+4, color.RGBA{R: 0x2A, G: 0x80, B: 0xFF, A: 0xff})

	// ----------------------
	// 4. Panel po prawej: tło i przyciski
	// ----------------------
	x0 := GridW * CellSize
	y0 := 0
	drawRect(screen, x0, y0, CanvasW-x0, GridH*CellSize, color.RGBA{R: 0x18, G: 0x18, B: 0x1C, A: 0xff})

	btnW = 180
	btnH = 34
	pad := 12
	bx := x0 + pad
	by := y0 + pad

	buttons := []struct {
		label string
		color color.RGBA
	}{
		{fmt.Sprintf("Tryb: %s (M)", g.modeLabel()), color.RGBA{R: 0x30, G: 0x30, B: 0x36, A: 0xff}},
		{"Wyczyść (C)", color.RGBA{R: 0x30, G: 0x30, B: 0x36, A: 0xff}},
		{fmt.Sprintf("Eksport: %s", g.exportModeLabel()), color.RGBA{R: 0x30, G: 0x30, B: 0x36, A: 0xff}},
		{fmt.Sprintf("Format: %s", g.exportFormatLabel()), color.RGBA{R: 0x30, G: 0x30, B: 0x36, A: 0xff}},
		{"Eksportuj C", color.RGBA{R: 0x2A, G: 0x80, B: 0xFF, A: 0xff}},
		{"Eksportuj PROGMEM", color.RGBA{R: 0x2A, G: 0x80, B: 0xFF, A: 0xff}},
		{"Eksportuj PNG", color.RGBA{R: 0x2A, G: 0x80, B: 0xFF, A: 0xff}},
		{"Start/Stop animacji", color.RGBA{R: 0x60, G: 0x60, B: 0x60, A: 0xff}},
		{"Kierunek animacji", color.RGBA{R: 0x30, G: 0x30, B: 0x36, A: 0xff}},
		{"Podgląd HEX/BIN", color.RGBA{R: 0x30, G: 0x30, B: 0x36, A: 0xff}},
	}

	for _, b := range buttons {
		col := b.color
		lbl := b.label
		if lbl == "Start/Stop animacji" && g.animRunning {
			col = color.RGBA{R: 0x34, G: 0xC7, B: 0x34, A: 0xff}
		}
		drawRect(screen, bx, by, btnW, btnH, col)
		drawText(screen, lbl, bx+8, by+23)
		by += btnH + pad
	}

	drawText(screen, fmt.Sprintf("Plik: %s", g.lastExport), bx, by+10)

	// ----------------------
	// 5. Podgląd tablicy C z przewijaniem
	// ----------------------
	textX := x0 + btnW + pad*2
	textY := pad
	textW := CanvasW - textX - pad
	textH := CanvasH - textY - pad - 120
	drawRect(screen, textX, textY, textW, textH, color.RGBA{R: 0x0E, G: 0x0E, B: 0x10, A: 0xff})
	g.textViewH = textH

	yline := textY + 6 - g.textScroll
	for _, ln := range strings.Split(g.previewText, "\n") {
		if yline+12 < textY || yline > textY+textH {
			yline += 12
			continue
		}
		ebitenutil.DebugPrintAt(screen, ln, textX+6, yline)
		yline += 12
	}

	// ----------------------
	// 6. Podgląd animacji
	// ----------------------
	playX := 12
	playY := GridH*CellSize + 280
	scaleW := 0.78
	scaleH := 2.0
	playW := int(float64(GridW*CellSize*4) * scaleW)
	playH := int(float64(CellSize) * scaleH)
	drawRect(screen, playX, playY, playW, playH, color.RGBA{R: 0x08, G: 0x08, B: 0x08, A: 0xff})

	animSize := int(float64(CellSize) * AnimCellScale)
	off := int(g.animX)
	for y := 0; y < GridH; y++ {
		for x := 0; x < GridW*4; x++ {
			px := playX + x*animSize + off
			py := playY + y*animSize
			if px+animSize < playX || px > playX+playW {
				continue
			}
			idx := g.cells[y][x%GridW]
			if idx != 0 {
				drawRect(screen, px+1, py+1, animSize-2, animSize-2, palette[idx])
			}
		}
	}

	// ----------------------
	// 6A. Podgląd zapisanych glyphów z przewijaniem
	// ----------------------
	glyphX := textX
	glyphY := textY + textH + 12
	//glyphW := textW
	glyphH := CanvasH - glyphY - 40
	g.glyphViewH = glyphH
	scale := 3   // <-- tutaj deklaracja przed użyciem
	spacing := 6 // <-- tutaj deklaracja przed użyciem
	perRow := 8

	startY := 6 - g.glyphScroll
	for i, glyph := range g.glyphs {
		cx := i % perRow
		cy := i / perRow
		x := glyphX + cx*(8*scale+spacing)
		y := startY + cy*(8*scale+20)
		if y+8*scale < glyphY || y > glyphY+glyphH {
			continue
		}
		drawRect(screen, x-2, y-2, 8*scale+4, 8*scale+4, color.RGBA{R: 0x22, G: 0x22, B: 0x26, A: 0xff})
		drawGlyphPreview(screen, glyph, x, y, scale, color.White)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", i), x, y+8*scale+2)
	}

	// ----------------------
	// 7. Pomoc u dołu
	// ----------------------
	help := "LPM: kliknij komórkę aby zmienić; M: zmiana trybu; C: wyczyść. Kliknij przyciski po prawej."
	drawText(screen, help, 8, CanvasH-26)
}

func drawGlyphPreview(
	screen *ebiten.Image,
	glyph []byte,
	x, y int,
	scale int,
	col color.Color,
) {
	for row := 0; row < 8; row++ {
		b := glyph[row]
		for bit := 0; bit < 8; bit++ {
			if (b & (1 << (7 - bit))) != 0 {
				rect := ebiten.NewImage(scale, scale)
				rect.Fill(col)

				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(
					float64(x+bit*scale),
					float64(y+row*scale),
				)
				screen.DrawImage(rect, op)
			}
		}
	}
}

func drawText(screen *ebiten.Image, s string, x, y int) {
	text.Draw(screen, s, fontFace, x, y, color.White)
}

func (g *Game) isInTextPreview(x, y int) bool {
	return x >= GridW*CellSize+180 &&
		y >= 12 &&
		y <= 12+g.textViewH
}

func (g *Game) isInGlyphPreview(x, y int) bool {
	return x >= GridW*CellSize+180 &&
		y >= 12+g.textViewH+12 &&
		y <= 12+g.textViewH+12+g.glyphViewH
}
