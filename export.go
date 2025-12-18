/*
Autor: SunRiver / Lothar TeaM
  WWW: https://forum.lothar-team.pl/
  Git: https://github.com/SunDUINO/Sun8x8_Font_generator.git
 Plik: export.go

*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"strings"
)

// -----------------------------
// funkcje eksportu i generatory
// -----------------------------

// eksport czystego C - zależnie od exportMode
func (g *Game) exportC() error {

	text := GenerateCFromGlyphs(
		g.glyphs,
		g.exportMode,
		false)

	g.previewText = text

	err := os.MkdirAll("export", os.ModePerm)
	if err != nil {
		return err
	}
	filename := "export/znaki.h"
	if err := os.WriteFile(filename, []byte(text), 0o644); err != nil {
		g.lastExport = "Błąd zapisu"
		return err
	}

	g.lastExport = filename
	return nil
}

// eksport PROGMEM (Arduino)
func (g *Game) exportPROGMEM() error {

	text := GenerateCFromGlyphs(
		g.glyphs,
		g.exportMode,
		true,
	)

	g.previewText = text

	err := os.MkdirAll("export", os.ModePerm)
	if err != nil {
		return err
	}
	filename := "export/znaki_progmem.h"

	if err := os.WriteFile(filename, []byte(text), 0o644); err != nil {
		g.lastExport = "Błąd zapisu"
		return err
	}

	g.lastExport = filename
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
		var row uint8
		for x := 0; x < GridW; x++ {
			if g.cells[y][x] != 0 {
				row |= 1 << (7 - x)
			}
		}
		b.WriteString(fmt.Sprintf("  0x%02X, // %08b\n", row, row))
	}

	b.WriteString("};\n")

	filename, err := chooseFilename("font_1bit", "c")
	if err != nil {
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
		b.WriteString(`#include <avr/pgmspace.h>
`)
	}

	b.WriteString("// 8x8 2-bit glyph (2 bity na piksel)\n")

	if progmem {
		b.WriteString("const uint16_t glyph8x8[] PROGMEM = {\n")
	} else {
		b.WriteString("uint16_t glyph8x8[] = {\n")
	}

	for y := 0; y < GridH; y++ {
		var row uint16
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
		b.WriteString(`#include <avr/pgmspace.h>
`)
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

// -----------------------------
// eksport do PNG
// -----------------------------

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

	filename, err := chooseFilename("font_png", "png")
	if err != nil {
		filename = "export/exported.png"
		log.Println("chooseFilename failed, using fallback:", err)
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("cannot create file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			log.Printf("warning: failed to close file %s: %v", filename, cerr)
		}
	}()

	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("png encode failed: %w", err)
	}

	g.lastExport = filename
	return nil
}

// -----------------------------
// eksport do JSON
// -----------------------------

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

// -----------------------------
// uaktualnienie tekstowego podglądu (HEX / BIN)
// -----------------------------

func (g *Game) updatePreviewText() {
	var sb strings.Builder

	sb.WriteString("// Podglad wygenerowanej tablicy\n")
	sb.WriteString(fmt.Sprintf(
		"// tryb eksportu: %s | format: %s\n\n",
		g.exportModeLabel(),
		g.exportFormatLabel(),
	))

	switch g.exportMode {

	case Export1Bit:
		for y := 0; y < GridH; y++ {
			var row uint8
			for x := 0; x < GridW; x++ {
				if g.cells[y][x] != 0 {
					row |= 1 << (7 - x)
				}
			}
			sb.WriteString(fmt.Sprintf(
				"ROW %d: 0x%02X  bin:%08b\n",
				y, row, row,
			))
		}

	case Export2Bit:
		for y := 0; y < GridH; y++ {
			var row uint16
			for x := 0; x < GridW; x++ {
				v := uint16(g.cells[y][x] % 4)
				shift := uint((7 - x) * 2)
				row |= v << shift
			}
			sb.WriteString(fmt.Sprintf(
				"ROW %d: 0x%04X  bin:%016b\n",
				y, row, row,
			))
		}

	case ExportRGB:
		for y := 0; y < GridH; y++ {
			for x := 0; x < GridW; x++ {
				idx := g.cells[y][x]
				c := palette[idx]
				rgb := rgb565(c)
				sb.WriteString(fmt.Sprintf(
					"%02d,%02d: 0x%04X  ",
					x, y, rgb,
				))
			}
			sb.WriteString("\n")
		}
	}

	g.previewText = sb.String()
}

// -----------------------------
// etykiety pomocnicze
// -----------------------------

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

func GenerateCFromGlyphs(glyphs [][]byte, exportMode int, progmem bool) string {
	var b strings.Builder

	if progmem {
		b.WriteString("#include <avr/pgmspace.h>\n\n")
	}

	b.WriteString("// Generated by Sun8x8 Font Generator\n\n")

	n := len(glyphs) // liczba wygenerowanych znaków

	switch exportMode {
	case Export1Bit:
		if progmem {
			b.WriteString(fmt.Sprintf("const uint8_t font[%d][8] PROGMEM = {\n", n))
		} else {
			b.WriteString(fmt.Sprintf("const uint8_t font[%d][8] = {\n", n))
		}

		for _, g := range glyphs {
			b.WriteString("  { ")
			for y := 0; y < 8; y++ {
				if y > 0 {
					b.WriteString(", ")
				}
				b.WriteString(fmt.Sprintf("0x%02X", g[y]))
			}
			b.WriteString(" },\n")
		}
	case Export2Bit, ExportRGB:
		if progmem {
			b.WriteString(fmt.Sprintf("const uint16_t font[%d][8] PROGMEM = {\n", n))
		} else {
			b.WriteString(fmt.Sprintf("const uint16_t font[%d][8] = {\n", n))
		}

		for _, g := range glyphs {
			b.WriteString("  { ")
			for y := 0; y < 8; y++ {
				if y > 0 {
					b.WriteString(", ")
				}
				b.WriteString(fmt.Sprintf("0x%04X", uint16(g[y])))
			}
			b.WriteString(" },\n")
		}
	}

	b.WriteString("};\n")
	return b.String()
}
