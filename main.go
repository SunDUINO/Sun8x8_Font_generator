// ----------------------------------------------------------------------------------
//
// Generator czcionek 8x8 dla matryc LED (Ebiten)
// Autor:  SunRiver / Lothar TeaM
// Strona: https://forum.lothar-team.pl/
//
// - obsługa 1-bit, 2-bit (2 poziomy per pixel) i RGB (RGB565)
// - eksport do: czystego C, PROGMEM (Arduino/AVR), i plików pomocniczych (PNG/JSON)
// - podgląd macierzy w HEX i BIN
//
// ----------------------------------------------------------------------------------

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sqweek/dialog"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	GridW    = 8
	GridH    = 8
	CellSize = 28
	CanvasW  = GridW*CellSize + 500 // zostawiam miejsce na panele z prawej
	CanvasH  = GridH*CellSize + 400 // dolny pasek na przyciski
)

// tryby edycji
const (
	ModeMono = iota
	ModeTwo
	ModeRGB
)

// eksport: bit-depth
const (
	Export1Bit = iota
	Export2Bit
	ExportRGB
)

var (
	palette = []color.RGBA{
		{0x00, 0x00, 0x00, 0xff}, // 0 = off (czarny)
		{0xff, 0xff, 0xff, 0xff}, // 1 = biały (mono)
		{0xff, 0x00, 0x00, 0xff}, // red
		{0x00, 0xff, 0x00, 0xff}, // green
		{0x00, 0x00, 0xff, 0xff}, // blue
		{0xff, 0xff, 0x00, 0xff}, // yellow
		{0xff, 0x00, 0xff, 0xff}, // magenta
		{0x00, 0xff, 0xff, 0xff}, // cyan
	}
)

var version = "0.0.4"
var fontFace font.Face

const AnimCellScale = 0.25 // 1/4 rozmiaru

// Game struktura gry / edytora
type Game struct {
	cells      [GridH][GridW]int // indeks palety; 0 = off
	mode       int
	monoColor  int
	twoA       int
	twoB       int
	mouseDown  bool
	lastExport string

	// UI: eksport
	exportMode   int // 1-bit / 2-bit / RGB
	exportFormat int // 0=c (czysty C tablica),1=PROGMEM

	// Tekstowy podgląd wygenerowanych tablic
	previewText string

	// animacja przewijania
	animX       float64
	animRunning bool
	animSpeed   float64 // nowa zmienna sterująca prędkością

	// suwak prędkości animacji
	sliderX       int
	sliderY       int
	sliderW       int
	sliderH       int
	sliderValue   float64 // 0.0 – 1.0
	sliderGrabbed bool

	// ADD -- kierunek animacji
	animDir int
}

func NewGame() *Game {
	g := &Game{}
	g.mode = ModeMono
	g.monoColor = 1
	g.twoA = 2
	g.twoB = 3
	g.exportMode = Export1Bit
	g.animSpeed = 0.5

	// ADD — inicjalizacja suwaka
	g.sliderX = 12                   // left offset
	g.sliderY = GridH*CellSize + 260 // tuż nad panelem animacji
	g.sliderW = 300
	g.sliderH = 14
	g.sliderValue = 0.5
	// ADD -- kierunek animacji
	g.animDir = -1 // domyślnie w lewo
	return g
}

func (g *Game) Layout(_, _ int) (int, int) { return CanvasW, CanvasH }

