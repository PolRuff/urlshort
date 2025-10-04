package main

import (
	"fmt"
	"net/http"

	"github.com/PolRuff/urlshort/internal/config"
	"github.com/PolRuff/urlshort/internal/handler"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			handler.ShortenHandler(w, r)
		} else {
			handler.RedirectHandler(w, r)
		}
	})

	fmt.Println("Server is running on http://localhost" + config.ServerPort)
	http.ListenAndServe(config.ServerPort, nil)
}
