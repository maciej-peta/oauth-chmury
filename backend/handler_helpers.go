package main

import (
	"net/http"
	"slices"
	"strings"
)

var acceptableOrigins = []string{"http://localhost:3000", "http://frontend:3000", "http://frontend.local"}

func setHandlerHeaders(w http.ResponseWriter, r *http.Request, methods ...string) {
	origin := r.Header.Get("Origin")
	if slices.Contains(acceptableOrigins, origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ", "))
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}