func (g *Game) Update() error {
	// mysz
	x, y := ebiten.CursorPosition()
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if !g.mouseDown {
			g.mouseDown = true
			if g.handleSlider(x, y) {
				// slider obsłużony
			} else {
				g.handleLeftClick(x, y)
			}
		} else {
			// jeśli trzymamy suwak
			if g.sliderGrabbed {
				g.sliderValue = float64(x-g.sliderX) / float64(g.sliderW)
				if g.sliderValue < 0 {
					g.sliderValue = 0
				}
				if g.sliderValue > 1 {
					g.sliderValue = 1
				}
				g.animSpeed = g.sliderValue // prędkość animacji 0-100%
			}
		}
	} else {
		g.mouseDown = false
		g.sliderGrabbed = false
	}

	// skróty
	if ebiten.IsKeyPressed(ebiten.KeyM) {
		g.mode = (g.mode + 1) % 3
		time.Sleep(140 * time.Millisecond)
	}
	if ebiten.IsKeyPressed(ebiten.KeyC) {
		g.clear()
		time.Sleep(120 * time.Millisecond)
	}

	// animacja
	if g.animRunning {
		g.animX += float64(g.animDir) * 60 * ebiten.CurrentTPS() / 60.0 * g.animSpeed
		if g.animDir < 0 && g.animX < -float64(GridW*CellSize) {
			g.animX = float64(GridW * CellSize)
		} else if g.animDir > 0 && g.animX > float64(GridW*CellSize) {
			g.animX = -float64(GridW * CellSize)
		}
	}

	return nil
}

// obsługa slidera prędkości animacji
func (g *Game) handleSlider(x, y int) bool {
	if x >= g.sliderX && x <= g.sliderX+g.sliderW && y >= g.sliderY && y <= g.sliderY+g.sliderH {
		g.sliderGrabbed = true
		return true
	}
	return false
}

// obsługa kliknięć: siatka i przyciski
func (g *Game) handleLeftClick(x, y int) {
	// kliknięcia na siatkę
	if y < GridH*CellSize {
		cx := x / CellSize
		yi := y / CellSize
		if cx >= 0 && cx < GridW && yi >= 0 && yi < GridH {
			switch g.mode {
			case ModeMono:
				if g.cells[yi][cx] == 0 {
					g.cells[yi][cx] = g.monoColor
				} else {
					g.cells[yi][cx] = 0
				}
			case ModeTwo:
				n := g.cells[yi][cx]
				if n == 0 {
					g.cells[yi][cx] = g.twoA
				} else if n == g.twoA {
					g.cells[yi][cx] = g.twoB
				} else {
					g.cells[yi][cx] = 0
				}
			case ModeRGB:
				n := g.cells[yi][cx]
				if n >= 4 {
					g.cells[yi][cx] = 0
				} else {
					g.cells[yi][cx] = n + 1
				}
			}
		}
		return
	}

	// kliknięcia poza siatką -> panel prawy/dolny
	px := GridW * CellSize
	py := 0
	bx := x - px
	by := y - py

	btnW := 180
	btnH := 34
	pad := 12

	click := func(bx0, by0 int) bool {
		return bx >= 0 && bx <= btnW && by >= by0 && by <= by0+btnH
	}

	// początkowa pozycja przycisków
	by0 := pad

	// 1: Tryb edycji (MONO/TWO/RGB)
	if click(bx, by0) {
		g.mode = (g.mode + 1) % 3
		return
	}
	by0 += btnH + pad

	// 2: Clear
	if click(bx, by0) {
		g.clear()
		return
	}
	by0 += btnH + pad

	// 3: Export: wybór 1bit/2bit/RGB
	if click(bx, by0) {
		g.exportMode = (g.exportMode + 1) % 3
		g.updatePreviewText()
		return
	}
	by0 += btnH + pad

	// 4: Format eksportu (C / PROGMEM)
	if click(bx, by0) {
		g.exportFormat = (g.exportFormat + 1) % 2
		g.updatePreviewText()
		return
	}
	by0 += btnH + pad

	// 5: Export C
	if click(bx, by0) {
		if err := g.exportC(); err != nil {
			log.Println("Export C failed:", err)
		} else {
			g.lastExport = "font.c"
		}
		return
	}
	by0 += btnH + pad

	// 6: Export PROGMEM
	if click(bx, by0) {
		if err := g.exportPROGMEM(); err != nil {
			log.Println("Export PROGMEM failed:", err)
		} else {
			g.lastExport = "font_progmem.c"
		}
		return
	}
	by0 += btnH + pad

	// 7: Export PNG
	if click(bx, by0) {
		if err := g.exportPNG(); err != nil {
			log.Println("Export PNG failed:", err)
		} else {
			g.lastExport = "exported.png"
		}
		return
	}
	by0 += btnH + pad

	// 8: toggle animacji
	if click(bx, by0) {
		g.animRunning = !g.animRunning
		if g.animRunning {
			g.animX = float64(GridW * CellSize)
		}
		return
	}
	by0 += btnH + pad

	// 9: toggle kierunku animacji
	if click(bx, by0) {
		g.animDir *= -1
		return
	}
	by0 += btnH + pad

	// 10: Generuj podgląd (HEX/BIN)
	if click(bx, by0) {
		g.updatePreviewText()
		return
	}
}

