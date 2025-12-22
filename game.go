/*
Autor: SunRiver / Lothar TeaM
  WWW: https://forum.lothar-team.pl/
  Git: https://github.com/SunDUINO/Sun8x8_Font_generator.git
 Plik: game.go

*/

package main

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// siatka 8x8 edytora znaków
const (
	GridW    = 8
	GridH    = 8
	CellSize = 28
	CanvasW  = GridW*CellSize + 500
	CanvasH  = GridH*CellSize + 400
)

// tryby edycji
const (
	ModeMono = iota
	ModeTwo
	ModeRGB
	ModeWS2812B
)

// eksport: bit-depth
const (
	Export1Bit = iota
	Export2Bit
	ExportRGB
)

const AnimCellScale = 0.25

// Game struktura edytora
type Game struct {
	cells [GridH][GridW]int

	mode      int
	monoColor int
	twoA      int
	twoB      int

	mouseDown bool

	// eksport
	exportMode   int
	exportFormat int
	lastExport   string
	previewText  string

	// animacja
	animX       float64
	animRunning bool
	animSpeed   float64
	animDir     int

	// suwak
	sliderX       int
	sliderY       int
	sliderW       int
	sliderH       int
	sliderValue   float64
	sliderGrabbed bool

	// NOWY SUWAK: kolor mono
	colorSliderX       int
	colorSliderY       int
	colorSliderW       int
	colorSliderH       int
	colorSliderValue   float64
	colorSliderGrabbed bool

	// font builder
	glyphs     [][]byte
	glyphIndex int

	// --- scroll glyphów ---
	glyphScroll int
	glyphViewH  int

	// --- scroll tekstu preview ---
	textScroll int
	textViewH  int

	cellScroll int // scroll siatki edytora

	serialStatus string

	activeGlyph   int      // aktualnie wybrany znak
	glyphViewOfs  int      // offset podglądu (0,1,2... → okno 8 znaków)
	displayGlyphs [][]byte // max 4 glyphy po 8 bajtów
}

func NewGame() *Game {
	g := &Game{}
	g.mode = ModeMono

	g.monoColor = 2 // kolor pixeli w siatce --> index zgodny z paletą w pliku palette.go
	g.twoA = 2
	g.twoB = 3
	g.exportMode = Export1Bit

	g.sliderX = 12
	g.sliderY = GridH*CellSize + 260
	g.sliderW = 200
	g.sliderH = 14
	g.sliderValue = 0.5

	g.animSpeed = 0.5
	g.animDir = -1

	// SUWAK: kolor mono
	g.colorSliderX = 12
	g.colorSliderY = GridH*CellSize + 68 + 32 + 8 + 10 // btnY + btnH + odstęp między przyciskami + dodatkowe pixele dla labela
	g.colorSliderW = 200
	g.colorSliderH = 14
	g.colorSliderValue = float64(g.monoColor) / float64(len(palette)-1)

	// ------ tutaj inicjalizacja serial -------
	//matrixSerial = NewSerialMatrix("COM13", GridW, GridH)
	port := detectSerialPort()
	matrixSerial = NewSerialMatrix(port, GridW, GridH)

	if matrixSerial != nil {
		g.serialStatus = "Podłączono do: " + port
	} else {
		g.serialStatus = "Brak połączenia z Pico"
	}

	g.activeGlyph = -1
	g.glyphViewOfs = 0

	return g
}

func (g *Game) Layout(_, _ int) (int, int) {
	return CanvasW, CanvasH
}

