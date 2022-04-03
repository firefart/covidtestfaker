package main

import (
	"bytes"
	"fmt"
	"image"
	"math/rand"

	"github.com/fogleman/gg"
)

type imageMode int

const (
	modeNormal   imageMode = 1
	modeDevalued imageMode = 2
)

func getRandomColor() (float64, float64, float64, float64) {
	r := rand.Float64()
	g := rand.Float64()
	b := rand.Float64()
	a := rand.Float64()
	return r, g, b, a
}

func (app *application) putTextOnImage(text string, mode imageMode) (*[]byte, error) {
	var img image.Image
	var posX, posY float64
	switch mode {
	case modeNormal:
		img = *app.imageNormal
		posX = 120.0
		posY = 150.0
	case modeDevalued:
		img = *app.imageDevalued
		posX = 120.0
		posY = 155.0
	default:
		return nil, fmt.Errorf("unknown mode %d", mode)
	}

	dc := gg.NewContextForImage(img)
	dc.DrawImage(img, 0, 0)
	dc.SetRGBA255(0, 0, 114, 200)
	dc.SetFontFace(*app.font)
	dc.DrawStringAnchored(text, posX, posY, 0.5, 0.5)
	// change first row of pixels to random rgba to circumvent
	// hash checking of the images
	for i := 0; i < dc.Width(); i++ {
		r, g, b, a := getRandomColor()
		dc.SetRGBA(r, g, b, a)
		dc.SetPixel(i, 0)
	}
	dc.Clip()
	buffer := new(bytes.Buffer)
	if err := dc.EncodePNG(buffer); err != nil {
		return nil, err
	}

	b := buffer.Bytes()
	return &b, nil
}
