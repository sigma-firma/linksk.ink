package main

import "net/http"

// registerRoutes registers the routes with the provided *http.ServeMux
func registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", root)
}