func (g *Game) Update() error {
	x, y := ebiten.CursorPosition()

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if !g.mouseDown {
			g.mouseDown = true
			// sprawdzanie wszystkich suwaków
			if !g.handleSlider(x, y) { // animacja
				// NOWY SUWAK: kolor mono
				if x >= g.colorSliderX && x <= g.colorSliderX+g.colorSliderW &&
					y >= g.colorSliderY && y <= g.colorSliderY+g.colorSliderH {
					g.colorSliderGrabbed = true
				} else {
					g.handleLeftClick(x, y)
				}
			}
		} else {
			// przeciąganie suwaków
			if g.sliderGrabbed {
				g.sliderValue = float64(x-g.sliderX) / float64(g.sliderW)
				if g.sliderValue < 0 {
					g.sliderValue = 0
				}
				if g.sliderValue > 1 {
					g.sliderValue = 1
				}
				g.animSpeed = g.sliderValue
			}
			if g.colorSliderGrabbed {
				g.colorSliderValue = float64(x-g.colorSliderX) / float64(g.colorSliderW)
				if g.colorSliderValue < 0 {
					g.colorSliderValue = 0
				}
				if g.colorSliderValue > 1 {
					g.colorSliderValue = 1
				}
				g.monoColor = int(g.colorSliderValue * float64(len(palette)-1))
			}
		}
	} else {
		g.mouseDown = false
		g.sliderGrabbed = false
		g.colorSliderGrabbed = false // "przyklejenie" suwaka koloru
	}

	// ---- po obsłudze kliknięć wysyłamy ramkę do matrycy ----

	if matrixSerial != nil {
		var err error
		if g.mode == ModeWS2812B {
			err = matrixSerial.SendFrame(g.buildWS2812Frame())
		} else {
			err = matrixSerial.SendFrame(g.buildDisplayFrame())
		}
		if err != nil {
			matrixSerial.Close()
			matrixSerial = nil
			g.serialStatus = "Odłączono"
		}
	}

	// jeśli brak połączenia, próbuj reconnect
	if matrixSerial == nil {
		port := detectSerialPort() // twoja funkcja wykrywająca COM
		if port != "" {
			matrixSerial = NewSerialMatrix(port, GridW, GridH)
			if matrixSerial != nil {
				g.serialStatus = "Podłączono " + port
			} else {
				g.serialStatus = "Nie można otworzyć portu"
			}
		} else {
			g.serialStatus = "Brak Pico"
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyM) {
		g.mode = (g.mode + 1) % 4
		time.Sleep(140 * time.Millisecond)
	}
	if ebiten.IsKeyPressed(ebiten.KeyC) {
		g.clear()
		time.Sleep(120 * time.Millisecond)
	}
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		err := g.exportC()
		if err != nil {
			g.lastExport = "Błąd zapisu pliku"
		} else {
			g.lastExport = "Plik znaki.h zapisany w katalogu export"
		}
		time.Sleep(200 * time.Millisecond) // debounce
	}

	if g.animRunning {
		g.animX += float64(g.animDir) * g.animSpeed * 40.0
		if g.animDir < 0 && g.animX < -float64(GridW*CellSize) {
			g.animX = float64(GridW * CellSize)
		} else if g.animDir > 0 && g.animX > float64(GridW*CellSize) {
			g.animX = -float64(GridW * CellSize)
		}

	}

	return nil
}

func (g *Game) clear() {
	for y := 0; y < GridH; y++ {
		for x := 0; x < GridW; x++ {
			g.cells[y][x] = 0
		}
	}
}

// obsługa slidera prędkości animacji
func (g *Game) handleSlider(x, y int) bool {
	if x >= g.sliderX && x <= g.sliderX+g.sliderW &&
		y >= g.sliderY && y <= g.sliderY+g.sliderH {
		g.sliderGrabbed = true
		return true
	}
	return false
}