func (g *Game) clear() {
	for y := 0; y < GridH; y++ {
		for x := 0; x < GridW; x++ {
			g.cells[y][x] = 0
		}
	}
}

// -----------------------------
// funkcje eksportu i generatory
// -----------------------------

func rgb565(c color.RGBA) uint16 {
	r := uint16(c.R >> 3)
	g := uint16(c.G >> 2)
	b := uint16(c.B >> 3)
	return (r << 11) | (g << 5) | b
}

func (g *Game) exportPNG() error {
	img := image.NewRGBA(image.Rect(0, 0, GridW, GridH))
	for y := 0; y < GridH; y++ {
		for x := 0; x < GridW; x++ {
			idx := g.cells[y][x]
			c := palette[0]
			if idx >= 0 && idx < len(palette) {
				c = palette[idx]
			}
			img.SetRGBA(x, y, c)
		}
	}

	filename, err := chooseFilename("font_png", "c")
	if err != nil {
		filename = "export/exported.png"
		log.Println("chooseFilename failed, using fallback:", err)
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("cannot create file: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("png encode failed: %w", err)
	}

	g.lastExport = filename
	return nil
}

// eksport czystego C - zależnie od exportMode
func (g *Game) exportC() error {
	switch g.exportMode {
	case Export1Bit:
		return g.saveC1bit(false)
	case Export2Bit:
		return g.saveC2bit(false)
	case ExportRGB:
		return g.saveCRGB(false)
	}
	return nil
}

// eksport PROGMEM (Arduino)
func (g *Game) exportPROGMEM() error {
	switch g.exportMode {
	case Export1Bit:
		return g.saveC1bit(true)
	case Export2Bit:
		return g.saveC2bit(true)
	case ExportRGB:
		return g.saveCRGB(true)
	}
	return nil
}

// zapis 1-bit -> 1 bajt na wiersz (8 pikseli)
func (g *Game) saveC1bit(progmem bool) error {
	var b bytes.Buffer

	if progmem {
		b.WriteString(`#include <avr/pgmspace.h>
`)
	}

	b.WriteString("// 8x8 1-bit glyph\n")

	if progmem {
		b.WriteString("const uint8_t glyph8x8[] PROGMEM = {\n")
	} else {
		b.WriteString("uint8_t glyph8x8[] = {\n")
	}

	for y := 0; y < GridH; y++ {
		var row uint8 = 0
		for x := 0; x < GridW; x++ {
			if g.cells[y][x] != 0 {
				row |= 1 << (7 - x)
			}
		}
		b.WriteString(fmt.Sprintf("  0x%02X, // %08b\n", row, row))
	}

	b.WriteString("};\n")

	// Spróbuj dialogu
	filename, err := chooseFilename("font_1bit", "c")
	if err != nil {
		// fallback — zapisz do export/font_1bit.c
		filename = "export/font_1bit.c"
		log.Println("chooseFilename failed, using fallback:", err)
	}

	if err := os.WriteFile(filename, b.Bytes(), 0o644); err != nil {
		return fmt.Errorf("cannot write file: %w", err)
	}

	g.lastExport = filename
	return nil
}

// zapis 2-bit -> 2 bity na piksel -> 16 bit (uint16) na wiersz
func (g *Game) saveC2bit(progmem bool) error {
	var b bytes.Buffer

	if progmem {
		b.WriteString(`#include <avr/pgmspace.h>`)
	}

	b.WriteString("// 8x8 2-bit glyph (2 bity na piksel)\n")

	if progmem {
		b.WriteString("const uint16_t glyph8x8[] PROGMEM = {\n")
	} else {
		b.WriteString("uint16_t glyph8x8[] = {\n")
	}

	for y := 0; y < GridH; y++ {
		var row uint16 = 0
		for x := 0; x < GridW; x++ {
			v := uint16(g.cells[y][x] % 4)
			shift := uint((7 - x) * 2)
			row |= v << shift
		}
		b.WriteString(fmt.Sprintf("  0x%04X, // %016b\n", row, row))
	}

	b.WriteString("};\n")

	filename, err := chooseFilename("font_2bit", "c")
	if err != nil {
		filename = "export/font_2bit.c"
		log.Println("chooseFilename failed, using fallback:", err)
	}

	if err := os.WriteFile(filename, b.Bytes(), 0o644); err != nil {
		return fmt.Errorf("cannot write file: %w", err)
	}

	g.lastExport = filename
	return nil
}

// zapis RGB -> uint16_t RGB565 per pixel, zapis jako tablica [8][8]
func (g *Game) saveCRGB(progmem bool) error {
	var b bytes.Buffer

	if progmem {
		b.WriteString(`#include <avr/pgmspace.h>`)
	}

	b.WriteString("// 8x8 RGB565 glyph (uint16_t per pixel)\n")

	if progmem {
		b.WriteString("const uint16_t glyph8x8[8][8] PROGMEM = {\n")
	} else {
		b.WriteString("uint16_t glyph8x8[8][8] = {\n")
	}

	for y := 0; y < GridH; y++ {
		b.WriteString("  { ")
		for x := 0; x < GridW; x++ {
			idx := g.cells[y][x]
			c := palette[idx]
			rgb := rgb565(c)
			b.WriteString(fmt.Sprintf("0x%04X", rgb))
			if x < GridW-1 {
				b.WriteString(", ")
			}
		}
		b.WriteString(" },\n")
	}

	b.WriteString("};\n")

	filename, err := chooseFilename("font_rgb", "c")
	if err != nil {
		filename = "export/font_rgb.c"
		log.Println("chooseFilename failed, using fallback:", err)
	}

	if err := os.WriteFile(filename, b.Bytes(), 0o644); err != nil {
		return fmt.Errorf("cannot write file: %w", err)
	}

	g.lastExport = filename
	return nil
}

// uaktualnienie tekstowego podglądu (HEX / BIN) w zależności od ustawień eksportu
func (g *Game) updatePreviewText() {
	var sb strings.Builder

	sb.WriteString("// Podglad wygenerowanej tablicy\n")
	sb.WriteString(fmt.Sprintf("// tryb eksportu: %s | format: %s\n\n",
		g.exportModeLabel(), g.exportFormatLabel()))

	switch g.exportMode {
	case Export1Bit:
		for y := 0; y < GridH; y++ {
			var row uint8 = 0
			for x := 0; x < GridW; x++ {
				if g.cells[y][x] != 0 {
					row |= 1 << (7 - x)
				}
			}
			sb.WriteString(fmt.Sprintf("ROW %d: 0x%02X  bin:%08b\n", y, row, row))
		}

	case Export2Bit:
		for y := 0; y < GridH; y++ {
			var row uint16 = 0
			for x := 0; x < GridW; x++ {
				v := uint16(g.cells[y][x] % 4)
				shift := uint((7 - x) * 2)
				row |= v << shift
			}
			sb.WriteString(fmt.Sprintf("ROW %d: 0x%04X  bin:%016b\n", y, row, row))
		}

	case ExportRGB:
		for y := 0; y < GridH; y++ {
			for x := 0; x < GridW; x++ {
				idx := g.cells[y][x]
				c := palette[idx]
				rgb := rgb565(c)
				sb.WriteString(fmt.Sprintf("%02d,%02d: 0x%04X  ", x, y, rgb))
			}
			sb.WriteString("\n")
		}
	}

	g.previewText = sb.String()
}

func (g *Game) exportJSON() error {
	out := make([][]string, GridH)
	for y := 0; y < GridH; y++ {
		out[y] = make([]string, GridW)
		for x := 0; x < GridW; x++ {
			idx := g.cells[y][x]
			if idx >= 0 && idx < len(palette) {
				c := palette[idx]
				out[y][x] = fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
			} else {
				out[y][x] = "#000000"
			}
		}
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}

	filename := "export/font.json"
	if err := os.WriteFile(filename, b, 0o644); err != nil {
		return err
	}

	g.lastExport = filename
	return nil
}

// etykiety pomocnicze
func (g *Game) exportModeLabel() string {
	switch g.exportMode {
	case Export1Bit:
		return "1-bit"
	case Export2Bit:
		return "2-bit"
	case ExportRGB:
		return "RGB565"
	}
	return "?"
}
func (g *Game) exportFormatLabel() string {
	if g.exportFormat == 0 {
		return "C"
	}
	return "PROGMEM"
}

// funkcja pomocnicza zapisu plików
func chooseFilename(prefix, ext string) (string, error) {
	// Tworzenie katalogu
	if err := os.MkdirAll("export", 0755); err != nil {
		return "", fmt.Errorf("cannot create export directory: %w", err)
	}

	// Okno zapisu
	path, err := dialog.File().
		Title("Nazwa pliku eksportu").
		SetStartDir("export").
		Filter("Plik "+ext, "*."+ext).
		Save()
	if err != nil {
		return "", err
	}

	return path, nil
}

// Draw Rysowanie UI i podglądów
func (g *Game) Draw(screen *ebiten.Image) {
	// tło
	screen.Fill(color.RGBA{R: 0x10, G: 0x10, B: 0x12, A: 0xff})

	// siatka
	for y := 0; y < GridH; y++ {
		for x := 0; x < GridW; x++ {
			px := x * CellSize
			py := y * CellSize
			ebitenutil.DrawRect(screen, float64(px+1), float64(py+1), float64(CellSize-2), float64(CellSize-2), color.RGBA{R: 0x22, G: 0x22, B: 0x26, A: 0xff})
			idx := g.cells[y][x]
			if idx != 0 {
				c := palette[idx]
				ebitenutil.DrawRect(screen, float64(px+3), float64(py+3), float64(CellSize-6), float64(CellSize-6), c)
			}
			ebitenutil.DrawRect(screen, float64(px), float64(py), float64(CellSize), 1, color.RGBA{R: 0x44, G: 0x44, B: 0x50, A: 0xff})
			ebitenutil.DrawRect(screen, float64(px), float64(py+CellSize-1), float64(CellSize), 1, color.RGBA{R: 0x44, G: 0x44, B: 0x50, A: 0xff})
			ebitenutil.DrawRect(screen, float64(px), float64(py), 1, float64(CellSize), color.RGBA{R: 0x44, G: 0x44, B: 0x50, A: 0xff})
			ebitenutil.DrawRect(screen, float64(px+CellSize-1), float64(py), 1, float64(CellSize), color.RGBA{R: 0x44, G: 0x44, B: 0x50, A: 0xff})
		}
	}

	// Opis suwaka i aktualna wartość prędkości
	speedPercent := int(g.sliderValue * 100) // jeśli animSpeed = sliderValue*2
	drawText(screen, fmt.Sprintf("Prędkość animacji: %d%%", speedPercent), g.sliderX, g.sliderY-20)

	// SUWAK prędkości animacji nad panelem podglądu
	ebitenutil.DrawRect(screen, float64(g.sliderX), float64(g.sliderY), float64(g.sliderW), float64(g.sliderH), color.RGBA{R: 0x30, G: 0x30, B: 0x36, A: 0xff})
	handleX := g.sliderX + int(g.sliderValue*float64(g.sliderW))
	ebitenutil.DrawRect(screen, float64(handleX-4), float64(g.sliderY-2), 8, float64(g.sliderH)+4, color.RGBA{R: 0x2A, G: 0x80, B: 0xFF, A: 0xff})

	// panel po prawej - przyciski i informacje
	x0 := GridW * CellSize
	y0 := 0
	// prosty panel tła
	ebitenutil.DrawRect(screen, float64(x0), float64(y0), float64(CanvasW-x0), float64(GridH*CellSize), color.RGBA{R: 0x18, G: 0x18, B: 0x1C, A: 0xff})

	// rysuj przyciski
	btnW := 180
	btnH := 34
	pad := 12
	bx := x0 + pad
	by := y0 + pad

	// przycisk 1: tryb edycji
	ebitenutil.DrawRect(screen, float64(bx), float64(by), float64(btnW), float64(btnH), color.RGBA{R: 0x30, G: 0x30, B: 0x36, A: 0xff})
	//ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Tryb edycji: %s (M)", g.modeLabel()), bx+8, by+10)
	drawText(screen, fmt.Sprintf("Tryb:   %s (M)", g.modeLabel()), bx+8, by+23)

	// clear
	by += btnH + pad
	ebitenutil.DrawRect(screen, float64(bx), float64(by), float64(btnW), float64(btnH), color.RGBA{R: 0x30, G: 0x30, B: 0x36, A: 0xff})
	//ebitenutil.DebugPrintAt(screen, "Wyczyść (C)", bx+8, by+10)
	text.Draw(screen, "Wyczyść (C)", fontFace, bx+8, by+23, color.White)

	// export mode wybór
	by += btnH + pad
	ebitenutil.DrawRect(screen, float64(bx), float64(by), float64(btnW), float64(btnH), color.RGBA{R: 0x30, G: 0x30, B: 0x36, A: 0xff})
	drawText(screen, fmt.Sprintf("Eksport: %s", g.exportModeLabel()), bx+8, by+23)

	// format C / PROGMEM
	by += btnH + pad
	ebitenutil.DrawRect(screen, float64(bx), float64(by), float64(btnW), float64(btnH), color.RGBA{R: 0x30, G: 0x30, B: 0x36, A: 0xff})
	drawText(screen, fmt.Sprintf("Format: %s", g.exportFormatLabel()), bx+8, by+23)

	// Export C
	by += btnH + pad
	ebitenutil.DrawRect(screen, float64(bx), float64(by), float64(btnW), float64(btnH), color.RGBA{R: 0x2A, G: 0x80, B: 0xFF, A: 0xff})
	drawText(screen, "Eksportuj C", bx+8, by+23)

	// Export PROGMEM
	by += btnH + pad
	ebitenutil.DrawRect(screen, float64(bx), float64(by), float64(btnW), float64(btnH), color.RGBA{R: 0x2A, G: 0x80, B: 0xFF, A: 0xff})
	drawText(screen, "Eksportuj PROGMEM", bx+8, by+23)

	// Export PNG
	by += btnH + pad
	ebitenutil.DrawRect(screen, float64(bx), float64(by), float64(btnW), float64(btnH), color.RGBA{R: 0x2A, G: 0x80, B: 0xFF, A: 0xff})
	drawText(screen, "Eksportuj PNG", bx+8, by+26)

	// Animacja toggle
	by += btnH + pad
	col := color.RGBA{R: 0x60, G: 0x60, B: 0x60, A: 0xff}
	if g.animRunning {
		col = color.RGBA{R: 0x34, G: 0xC7, B: 0x34, A: 0xff}
	}
	ebitenutil.DrawRect(screen, float64(bx), float64(by), float64(btnW), float64(btnH), col)
	if g.animRunning {
		drawText(screen, "Stop animacji", bx+8, by+23)
	} else {
		drawText(screen, "Start animacji", bx+8, by+23)
	}
	by += btnH + pad
	ebitenutil.DrawRect(screen, float64(bx), float64(by), float64(btnW), float64(btnH), color.RGBA{R: 0x30, G: 0x30, B: 0x36, A: 0xff})
	dirText := "Kierunek: "
	if g.animDir < 0 {
		dirText += "←"
	} else {
		dirText += "→"
	}
	drawText(screen, dirText, bx+8, by+23)

	// Generuj podgląd
	by += btnH + pad
	ebitenutil.DrawRect(screen, float64(bx), float64(by), float64(btnW), float64(btnH), color.RGBA{R: 0x30, G: 0x30, B: 0x36, A: 0xff})
	drawText(screen, "Podgląd HEX/BIN", bx+8, by+23)

	// info o ostatnim eksporcie
	drawText(screen, fmt.Sprintf("Plik: %s", g.lastExport), bx, by+btnH+pad+10)

	// panel tekstowy z prawej - previewText
	textX := x0 + btnW + pad*2
	textY := pad
	// ramka
	w := CanvasW - textX - pad
	h := CanvasH - textY - pad - 120 // miejsce na animację + stopka
	ebitenutil.DrawRect(screen, float64(textX), float64(textY), float64(w), float64(h), color.RGBA{R: 0x0E, G: 0x0E, B: 0x10, A: 0xff})
	// wypisz tekst (linia po linii)
	lines := strings.Split(g.previewText, "\n")
	yline := textY + 6
	for i, ln := range lines {
		if i > 30 {
			break
		}
		// krótki wrap
		if len(ln) > 80 {
			ln = ln[:80] + "..."
		}
		ebitenutil.DebugPrintAt(screen, ln, textX+6, yline)
		yline += 12
	}

	// podgląd animacji – symulacja szerokiej matrycy 32x8
	playX := 12
	playY := GridH*CellSize + 280

	scaleW := 0.78 // dopasowanie do okna
	scaleH := 2.0  // 2x wysokość 8x8

	playW := int(float64(GridW*CellSize*4) * scaleW)
	playH := int(float64(CellSize) * scaleH)

	ebitenutil.DrawRect(
		screen,
		float64(playX),
		float64(playY),
		float64(playW),
		float64(playH),
		color.RGBA{R: 0x08, G: 0x08, B: 0x08, A: 0xff},
	)

	// narysuj przewijający się glyph (każdy "piksel" symulowany jako box o szerokości CellSize)
	// animX kontroluje przesunięcie w px
	// -- zmiana rozmiaru do 1/4  edytora
	{
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
					ebitenutil.DrawRect(
						screen,
						float64(px+1), float64(py+1),
						float64(animSize-2), float64(animSize-2),
						c,
					)
				}
			}
		}
	}

	// krótka pomoc u dołu
	help := "LPM: kliknij komórkę aby zmienić; M: zmiana trybu; C: wyczyść. Kliknij przyciski po prawej."
	drawText(screen, help, 8, CanvasH-26)
}

func (g *Game) modeLabel() string {
	switch g.mode {
	case ModeMono:
		return "MONO"
	case ModeTwo:
		return "2-KOLORY"
	case ModeRGB:
		return "RGB"
	}
	return "?"
}

// pomocnik: zmiana tekstów na UNIKODE
func drawText(screen *ebiten.Image, s string, x, y int) {
	text.Draw(screen, s, fontFace, x, y, color.White)
}

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
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
