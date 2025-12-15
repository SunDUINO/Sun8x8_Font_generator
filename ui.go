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
	// 2. Siatka edytora
	// ----------------------
	for y := 0; y < GridH; y++ {
		for x := 0; x < GridW; x++ {
			px := x * CellSize
			py := y * CellSize

			// tło komórki
			drawRect(screen, px+1, py+1, CellSize-2, CellSize-2, color.RGBA{R: 0x22, G: 0x22, B: 0x26, A: 0xff})

			// wypełnienie paletą
			idx := g.cells[y][x]
			if idx != 0 {
				c := palette[idx]
				drawRect(screen, px+3, py+3, CellSize-6, CellSize-6, c)
			}

			// obramowanie
			drawRect(screen, px, py, CellSize, 1, color.RGBA{R: 0x44, G: 0x44, B: 0x50, A: 0xff})
			drawRect(screen, px, py+CellSize-1, CellSize, 1, color.RGBA{R: 0x44, G: 0x44, B: 0x50, A: 0xff})
			drawRect(screen, px, py, 1, CellSize, color.RGBA{R: 0x44, G: 0x44, B: 0x50, A: 0xff})
			drawRect(screen, px+CellSize-1, py, 1, CellSize, color.RGBA{R: 0x44, G: 0x44, B: 0x50, A: 0xff})
		}
	}

	// 2A --- Przycisk: Dodaj znak ---
	btnX := 12
	btnY := GridH*CellSize + 12
	btnW := GridW*CellSize - 24
	btnH := 32

	drawRect(screen, btnX, btnY, btnW, btnH, color.RGBA{R: 0x2A, G: 0x80, B: 0xFF, A: 0xff})
	drawText(
		screen,
		fmt.Sprintf("Zapisz znak (%d)", g.glyphIndex),
		btnX+12,
		btnY+22,
	)

	// ----------------------
	// 3. Suwak prędkości animacji
	// ----------------------
	speedPercent := int(g.sliderValue * 100)
	drawText(screen, fmt.Sprintf("Prędkość animacji: %d%%", speedPercent), g.sliderX, g.sliderY-20)
	drawRect(screen, g.sliderX, g.sliderY, g.sliderW, g.sliderH, color.RGBA{R: 0x30, G: 0x30, B: 0x36, A: 0xff})
	handleX := g.sliderX + int(g.sliderValue*float64(g.sliderW))
	drawRect(screen, handleX-4, g.sliderY-2, 8, g.sliderH+4, color.RGBA{R: 0x2A, G: 0x80, B: 0xFF, A: 0xff})

	// ----------------------
	// 4. Panel po prawej - tło i przyciski
	// ----------------------
	x0 := GridW * CellSize
	y0 := 0
	drawRect(screen, x0, y0, CanvasW-x0, GridH*CellSize, color.RGBA{R: 0x18, G: 0x18, B: 0x1C, A: 0xff})

	btnW = 180
	btnH = 34
	pad := 12
	bx := x0 + pad
	by := y0 + pad

	// Tryb edycji
	drawRect(screen, bx, by, btnW, btnH, color.RGBA{R: 0x30, G: 0x30, B: 0x36, A: 0xff})
	drawText(screen, fmt.Sprintf("Tryb: %s (M)", g.modeLabel()), bx+8, by+23)

	// Clear
	by += btnH + pad
	drawRect(screen, bx, by, btnW, btnH, color.RGBA{R: 0x30, G: 0x30, B: 0x36, A: 0xff})
	drawText(screen, "Wyczyść (C)", bx+8, by+23)

	// Eksport: tryb
	by += btnH + pad
	drawRect(screen, bx, by, btnW, btnH, color.RGBA{R: 0x30, G: 0x30, B: 0x36, A: 0xff})
	drawText(screen, fmt.Sprintf("Eksport: %s", g.exportModeLabel()), bx+8, by+23)

	// Eksport: format
	by += btnH + pad
	drawRect(screen, bx, by, btnW, btnH, color.RGBA{R: 0x30, G: 0x30, B: 0x36, A: 0xff})
	drawText(screen, fmt.Sprintf("Format: %s", g.exportFormatLabel()), bx+8, by+23)

	// Export C
	by += btnH + pad
	drawRect(screen, bx, by, btnW, btnH, color.RGBA{R: 0x2A, G: 0x80, B: 0xFF, A: 0xff})
	drawText(screen, "Eksportuj C (E)", bx+8, by+23)

	// Export PROGMEM
	by += btnH + pad
	drawRect(screen, bx, by, btnW, btnH, color.RGBA{R: 0x2A, G: 0x80, B: 0xFF, A: 0xff})
	drawText(screen, "Eksportuj PROGMEM", bx+8, by+23)

	// Export PNG
	by += btnH + pad
	drawRect(screen, bx, by, btnW, btnH, color.RGBA{R: 0x2A, G: 0x80, B: 0xFF, A: 0xff})
	drawText(screen, "Eksportuj PNG", bx+8, by+26)

	// Toggle animacji
	by += btnH + pad
	col := color.RGBA{R: 0x60, G: 0x60, B: 0x60, A: 0xff}
	if g.animRunning {
		col = color.RGBA{R: 0x34, G: 0xC7, B: 0x34, A: 0xff}
	}
	drawRect(screen, bx, by, btnW, btnH, col)
	if g.animRunning {
		drawText(screen, "Stop animacji", bx+8, by+23)
	} else {
		drawText(screen, "Start animacji", bx+8, by+23)
	}

	// Kierunek animacji
	by += btnH + pad
	drawRect(screen, bx, by, btnW, btnH, color.RGBA{R: 0x30, G: 0x30, B: 0x36, A: 0xff})
	dirText := "Kierunek: "
	if g.animDir < 0 {
		dirText += "←"
	} else {
		dirText += "→"
	}
	drawText(screen, dirText, bx+8, by+23)

	// Podgląd HEX/BIN
	by += btnH + pad
	drawRect(screen, bx, by, btnW, btnH, color.RGBA{R: 0x30, G: 0x30, B: 0x36, A: 0xff})
	drawText(screen, "Podgląd HEX/BIN", bx+8, by+23)

	// Informacja o ostatnim eksporcie
	drawText(screen, fmt.Sprintf("Plik: %s", g.lastExport), bx, by+btnH+pad+10)

	// ----------------------
	// 5. Panel tekstowy z prawej - previewText
	// ----------------------
	textX := x0 + btnW + pad*2
	textY := pad
	w := CanvasW - textX - pad
	h := CanvasH - textY - pad - 120
	drawRect(screen, textX, textY, w, h, color.RGBA{R: 0x0E, G: 0x0E, B: 0x10, A: 0xff})

	lines := strings.Split(g.previewText, "\n")
	yline := textY + 6
	for i, ln := range lines {
		if i > 30 {
			break
		}
		if len(ln) > 80 {
			ln = ln[:80] + "..."
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

	off := int(g.animX)
	animSize := int(float64(CellSize) * AnimCellScale)
	for y := 0; y < GridH; y++ {
		for x := 0; x < GridW*4; x++ {
			px := playX + x*animSize + off
			py := playY + y*animSize
			if px+animSize < playX || px > playX+playW {
				continue
			}
			srcX := x % GridW
			srcY := y
			idx := g.cells[srcY][srcX]
			if idx != 0 {
				c := palette[idx]
				drawRect(screen, px+1, py+1, animSize-2, animSize-2, c)
			}
		}
	}

	// ----------------------
	// 6A. Podgląd zapisanych znaków (glyphs)
	// ----------------------
	glyphX := textX
	glyphY := textY + h + 12

	scale := 3   // powiększenie pojedynczego piksela
	spacing := 6 // odstęp
	perRow := 8  // ile znaków w jednym rzędzie

	for i, glyph := range g.glyphs {
		cx := i % perRow
		cy := i / perRow

		x := glyphX + cx*(8*scale+spacing)
		y := glyphY + cy*(8*scale+20)

		// tło miniatury
		drawRect(
			screen,
			x-2,
			y-2,
			8*scale+4,
			8*scale+4,
			color.RGBA{R: 0x22, G: 0x22, B: 0x26, A: 0xff},
		)

		// sam glyph
		drawGlyphPreview(
			screen,
			glyph,
			x,
			y,
			scale,
			color.White,
		)

		// numer znaku (indeks)
		ebitenutil.DebugPrintAt(
			screen,
			fmt.Sprintf("%d", i),
			x,
			y+8*scale+2,
		)
	}

	// ----------------------
	// 6B. Status Serial
	// ----------------------
	statusY := CanvasH - 40
	if g.serialStatus != "" {
		text.Draw(screen, "Serial: "+g.serialStatus, fontFace, 8, statusY, color.White)
	}

	// ----------------------
	// 7. Pomoc u dołu
	// ----------------------
	helpY := CanvasH - 20
	help := "LPM: kliknij komórkę aby zmienić; M: zmiana trybu; C: wyczyść. Kliknij przyciski po prawej."
	text.Draw(screen, help, fontFace, 8, helpY, color.White)
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
