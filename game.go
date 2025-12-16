package main

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

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
)

// eksport: bit-depth
const (
	Export1Bit = iota
	Export2Bit
	ExportRGB
)

const AnimCellScale = 0.25

// Game struktura gry / edytora
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

var matrixSerial *SerialMatrix

func NewGame() *Game {
	g := &Game{}
	g.mode = ModeMono
	g.monoColor = 1
	g.twoA = 2
	g.twoB = 3
	g.exportMode = Export1Bit

	g.sliderX = 12
	g.sliderY = GridH*CellSize + 260
	g.sliderW = 300
	g.sliderH = 14
	g.sliderValue = 0.5

	g.animSpeed = 0.5
	g.animDir = -1

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
			if !g.handleSlider(x, y) {
				g.handleLeftClick(x, y)
			}
		} else if g.sliderGrabbed {
			g.sliderValue = float64(x-g.sliderX) / float64(g.sliderW)
			if g.sliderValue < 0 {
				g.sliderValue = 0
			}
			if g.sliderValue > 1 {
				g.sliderValue = 1
			}
			g.animSpeed = g.sliderValue
		}
	} else {
		g.mouseDown = false
		g.sliderGrabbed = false
	}

	// ---- po obsłudze kliknięć wysyłamy ramkę do matrycy ----

	if matrixSerial != nil {
		err := matrixSerial.SendFrame(g.buildDisplayFrame())
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
		g.mode = (g.mode + 1) % 3
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
		g.mode = (g.mode + 1) % 3
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

func (g *Game) buildDisplayFrame() [][]int {
	frame := make([][]int, 8)
	for y := 0; y < 8; y++ {
		frame[y] = make([]int, 32) // 4 matryce po 8 kolumn
	}

	// M0 – aktualnie edytowany znak
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			frame[y][x] = g.cells[y][x]
		}
	}

	// M1–M3 – zapisane znaki
	for i, glyph := range g.displayGlyphs {
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

func (g *Game) modeLabel() string {
	switch g.mode {
	case ModeMono:
		return "MONO"
	case ModeTwo:
		return "2-BIT"
	case ModeRGB:
		return "RGB"
	}
	return "?"
}
