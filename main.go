package main

import (
	"embed"
	"image"
	"sync"
	"text/template"

	"github.com/golang/freetype/truetype"
	log "github.com/sirupsen/logrus"
	"golang.org/x/image/font"
)

type application struct {
	port          int
	wg            sync.WaitGroup
	imageNormal   *image.Image
	imageDevalued *image.Image
	font          *font.Face
	templateCache map[string]*template.Template
}

//go:embed assets
//go:embed templates
var content embed.FS

const (
	fontSize   = 32
	fontDPI    = 72
	maxCodeLen = 7
	port       = 8080
)

func main() {
	// fontBytes, err := content.ReadFile("assets/PermanentMarker-Regular.ttf")
	fontBytes, err := content.ReadFile("assets/NanumPenScript-Regular.ttf")
	if err != nil {
		log.Println(err)
		return
	}
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		log.Println(err)
		return
	}
	font := truetype.NewFace(f, &truetype.Options{
		Size: fontSize,
		DPI:  fontDPI,
	})

	fileNormal, err := content.Open("assets/normal.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer fileNormal.Close()
	imageNormal, _, err := image.Decode(fileNormal)
	if err != nil {
		log.Fatal(err)
	}

	fileDevalued, err := content.Open("assets/devalued.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer fileDevalued.Close()
	imageDevalued, _, err := image.Decode(fileDevalued)
	if err != nil {
		log.Fatal(err)
	}

	templateCache, err := newTemplateCache()
	if err != nil {
		log.Fatal(err)
	}

	app := application{
		port:          port,
		font:          &font,
		imageNormal:   &imageNormal,
		imageDevalued: &imageDevalued,
		templateCache: templateCache,
	}

	if err := app.serve(); err != nil {
		log.Fatal(err)
	}
}
