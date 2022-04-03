package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
)

func (app *application) loggingMiddleware(next http.Handler) http.Handler {
	return handlers.CustomLoggingHandler(os.Stdout, next, customLogFormatter)
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				app.logError(w, http.StatusInternalServerError, fmt.Errorf("%s", err), true)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
