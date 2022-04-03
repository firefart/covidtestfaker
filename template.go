package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"text/template"
)

type templateData struct {
	NormalImage   *[]byte
	DevaluedImage *[]byte
	MaxCodeLen    int
}

// newTemplateCache holds all templates so we have to parse them
// only once on application start
func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	// add all page templates
	pages, err := fs.Glob(content, "templates/*.page.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFS(content, page)
		if err != nil {
			return nil, err
		}

		// include needed layout templates
		ts, err = ts.Funcs(functions).ParseFS(content, "templates/*.layout.tmpl")
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}
	return cache, nil
}

// render is called to render the specified template
func (app *application) render(w http.ResponseWriter, r *http.Request, name string, td *templateData) {
	ts, ok := app.templateCache[name]
	if !ok {
		app.logError(w, http.StatusInternalServerError, fmt.Errorf("the template %s does not exist", name), false)
		return
	}

	buf := new(bytes.Buffer)

	err := ts.Execute(buf, td)
	if err != nil {
		app.logError(w, http.StatusInternalServerError, err, false)
		return
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		app.logError(w, http.StatusInternalServerError, fmt.Errorf("error on writing: %w", err), false)
		return
	}
}

var functions = template.FuncMap{
	"base64": base64Encode,
}

func base64Encode(content []byte) (string, error) {
	return base64.StdEncoding.EncodeToString(content), nil
}
