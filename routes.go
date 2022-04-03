package main

import (
	"embed"
	"net/http"

	"github.com/gorilla/mux"
)

//go:embed static
var static embed.FS

func (app *application) routes() http.Handler {
	router := mux.NewRouter()
	router.Use(app.loggingMiddleware)
	router.Use(app.recoverPanic)

	router.HandleFunc("/", app.index)
	router.HandleFunc("/generateImages", app.generateImages)
	router.PathPrefix("/static").Handler(http.FileServer(http.FS(static)))

	// catch all on the end to log all requests
	router.PathPrefix("/").HandlerFunc(app.notFound)

	http.Handle("/", router)

	return router
}
