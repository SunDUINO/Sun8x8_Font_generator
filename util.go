package main

import (
	"fmt"
	"image/color"
	"os"

	"github.com/sqweek/dialog"
)

func rgb565(c color.RGBA) uint16 {
	r := uint16(c.R >> 3)
	g := uint16(c.G >> 2)
	b := uint16(c.B >> 3)
	return (r << 11) | (g << 5) | b
}

func chooseFilename(prefix, ext string) (string, error) {
	if err := os.MkdirAll("export", 0755); err != nil {
		return "", err
	}

	path, err := dialog.File().
		Title("Nazwa pliku eksportu").
		SetStartDir("export").
		Save()
	if err != nil {
		return "", err
	}

	if path == "" {
		path = fmt.Sprintf("export/%s.%s", prefix, ext)
	}
	return path, nil
}