// obsługa kliknięć: siatka i przyciski
func (g *Game) handleLeftClick(x, y int) {

	// NOWY SUWAK: kliknięcie na suwak koloru
	if x >= g.colorSliderX && x <= g.colorSliderX+g.colorSliderW &&
		y >= g.colorSliderY && y <= g.colorSliderY+g.colorSliderH {
		g.colorSliderGrabbed = true
		return
	}

	// --- przycisk pod gridem: zapisz znak ---
	btnX := 12
	btnY := GridH*CellSize + 12
	btnW := GridW*CellSize - 24
	btnH := 32

	if x >= btnX && x <= btnX+btnW &&
		y >= btnY && y <= btnY+btnH {

		g.addGlyph()
		return
	}

	// --- przyciski przewijania zapisanych znaków ---
	btnPrevX := btnX
	btnPrevY := btnY + btnH + 8
	btnPrevW := 60
	btnPrevH := 28

	btnNextX := btnPrevX + btnPrevW + 8
	btnNextY := btnPrevY
	btnNextW := 60
	btnNextH := 28

	// Poprzedni znak
	if x >= btnPrevX && x <= btnPrevX+btnPrevW &&
		y >= btnPrevY && y <= btnPrevY+btnPrevH {
		//fmt.Println("Poprzedni klik:", x, y)
		if g.activeGlyph > 0 {
			g.activeGlyph--
			g.glyphIndex = g.activeGlyph
			g.loadGlyphToGrid(g.activeGlyph)

			// przesuwamy widok, jeśli aktywny znak wychodzi poza 8-znakowy widok
			if g.activeGlyph < g.glyphViewOfs {
				g.glyphViewOfs = g.activeGlyph
			}
		}
		return
	}

	// Następny znak
	if x >= btnNextX && x <= btnNextX+btnNextW &&
		y >= btnNextY && y <= btnNextY+btnNextH {
		//fmt.Println("Następny klik:", x, y)
		if g.activeGlyph < len(g.glyphs)-1 {
			g.activeGlyph++
			g.glyphIndex = g.activeGlyph
			g.loadGlyphToGrid(g.activeGlyph)

			// przesuwamy widok jeśli aktywny znak wychodzi poza 8-znakowy widok
			if g.activeGlyph >= g.glyphViewOfs+8 {
				g.glyphViewOfs = g.activeGlyph - 7
			}
		}
		return
	}

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
			case ModeWS2812B:
				if g.cells[yi][cx] == g.monoColor {
					g.cells[yi][cx] = 0 // wyłącz diodę jeśli ma ten sam kolor
				} else {
					g.cells[yi][cx] = g.monoColor // ustaw kolor z suwaka
				}
			}

		}
		return
	}

	// panel prawy
	px := GridW * CellSize
	bx := x - px
	by0 := 12 // początek pierwszego przycisku
	btnW = 180
	btnH = 34
	pad := 12

	click := func(by int) bool {
		return bx >= 0 && bx <= btnW && y >= by && y <= by+btnH
	}

	// 1. Tryb
	if click(by0) {
		g.mode = (g.mode + 1) % 4
		return
	}
	by0 += btnH + pad

	// 2. Clear
	if click(by0) {
		g.clear()
		return
	}
	by0 += btnH + pad

	// 3. Eksport: tryb
	if click(by0) {
		g.exportMode = (g.exportMode + 1) % 3
		g.updatePreviewText()
		return
	}
	by0 += btnH + pad

	// 4. Eksport: format
	if click(by0) {
		g.exportFormat = (g.exportFormat + 1) % 2
		g.updatePreviewText()
		return
	}
	by0 += btnH + pad

	// 5. Eksport C
	if click(by0) {
		if err := g.exportC(); err != nil {
			g.lastExport = "Błąd eksportu C"
		}
		return

	}
	by0 += btnH + pad

	// 6. Eksport PROGMEM
	if click(by0) {
		if err := g.exportPROGMEM(); err != nil {
			g.lastExport = "Błąd eksportu PROGMEM"
		}
		return
	}
	by0 += btnH + pad

	// 7. Eksport PNG
	if click(by0) {
		if err := g.exportPNG(); err != nil {
			g.lastExport = "Błąd eksportu PNG"
		}
		return
	}
	by0 += btnH + pad

	// 8. Animacja start/stop
	if click(by0) {
		g.animRunning = !g.animRunning
		if g.animRunning {
			g.animX = float64(GridW * CellSize)
		}
		return
	}
	by0 += btnH + pad

	// 9. Zmiana kierunku animacji
	if click(by0) {
		g.animDir *= -1
		return
	}
	by0 += btnH + pad

	// 10. Podgląd HEX/BIN
	if click(by0) {
		g.updatePreviewText()
		return
	}

	_, wy := ebiten.Wheel()
	if wy != 0 {
		x, y := ebiten.CursorPosition()

		// --- obszar glyph preview ---
		if g.isInGlyphPreview(x, y) {
			g.glyphScroll -= int(wy) * 16
			if g.glyphScroll < 0 {
				g.glyphScroll = 0
			}
		}

		// --- obszar tekstowego preview ---
		if g.isInTextPreview(x, y) {
			g.textScroll -= int(wy) * 16
			if g.textScroll < 0 {
				g.textScroll = 0
			}
		}
	}

}

