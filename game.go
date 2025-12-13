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

}

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

	if ebiten.IsKeyPressed(ebiten.KeyM) {
		g.mode = (g.mode + 1) % 3
		time.Sleep(140 * time.Millisecond)
	}
	if ebiten.IsKeyPressed(ebiten.KeyC) {
		g.clear()
		time.Sleep(120 * time.Millisecond)
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
	by := y

	btnW = 180
	btnH = 34
	pad := 12

	click := func(by0 int) bool {
		return bx >= 0 && bx <= btnW &&
			by >= by0 && by <= by0+btnH
	}

	by0 := pad

	if click(by0) {
		g.mode = (g.mode + 1) % 3
		return
	}
	by0 += btnH + pad

	if click(by0) {
		g.clear()
		return
	}
	by0 += btnH + pad

	if click(by0) {
		g.exportMode = (g.exportMode + 1) % 3
		g.updatePreviewText()
		return
	}
	by0 += btnH + pad

	if click(by0) {
		g.exportFormat = (g.exportFormat + 1) % 2
		g.updatePreviewText()
		return
	}
	by0 += btnH + pad

	if click(by0) {
		_ = g.exportC()
		return
	}
	by0 += btnH + pad

	if click(by0) {
		_ = g.exportPROGMEM()
		return
	}
	by0 += btnH + pad

	if click(by0) {
		_ = g.exportPNG()
		return
	}
	by0 += btnH + pad

	if click(by0) {
		g.animRunning = !g.animRunning
		if g.animRunning {
			g.animX = float64(GridW * CellSize)
		}
		return
	}
	by0 += btnH + pad

	if click(by0) {
		g.animDir *= -1
		return
	}
	by0 += btnH + pad

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
