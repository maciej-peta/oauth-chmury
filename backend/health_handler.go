package main

import (
	"log"
	"net/http"
)

func healthHandler(writer http.ResponseWriter, _ *http.Request) {

	if err := db.Ping(); err != nil {
		http.Error(writer, "DB unreachable", http.StatusServiceUnavailable)
		log.Println("Healthcheck: backend - lost connection to db")
		return
	}
	log.Println("Healthcheck: backend - ok")
	writer.WriteHeader(http.StatusOK)
}