func (g *Game) cellsToSlice() [][]int {
	s := make([][]int, GridH)
	for y := 0; y < GridH; y++ {
		s[y] = make([]int, GridW)
		for x := 0; x < GridW; x++ {
			s[y][x] = g.cells[y][x]
		}
	}
	return s
}

// buildDisplayFrame tworzy pełną ramkę do wyświetlenia na matrycach
func (g *Game) buildDisplayFrame() [][]int {
	frame := make([][]int, 8)
	for y := 0; y < 8; y++ {
		frame[y] = make([]int, 32) // 4 matryce po 8 kolumn
	}

	// M0 – aktualnie edytowany znak (kolumny 0–7)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			frame[y][x] = g.cells[y][x]
		}
	}

	// M1–M3 – zapisane znaki (wykorzystujemy tylko g.displayGlyphs)
	for i, glyph := range g.displayGlyphs {
		if i >= 3 {
			break
		}
		for y := 0; y < 8; y++ {
			b := glyph[y]
			for bit := 0; bit < 8; bit++ {
				if (b & (1 << (7 - bit))) != 0 {
					frame[y][(i+1)*8+bit] = 1
				}
			}
		}
	}

	return frame
}

// buildWS2812Frame buduje ramkę dla matrycy WS2812B w układzie zig-zag 8x8
func (g *Game) buildWS2812Frame() [][]int {
	frame := make([][]int, 8)
	for y := 0; y < 8; y++ {
		frame[y] = make([]int, 8)
	}

	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			// określenie fizycznego indeksu kolumny w wierszu zig-zag
			idxX := x
			if y%2 == 1 { // co drugi wiersz od prawej do lewej
				idxX = 7 - x
			}

			// kolor z komórki (0 = wyłączona, >0 = kolor z suwaka)
			frame[y][idxX] = g.cells[y][x]
		}
	}

	return frame
}

// loadGlyphToGrid >> wczytuje znak do siatki
func (g *Game) loadGlyphToGrid(idx int) {
	if idx < 0 || idx >= len(g.glyphs) {
		return
	}
	glyph := g.glyphs[idx]
	for y := 0; y < GridH; y++ {
		row := glyph[y]
		for x := 0; x < GridW; x++ {
			if (row & (1 << (7 - x))) != 0 {
				g.cells[y][x] = 1
			} else {
				g.cells[y][x] = 0
			}
		}
	}

	// odśwież wyświetlane glyphy w matrycach M1–M3
	g.updateDisplayGlyphs()
}

// aktualizacja wyświetlanego znaku
func (g *Game) updateDisplayGlyphs() {
	g.displayGlyphs = [][]byte{}
	for i := 0; i < 3; i++ {
		idx := g.activeGlyph + i - 2 // M1 = poprzedni, M2 = kolejny...
		if idx >= 0 && idx < len(g.glyphs) && idx != g.activeGlyph {
			g.displayGlyphs = append(g.displayGlyphs, g.glyphs[idx])
		}
	}
}

// labelki
func (g *Game) modeLabel() string {
	switch g.mode {
	case ModeMono:
		return "MONO"
	case ModeTwo:
		return "2-BIT"
	case ModeRGB:
		return "RGB"
	case ModeWS2812B:
		return "WS2812B"
	}
	return "?"
}
