package main

import (
	"fmt"
	"net/http"
	"runtime/debug"

	log "github.com/sirupsen/logrus"
)

func (app *application) logError(w http.ResponseWriter, status int, err error, withTrace bool) {
	w.Header().Set("Connection", "close")
	errorText := err.Error()
	log.Error(errorText)
	if withTrace {
		log.Errorf("%s", debug.Stack())
	}

	w.Header().Set("Content-Type", "application/text")

	switch status {
	case http.StatusForbidden:
		http.Error(w, `forbidden`, status)
	case http.StatusBadRequest:
		http.Error(w, err.Error(), status)
	case http.StatusInternalServerError:
		http.Error(w, `There was an error processing your request`, status)
	default:
		http.Error(w, `There was an error processing your request.`, status)
	}
}

func (app *application) notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	if _, err := w.Write([]byte(`not found`)); err != nil {
		log.Infof("could not send 404: %v", err)
		return
	}
}

func (app *application) index(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "home.page.tmpl", &templateData{
		MaxCodeLen: maxCodeLen,
	})
}

func (app *application) generateImages(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		app.logError(w, http.StatusBadRequest, fmt.Errorf("missing code"), false)
		return
	}

	if len(code) > maxCodeLen {
		app.logError(w, http.StatusBadRequest, fmt.Errorf("code too long"), false)
		return
	}

	normalImg, err := app.putTextOnImage(code, modeNormal)
	if err != nil {
		app.logError(w, http.StatusBadRequest, fmt.Errorf("could not generate image: %v", err), false)
		return
	}
	devaluedImg, err := app.putTextOnImage(code, modeDevalued)
	if err != nil {
		app.logError(w, http.StatusBadRequest, fmt.Errorf("could not generate image: %v", err), false)
		return
	}

	app.render(w, r, "images.page.tmpl", &templateData{
		NormalImage:   normalImg,
		DevaluedImage: devaluedImg,
	})
}
