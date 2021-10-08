package main

import (
	"fmt"
	"net/http"
)


func (app *application) Auth(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		saved, err := app.authenticateToken(r)
		fmt.Printf("SAVED => %+v", saved)
		if err != nil {
			app.invalidCredentials(w)
			return
		}

		next.ServeHTTP(w,r);
	}) 
}